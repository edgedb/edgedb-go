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

import (
	"context"
	"fmt"
	"net"

	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/protocol/aspect"
	"github.com/edgedb/edgedb-go/edgedb/protocol/cardinality"
	"github.com/edgedb/edgedb-go/edgedb/protocol/codecs"
	"github.com/edgedb/edgedb-go/edgedb/protocol/format"
	"github.com/edgedb/edgedb-go/edgedb/protocol/message"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

func (c *Client) queryCodecs(query string, ioFmt uint8) (queryCodecs, bool) {
	// todo this isn't thread safe
	key := queryCacheKey{query, ioFmt}
	codecs, ok := c.queryCache[key]
	return codecs, ok
}

func (c *Client) cacheQuery(
	query string,
	ioFmt uint8,
	in,
	out codecs.DecodeEncoder,
) {
	// todo this isn't thread safe
	key := queryCacheKey{query: query, format: ioFmt}
	c.queryCache[key] = queryCodecs{in: in, out: out}
}

func (c *Client) granularFlow(
	ctx context.Context,
	conn net.Conn,
	query string,
	ioFmt uint8,
	args []interface{},
) ([]interface{}, error) {
	codecs, ok := c.queryCodecs(query, ioFmt)
	if ok {
		return c.optimistic(ctx, conn, codecs, query, ioFmt, args)
	}

	return c.pesimistic(ctx, conn, query, ioFmt, args)
}

func (c *Client) pesimistic(
	ctx context.Context,
	conn net.Conn,
	query string,
	ioFmt uint8,
	args []interface{},
) ([]interface{}, error) {
	inID, outID, err := prepare(ctx, conn, query, ioFmt)
	if err != nil {
		return nil, err
	}

	in, inOK := c.codecCache[inID]
	out, outOK := c.codecCache[outID]
	if !inOK || !outOK {
		err := c.describe(ctx, conn)
		if err != nil {
			return nil, err
		}

		in = c.codecCache[inID]
		out = c.codecCache[outID]
		c.cacheQuery(query, ioFmt, in, out)
	}

	if ioFmt == format.JSON {
		// treat json format as bytes instead of string
		out = c.codecCache[types.UUID{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2,
		}]
	}

	return execute(ctx, conn, in, out, args)
}

func prepare(
	ctx context.Context,
	conn net.Conn,
	query string,
	ioFmt uint8,
) (in types.UUID, out types.UUID, err error) {
	buf := []byte{message.Prepare, 0, 0, 0, 0}
	protocol.PushUint16(&buf, 0) // no headers
	protocol.PushUint8(&buf, ioFmt)
	protocol.PushUint8(&buf, cardinality.Many) // todo is this correct?
	protocol.PushBytes(&buf, []byte{})         // no statement name
	protocol.PushString(&buf, query)
	protocol.PutMsgLength(buf)

	buf = append(buf, message.Sync, 0, 0, 0, 4)

	err = writeAndRead(ctx, conn, &buf)
	if err != nil {
		return in, out, err
	}

	for len(buf) > 4 {
		msg := protocol.PopMessage(&buf)
		mType := protocol.PopUint8(&msg)

		switch mType {
		case message.PrepareComplete:
			protocol.PopUint32(&msg)     // message length
			protocol.PopUint16(&msg)     // number of headers, assume 0
			protocol.PopUint8(&msg)      // cardianlity
			in = protocol.PopUUID(&msg)  // input type id
			out = protocol.PopUUID(&msg) // output type id
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return in, out, decodeError(&msg)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return in, out, nil
}

func (c *Client) describe(ctx context.Context, conn net.Conn) error {
	buf := []byte{message.DescribeStatement, 0, 0, 0, 0}
	protocol.PushUint16(&buf, 0) // no headers
	protocol.PushUint8(&buf, aspect.DataDescription)
	protocol.PushUint32(&buf, 0) // no statement name
	protocol.PutMsgLength(buf)

	buf = append(buf, message.Sync, 0, 0, 0, 4)

	err := writeAndRead(ctx, conn, &buf)
	if err != nil {
		return err
	}

	for len(buf) > 4 {
		msg := protocol.PopMessage(&buf)
		mType := protocol.PopUint8(&msg)

		switch mType {
		case message.CommandDataDescription:
			protocol.PopUint32(&msg)              // message length
			protocol.PopUint16(&msg)              // num headers is always 0
			protocol.PopUint8(&msg)               // cardianlity
			protocol.PopUUID(&msg)                // input descriptor ID
			descriptor := protocol.PopBytes(&msg) // input descriptor
			codecs.UpdateCache(c.codecCache, &descriptor)

			protocol.PopUUID(&msg)               // output descriptor ID
			descriptor = protocol.PopBytes(&msg) // input descriptor
			codecs.UpdateCache(c.codecCache, &descriptor)
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return decodeError(&msg)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return nil
}

func execute(
	ctx context.Context,
	conn net.Conn,
	in codecs.DecodeEncoder,
	out codecs.DecodeEncoder,
	args []interface{},
) ([]interface{}, error) {
	buf := []byte{message.Execute, 0, 0, 0, 0}
	protocol.PushUint16(&buf, 0)       // no headers
	protocol.PushBytes(&buf, []byte{}) // no statement name
	in.Encode(&buf, args)
	protocol.PutMsgLength(buf)

	buf = append(buf, message.Sync, 0, 0, 0, 4)

	err := writeAndRead(ctx, conn, &buf)
	if err != nil {
		return nil, err
	}

	result := make(types.Set, 0)
	for len(buf) > 0 {
		msg := protocol.PopMessage(&buf)
		mType := protocol.PopUint8(&msg)

		switch mType {
		case message.Data:
			protocol.PopUint32(&msg) // message length
			protocol.PopUint16(&msg) // number of data elements (always 1)
			result = append(result, out.Decode(&msg))
		case message.CommandComplete:
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return nil, decodeError(&msg)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return result, nil
}

var buffer [8192]byte

func (c *Client) optimistic(
	ctx context.Context,
	conn net.Conn,
	codecs queryCodecs,
	query string,
	ioFmt uint8,
	args []interface{},
) ([]interface{}, error) {
	inID := codecs.in.ID()
	outID := codecs.out.ID()

	out := codecs.out
	if ioFmt == format.JSON {
		// treat json format as bytes instead of string
		out = c.codecCache[types.UUID{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2,
		}]
	}

	buf := buffer[:0]
	buf = append(buf,
		message.OptimisticExecute,
		0, 0, 0, 0, // message length slot, to be filled in later
		0, 0, // no headers
		ioFmt,
		cardinality.Many, // todo is this correct?
	)

	protocol.PushString(&buf, query)
	buf = append(buf, inID[:]...)
	buf = append(buf, outID[:]...)
	codecs.in.Encode(&buf, args)
	protocol.PutMsgLength(buf)

	buf = append(buf, message.Sync, 0, 0, 0, 4)

	err := writeAndRead(ctx, conn, &buf)
	if err != nil {
		return nil, err
	}

	result := make(types.Set, 0)
	for len(buf) > 0 {
		msg := protocol.PopMessage(&buf)
		mType := protocol.PopUint8(&msg)

		switch mType {
		case message.Data:
			// skip the following fields
			// message length
			// number of data elements (always 1)
			msg = msg[6:]
			result = append(result, out.Decode(&msg))
		case message.CommandComplete:
		case message.ReadyForCommand:
		case message.ErrorResponse:
			return nil, decodeError(&msg)
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}

	return result, nil
}
