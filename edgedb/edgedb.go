package main

// todo add context

import (
	"fmt"
	"io"
	"net"

	"github.com/fmoor/edgedb-golang/edgedb/cardinality"
	"github.com/fmoor/edgedb-golang/edgedb/codecs"
	"github.com/fmoor/edgedb-golang/edgedb/format"
	"github.com/fmoor/edgedb-golang/edgedb/message"
	"github.com/fmoor/edgedb-golang/edgedb/protocol"
)

// Conn client
type Conn struct {
	conn   io.ReadWriteCloser
	secret []byte
}

// ConnConfig options for configuring a connection
type ConnConfig struct {
	Database string
	User     string
	// todo support authentication etc.
}

// Close the db connection
func (edb *Conn) Close() error {
	// todo adjust return value if close returns an error
	defer edb.conn.Close()
	buf := []byte{message.Terminate, 0, 0, 0, 4}
	if _, err := edb.conn.Write(buf); err != nil {
		return fmt.Errorf("error while terminating: %v", err)
	}
	return nil
}

func (edb *Conn) writeAndRead(bts []byte) []byte {
	buf := bts
	for len(buf) > 0 {
		msg := protocol.PopMessage(&buf)
		fmt.Printf("writing message %q:\n% x\n", msg[0], msg)
	}

	if _, err := edb.conn.Write(bts); err != nil {
		panic(err)
	}

	rcv := []byte{}
	n := 1024
	var err error
	for n == 1024 {
		tmp := make([]byte, 1024)
		n, err = edb.conn.Read(tmp)
		if err != nil {
			panic(err)
		}
		rcv = append(rcv, tmp[:n]...)
	}

	buf = rcv
	for len(buf) > 0 {
		msg := protocol.PopMessage(&buf)
		fmt.Printf("read message %q:\n% x\n", msg[0], msg)
	}

	return rcv
}

// Query the database
func (edb *Conn) Query(query string) (interface{}, error) {
	msg := []byte{message.Prepare, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushUint8(&msg, format.Binary)
	protocol.PushUint8(&msg, cardinality.Many)
	protocol.PushBytes(&msg, []byte{}) // no statement name
	protocol.PushString(&msg, query)
	protocol.PutMsgLength(msg)

	pyld := msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)

	rcv := edb.writeAndRead(pyld)
	var resultDescriptorID protocol.UUID

	for len(rcv) > 4 {
		bts := protocol.PopMessage(&rcv)

		switch protocol.PopUint8(&bts) {
		case message.PrepareComplete:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint16(&bts) // number of headers, assume 0
			protocol.PopUint8(&bts)  // cardianlity
			protocol.PopUUID(&bts)   // argument type id
			resultDescriptorID = protocol.PopUUID(&bts)
		case message.ReadyForCommand:
			break
		case message.ErrorResponse:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint8(&bts)  // severity
			protocol.PopUint32(&bts) // code
			message := protocol.PopString(&bts)
			panic(message)
		}
	}

	msg = []byte{message.DescribeStatement, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0)   // no headers
	protocol.PushUint8(&msg, 0x54) // aspect DataDescription
	protocol.PushUint32(&msg, 0)   // no statement name
	protocol.PutMsgLength(msg)

	pyld = msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)

	rcv = edb.writeAndRead(pyld)

	decoderLookup := codecs.CodecLookup{}

	for len(rcv) > 4 {
		bts := protocol.PopMessage(&rcv)

		switch protocol.PopUint8(&bts) {
		case message.CommandDataDescription:
			protocol.PopUint32(&bts)              // message length
			protocol.PopUint16(&bts)              // number of headers is always 0
			protocol.PopUint8(&bts)               // cardianlity
			protocol.PopUUID(&bts)                // argument descriptor ID
			protocol.PopBytes(&bts)               // argument descriptor
			protocol.PopUUID(&bts)                // output descriptor ID
			descriptor := protocol.PopBytes(&bts) // argument descriptor

			for k, v := range codecs.Get(&descriptor) {
				decoderLookup[k] = v
			}
		case message.ErrorResponse:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint8(&bts)  // severity
			protocol.PopUint32(&bts) // code
			message := protocol.PopString(&bts)
			panic(message)
		}
	}

	msg = []byte{message.Execute, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0)                 // no headers
	protocol.PushBytes(&msg, []byte{})           // no statement name
	protocol.PushBytes(&msg, []byte{0, 0, 0, 0}) // no argument data
	protocol.PutMsgLength(msg)

	pyld = msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)

	rcv = edb.writeAndRead(pyld)

	decoder := decoderLookup[resultDescriptorID]
	out := []interface{}{}

	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)

		switch protocol.PopUint8(&bts) {
		case message.Data:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint16(&bts) // number of data elements (always 1)
			out = append(out, decoder.Decode(&bts))
		case message.CommandComplete:
			continue
		case message.ReadyForCommand:
			continue
		case message.ErrorResponse:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint8(&bts)  // severity
			protocol.PopUint32(&bts) // code
			message := protocol.PopString(&bts)
			panic(message)
		}
	}
	return out, nil
}

// Connect to a database
func Connect(config ConnConfig) (edb *Conn, err error) {
	conn, err := net.Dial("tcp", "127.0.0.1:5656")
	if err != nil {
		return edb, fmt.Errorf("tcp connection error while connecting: %v", err)
	}
	edb = &Conn{conn, nil}

	msg := []byte{message.ClientHandshake, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // major version
	protocol.PushUint16(&msg, 8) // minor version
	protocol.PushUint16(&msg, 2) // number of parameters
	protocol.PushString(&msg, "database")
	protocol.PushString(&msg, config.Database)
	protocol.PushString(&msg, "user")
	protocol.PushString(&msg, config.User)
	protocol.PushUint16(&msg, 0) // no extensions
	protocol.PutMsgLength(msg)

	rcv := edb.writeAndRead(msg)

	var secret []byte

	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)

		switch protocol.PopUint8(&bts) {
		case message.ServerHandshake:
			// todo close the connection if protocol version can't be supported
			// https://edgedb.com/docs/internals/protocol/overview#connection-phase
		case message.ServerKeyData:
			secret = bts[5:]
		case message.ReadyForCommand:
			return &Conn{conn, secret}, nil
		case message.ErrorResponse:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint8(&bts)  // severity
			protocol.PopUint32(&bts) // code
			message := protocol.PopString(&bts)
			panic(message)
		}
	}
	return edb, nil
}

func main() {
	options := ConnConfig{"edgedb", "edgedb"}
	edb, err := Connect(options)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer edb.Close()
	result, _ := edb.Query(`SELECT sys::Database{name, id, builtin};`)
	fmt.Println(result)
}
