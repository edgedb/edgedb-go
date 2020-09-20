package main

import (
	"fmt"
	"io"
	"net"

	"github.com/fmoor/edgedb-golang/edgedb/codecs"
	"github.com/fmoor/edgedb-golang/edgedb/message"
	"github.com/fmoor/edgedb-golang/edgedb/payload"
	"github.com/fmoor/edgedb-golang/edgedb/protocol"
)

// EdgeDB client
type EdgeDB struct {
	conn   io.ReadWriteCloser
	secret []byte
}

// Close the db connection
func (edb *EdgeDB) Close() error {
	// todo adjust return value if close returns an error
	defer edb.conn.Close()
	buf := []byte{0x58, 0, 0, 0, 0}
	if _, err := edb.conn.Write(buf); err != nil {
		return fmt.Errorf("error while terminating: %v", err)
	}
	return nil
}

func (edb *EdgeDB) writeAndRead(bts []byte) []byte {
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
func (edb *EdgeDB) Query(query string) (interface{}, error) {
	msg := message.Make(message.PrepareType)
	msg.PushUint16(0) // no headers
	msg.PushUint8(message.BinaryFormat)
	msg.PushUint8(message.Many)
	msg.PushBytes([]byte{}) // no statement name
	msg.PushString(query)

	pyld := payload.Make(msg.ToBytes())
	pyld.Push(message.SyncMsg.ToBytes())

	rcv := edb.writeAndRead(pyld.ToBytes())
	var resultDescriptorID protocol.UUID

	for len(rcv) > 4 {
		bts := protocol.PopMessage(&rcv)

		switch protocol.PopUint8(&bts) {
		case message.PrepareCmpltType:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint16(&bts) // number of headers, assume 0
			protocol.PopUint8(&bts)  // cardianlity
			protocol.PopUUID(&bts)   // argument type id
			resultDescriptorID = protocol.PopUUID(&bts)
		case message.RdyForCmdType:
			break
		case message.ErrorResponseType:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint8(&bts)  // severity
			protocol.PopUint32(&bts) // code
			message, _ := protocol.PopString(&bts)
			panic(message)
		}
	}

	msg = message.Make(message.DescStmtType)
	msg.PushUint16(0)   // no headers
	msg.PushUint8(0x54) // aspect DataDescription
	msg.PushUint32(0)   // no statement name

	pyld = payload.Make(msg.ToBytes())
	pyld.Push(message.SyncMsg.ToBytes())

	rcv = edb.writeAndRead(pyld.ToBytes())

	decoderLookup := codecs.DecoderLookup{}

	for len(rcv) > 4 {
		bts := protocol.PopMessage(&rcv)

		switch protocol.PopUint8(&bts) {
		case message.CmdDataDescType:
			protocol.PopUint32(&bts)                 // message length
			protocol.PopUint16(&bts)                 // number of headers is always 0
			protocol.PopUint8(&bts)                  // cardianlity
			protocol.PopUUID(&bts)                   // argument descriptor ID
			protocol.PopBytes(&bts)                  // argument descriptor
			protocol.PopUUID(&bts)                   // output descriptor ID
			descriptor, _ := protocol.PopBytes(&bts) // argument descriptor

			for k, v := range codecs.Get(&descriptor) {
				decoderLookup[k] = v
			}
		case message.ErrorResponseType:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint8(&bts)  // severity
			protocol.PopUint32(&bts) // code
			message, _ := protocol.PopString(&bts)
			panic(message)
		}
	}

	msg = message.Make(message.ExecuteType)
	msg.PushUint16(0)                 // no headers
	msg.PushBytes([]byte{})           // no statement name
	msg.PushBytes([]byte{0, 0, 0, 0}) // no argument data

	pyld = payload.Make(msg.ToBytes())
	pyld.Push(message.SyncMsg.ToBytes())

	rcv = edb.writeAndRead(pyld.ToBytes())

	decoder := decoderLookup[resultDescriptorID]
	out := "Set{ "

	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)

		switch protocol.PopUint8(&bts) {
		case message.DataType:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint16(&bts) // number of data elements (always 1)
			out += " " + decoder.Decode(&bts)
		case message.CmdCmpltType:
			continue
		case message.RdyForCmdType:
			continue
		case message.ErrorResponseType:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint8(&bts)  // severity
			protocol.PopUint32(&bts) // code
			message, _ := protocol.PopString(&bts)
			panic(message)
		}
	}
	out += " }"
	return out, nil
}

// Connect to a database
func Connect() (edb EdgeDB, err error) {
	conn, err := net.Dial("tcp", "127.0.0.1:5656")
	if err != nil {
		return edb, fmt.Errorf("tcp connection error while connecting: %v", err)
	}
	edb = EdgeDB{conn, nil}

	msg := message.Make(0x56)
	msg.PushUint16(0) // major version
	msg.PushUint16(8) // minor version
	msg.PushParams(message.Params{
		"user":     "edgedb",
		"database": "edgedb",
	})
	msg.PushUint16(0) // no extensions

	rcv := edb.writeAndRead(msg.ToBytes())

	var secret []byte

	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)

		switch protocol.PopUint8(&bts) {
		case message.ServerHandshakeType:
			// todo close the connection if protocol version can't be supported
			// https://edgedb.com/docs/internals/protocol/overview#connection-phase
		case message.ServerKeyDataType:
			secret = bts[5:]
		case message.RdyForCmdType:
			return EdgeDB{conn, secret}, nil
		case message.ErrorResponseType:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint8(&bts)  // severity
			protocol.PopUint32(&bts) // code
			message, _ := protocol.PopString(&bts)
			panic(message)
		}
	}
	return edb, nil
}

func main() {
	edb, err := Connect()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer edb.Close()
	result, _ := edb.Query(`SELECT sys::Database{name, id, builtin};`)
	fmt.Println(result)
}
