// This source file is part of the EdgeDB open source project.
//
// Copyright 2020-present EdgeDB Inc. and the EdgeDB authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package edgedb

// todo add context.Context

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/xdg/scram"

	"github.com/edgedb/edgedb-go/edgedb/marshal"
	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/protocol/aspect"
	"github.com/edgedb/edgedb-go/edgedb/protocol/cardinality"
	"github.com/edgedb/edgedb-go/edgedb/protocol/codecs"
	"github.com/edgedb/edgedb-go/edgedb/protocol/format"
	"github.com/edgedb/edgedb-go/edgedb/protocol/message"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

// todo add examples

var (
	// ErrorZeroResults is returned when a query has no results.
	// todo should this be returned from Query() and QueryJSON()? :thinking:
	ErrorZeroResults = errors.New("zero results")
)

type queryCodecIDs struct {
	inputID  types.UUID
	outputID types.UUID
}

type queryCacheKey struct {
	query  string
	format int
}

// Conn client
type Conn struct {
	conn       io.ReadWriteCloser
	secret     []byte
	codecCache codecs.CodecLookup
	queryCache map[queryCacheKey]queryCodecIDs
}

// Close the db connection
func (conn *Conn) Close() (err error) {
	defer func() {
		err = conn.conn.Close()
	}()
	msg := []byte{message.Terminate, 0, 0, 0, 4}
	if _, err := conn.conn.Write(msg); err != nil {
		return fmt.Errorf("error while terminating: %v", err)
	}
	return nil
}

func (conn *Conn) writeAndRead(bts []byte) []byte {
	// tmp := bts
	// for len(tmp) > 0 {
	// 	msg := protocol.PopMessage(&tmp)
	// 	fmt.Printf("writing message %q:\n% x\n", msg[0], msg)
	// }

	if _, err := conn.conn.Write(bts); err != nil {
		panic(err)
	}

	rcv := []byte{}
	// todo evaluate buffer size
	n := 1024
	var err error
	for n == 1024 {
		tmp := make([]byte, 1024)
		n, err = conn.conn.Read(tmp)
		if err != nil {
			panic(err)
		}
		rcv = append(rcv, tmp[:n]...)
	}

	// tmp = rcv
	// for len(tmp) > 0 {
	// 	msg := protocol.PopMessage(&tmp)
	// 	fmt.Printf("read message %q:\n% x\n", msg[0], msg)
	// }

	return rcv
}

// Execute an EdgeQL command (or commands).
func (conn *Conn) Execute(query string) error {
	msg := []byte{message.ExecuteScript, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushString(&msg, query)
	protocol.PutMsgLength(msg)

	rcv := conn.writeAndRead(msg)
	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.CommandComplete:
			continue
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return decodeError(&bts)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return nil
}

// QueryOne runs a singleton-returning query and return its element.
// If the query executes successfully but doesn't return a result
// ErrorZeroResults is returned.
func (conn *Conn) QueryOne(
	query string,
	out interface{},
	args ...interface{},
) error {
	// todo assert cardinality
	result, err := conn.query(query, format.Binary, args...)
	if err != nil {
		return err
	}

	if len(result) == 0 {
		return ErrorZeroResults
	}
	marshal.Marshal(&out, result[0])
	return nil
}

// Query runs a query and returns the results.
func (conn *Conn) Query(
	query string,
	out interface{},
	args ...interface{},
) error {
	// todo assert that out is a pointer to a slice
	result, err := conn.query(query, format.Binary, args...)
	if err != nil {
		return err
	}

	marshal.Marshal(&out, result)
	return nil
}

// QueryJSON runs a query and return the results as JSON.
func (conn *Conn) QueryJSON(
	query string,
	args ...interface{},
) ([]byte, error) {
	result, err := conn.query(query, format.JSON, args...)
	if err != nil {
		return nil, err
	}

	return []byte(result[0].(string)), nil
}

// QueryOneJSON runs a singleton-returning query
// and return its element in JSON.
// If the query executes successfully but doesn't return a result
// []byte{}, ErrorZeroResults is returned.
func (conn *Conn) QueryOneJSON(
	query string,
	args ...interface{},
) ([]byte, error) {
	// todo assert cardinally
	result, err := conn.query(query, format.JSON, args...)
	if err != nil {
		return nil, err
	}

	jsonStr := result[0].(string)
	if len(jsonStr) == 2 { // "[]"
		return nil, ErrorZeroResults
	}

	return []byte(jsonStr[1 : len(jsonStr)-1]), nil
}

func (conn *Conn) query(
	query string,
	ioFormat int,
	args ...interface{},
) ([]interface{}, error) {
	key := queryCacheKey{query, ioFormat}
	_, hasCodecs := conn.queryCache[key]

	if !hasCodecs {
		return conn.execute(query, ioFormat, args...)
	}

	return conn.optimisticExecute(query, ioFormat, args...)
}

func (conn *Conn) optimisticExecute(
	query string,
	ioFormat int,
	args ...interface{},
) ([]interface{}, error) {
	key := queryCacheKey{query, ioFormat}
	codecIDs := conn.queryCache[key]
	inputCodec := conn.codecCache[codecIDs.inputID]
	outputCodec := conn.codecCache[codecIDs.outputID]

	msg := []byte{message.OptimisticExecute, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushUint8(&msg, uint8(ioFormat))

	// todo should this be more intelligent?
	protocol.PushUint8(&msg, cardinality.Many)

	protocol.PushString(&msg, query)
	msg = append(msg, codecIDs.inputID[:]...)
	msg = append(msg, codecIDs.outputID[:]...)
	inputCodec.Encode(&msg, args)
	protocol.PutMsgLength(msg)

	pyld := msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)

	out := make(types.Set, 0)

	rcv := conn.writeAndRead(pyld)
	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.Data:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint16(&bts) // number of data elements (always 1)
			out = append(out, outputCodec.Decode(&bts))
		case message.CommandComplete:
			continue
		case message.ReadyForCommand:
			continue
		case message.ErrorResponse:
			return nil, decodeError(&bts)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return out, nil
}

func (conn *Conn) execute(
	query string,
	ioFormat int,
	args ...interface{},
) ([]interface{}, error) {
	msg := []byte{message.Prepare, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushUint8(&msg, uint8(ioFormat))

	// todo should this be more intelligent?
	protocol.PushUint8(&msg, cardinality.Many)

	protocol.PushBytes(&msg, []byte{}) // no statement name
	protocol.PushString(&msg, query)
	protocol.PutMsgLength(msg)

	pyld := msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)
	var inputCodecID types.UUID
	var outputCodecID types.UUID

	rcv := conn.writeAndRead(pyld)
	for len(rcv) > 4 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.PrepareComplete:
			protocol.PopUint32(&bts) // message length

			// number of headers, assume 0
			protocol.PopUint16(&bts)

			protocol.PopUint8(&bts)                // cardianlity
			inputCodecID = protocol.PopUUID(&bts)  // input type id
			outputCodecID = protocol.PopUUID(&bts) // output type id
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return nil, decodeError(&bts)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	inputCodec, haveArg := conn.codecCache[inputCodecID]
	outputCodec, haveRes := conn.codecCache[outputCodecID]

	if !haveArg || !haveRes {
		err := conn.cacheMissingCodecs()
		if err != nil {
			return nil, err
		}
		inputCodec = conn.codecCache[inputCodecID]
		outputCodec = conn.codecCache[outputCodecID]
	}

	key := queryCacheKey{query, ioFormat}
	conn.queryCache[key] = queryCodecIDs{inputCodecID, outputCodecID}

	msg = []byte{message.Execute, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0)       // no headers
	protocol.PushBytes(&msg, []byte{}) // no statement name
	inputCodec.Encode(&msg, args)
	protocol.PutMsgLength(msg)

	pyld = msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)

	rcv = conn.writeAndRead(pyld)
	out := make(types.Set, 0)

	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.Data:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint16(&bts) // number of data elements (always 1)
			out = append(out, outputCodec.Decode(&bts))
		case message.CommandComplete:
			continue
		case message.ReadyForCommand:
			continue
		case message.ErrorResponse:
			return nil, decodeError(&bts)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return out, nil
}

func (conn *Conn) cacheMissingCodecs() error {
	msg := []byte{message.DescribeStatement, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushUint8(&msg, aspect.DataDescription)
	protocol.PushUint32(&msg, 0) // no statement name
	protocol.PutMsgLength(msg)

	pyld := msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)

	rcv := conn.writeAndRead(pyld)
	for len(rcv) > 4 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.CommandDataDescription:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint16(&bts) // number of headers is always 0
			protocol.PopUint8(&bts)  // cardianlity

			protocol.PopUUID(&bts)                // input descriptor ID
			descriptor := protocol.PopBytes(&bts) // input descriptor
			for k, v := range codecs.Pop(&descriptor) {
				conn.codecCache[k] = v
			}

			protocol.PopUUID(&bts)               // output descriptor ID
			descriptor = protocol.PopBytes(&bts) // input descriptor
			for k, v := range codecs.Pop(&descriptor) {
				conn.codecCache[k] = v
			}
		case message.ReadyForCommand:
			continue
		case message.ErrorResponse:
			return decodeError(&bts)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return nil
}

func (conn *Conn) authenticate(username string, password string) error {
	client, err := scram.SHA256.NewClient(username, password, "")
	if err != nil {
		panic(err)
	}

	conv := client.NewConversation()
	scramMsg, err := conv.Step("")
	if err != nil {
		panic(err)
	}

	msg := []byte{message.AuthenticationSASLInitialResponse, 0, 0, 0, 0}
	protocol.PushString(&msg, "SCRAM-SHA-256")
	protocol.PushString(&msg, scramMsg)
	protocol.PutMsgLength(msg)

	rcv := conn.writeAndRead(msg)
	mType := protocol.PopUint8(&rcv)

	switch mType {
	case message.Authentication:
		protocol.PopUint32(&rcv) // message length
		authStatus := protocol.PopUint32(&rcv)
		if authStatus != 0xb {
			panic(fmt.Sprintf(
				"unexpected authentication status: 0x%x",
				authStatus,
			))
		}

		scramRcv := protocol.PopString(&rcv)
		scramMsg, err = conv.Step(scramRcv)
		if err != nil {
			panic(err)
		}
	case message.ErrorResponse:
		return decodeError(&rcv)
	default:
		panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
	}

	msg = []byte{message.AuthenticationSASLResponse, 0, 0, 0, 0}
	protocol.PushString(&msg, scramMsg)
	protocol.PutMsgLength(msg)

	rcv = conn.writeAndRead(msg)
	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.Authentication:
			protocol.PopUint32(&bts) // message length
			authStatus := protocol.PopUint32(&bts)

			switch authStatus {
			case 0:
				continue
			case 0xc:
				scramRcv := protocol.PopString(&bts)
				_, err = conv.Step(scramRcv)
				if err != nil {
					panic(err)
				}
			default:
				panic(fmt.Sprintf(
					"unexpected authentication status: 0x%x",
					authStatus,
				))
			}
		case message.ServerKeyData:
			conn.secret = bts[5:]
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return decodeError(&bts)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return nil
}

// Connect establishes a connection to an EdgeDB server.
func Connect(opts Options) (conn *Conn, err error) {
	// todo add connection timeout
	tcpConn, err := net.Dial(opts.socType(), opts.dialHost())

	if err != nil {
		err = fmt.Errorf("tcp connection error while connecting: %v", err)
		return nil, err
	}
	conn = &Conn{
		tcpConn,
		nil,
		codecs.CodecLookup{},
		map[queryCacheKey]queryCodecIDs{},
	}

	msg := []byte{message.ClientHandshake, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // major version
	protocol.PushUint16(&msg, 8) // minor version
	protocol.PushUint16(&msg, 2) // number of parameters
	protocol.PushString(&msg, "database")
	protocol.PushString(&msg, opts.Database)
	protocol.PushString(&msg, "user")
	protocol.PushString(&msg, opts.User)
	protocol.PushUint16(&msg, 0) // no extensions
	protocol.PutMsgLength(msg)

	rcv := conn.writeAndRead(msg)
	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.ServerHandshake:
			// The client _MUST_ close the connection
			// if the protocol version can't be supported.
			// https://edgedb.com/docs/internals/protocol/overview
			protocol.PopUint32(&bts) // message length
			major := protocol.PopUint16(&bts)
			minor := protocol.PopUint16(&bts)

			if major != 0 || minor != 8 {
				err := conn.conn.Close()
				if err != nil {
					return nil, err
				}
				err = fmt.Errorf(
					"unsupported protocol version: %v.%v",
					major,
					minor,
				)
				return nil, err
			}
		case message.ServerKeyData:
			conn.secret = bts[5:]
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return nil, decodeError(&bts)
		case message.Authentication:
			protocol.PopUint32(&bts) // message length
			authStatus := protocol.PopUint32(&bts)

			if authStatus != 0 {
				err := conn.authenticate(opts.User, opts.Password)
				if err != nil {
					return nil, err
				}
			}
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}
	return conn, nil
}
