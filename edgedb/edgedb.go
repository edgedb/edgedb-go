package edgedb

// todo add context.Context

import (
	"fmt"
	"io"
	"net"

	"github.com/fmoor/edgedb-golang/edgedb/marshal"
	"github.com/fmoor/edgedb-golang/edgedb/protocol"
	"github.com/fmoor/edgedb-golang/edgedb/protocol/aspect"
	"github.com/fmoor/edgedb-golang/edgedb/protocol/cardinality"
	"github.com/fmoor/edgedb-golang/edgedb/protocol/codecs"
	"github.com/fmoor/edgedb-golang/edgedb/protocol/format"
	"github.com/fmoor/edgedb-golang/edgedb/protocol/message"
	"github.com/fmoor/edgedb-golang/edgedb/types"
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
func (conn *Conn) Close() error {
	// todo adjust return value if conn.conn.Close() close returns an error
	defer conn.conn.Close()
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

// Transaction creates a new trasaction struct.
func (conn *Conn) Transaction() error {
	// https://www.edgedb.com/docs/clients/00_python/api/blocking_con#edgedb.BlockingIOConnection.transaction
	// todo implement Transaction()
	panic("not implemented")
}

// Execute an EdgeQL command (or commands).
func (conn *Conn) Execute(query string) error {
	// https://www.edgedb.com/docs/clients/00_python/api/blocking_con#edgedb.BlockingIOConnection.execute
	// todo implement Execute()
	panic("not implemented")
}

// QueryOne runs a singleton-returning query and return its element.
func (conn *Conn) QueryOne(query string, out interface{}, args map[string]interface{}) error {
	// https://www.edgedb.com/docs/clients/00_python/api/blocking_con#edgedb.BlockingIOConnection.query_one
	// todo implement QueryOne()
	panic("not implemented")
}

// Query runs a query and returns the results.
func (conn *Conn) Query(query string, out interface{}) error {
	return conn.QueryWithArgs(query, out, map[string]interface{}{})
}

func (conn *Conn) QueryWithArgs(query string, out interface{}, args map[string]interface{}) error {
	// todo assert that out is a pointer to a slice
	result, err := conn.query(query, args)
	if err != nil {
		return err
	} else if len(result) == 0 {
		return nil
	}

	marshal.Marshal(&out, result)
	return nil
}

// QueryJSON runs a query and return the results as JSON.
func (conn *Conn) QueryJSON(query string, out *string, args map[string]interface{}) error {
	// https://www.edgedb.com/docs/clients/00_python/api/blocking_con#edgedb.BlockingIOConnection.query_json
	// todo implement QueryJSON()
	panic("not implemented")
}

// QueryOneJSON runs a singleton-returning query and return its element in JSON
func (conn *Conn) QueryOneJSON(query string, out *string, args map[string]interface{}) error {
	// https://www.edgedb.com/docs/clients/00_python/api/blocking_con#edgedb.BlockingIOConnection.query_one_json
	// todo implement QueryOneJSON()
	panic("not implemented")
}

func (conn *Conn) query(query string, args map[string]interface{}) ([]interface{}, error) {
	msg := []byte{message.Prepare, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushUint8(&msg, format.Binary)
	protocol.PushUint8(&msg, cardinality.Many)
	protocol.PushBytes(&msg, []byte{}) // no statement name
	protocol.PushString(&msg, query)
	protocol.PutMsgLength(msg)

	pyld := msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)
	var argumentCodecID types.UUID
	var resultCodecID types.UUID

	rcv := conn.writeAndRead(pyld)
	for len(rcv) > 4 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.PrepareComplete:
			protocol.PopUint32(&bts)                 // message length
			protocol.PopUint16(&bts)                 // number of headers, assume 0
			protocol.PopUint8(&bts)                  // cardianlity
			argumentCodecID = protocol.PopUUID(&bts) // argument type id
			resultCodecID = protocol.PopUUID(&bts)   // result type id
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
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushUint8(&msg, aspect.DataDescription)
	protocol.PushUint32(&msg, 0) // no statement name
	protocol.PutMsgLength(msg)

	pyld = msg
	pyld = append(pyld, message.Sync, 0, 0, 0, 4)

	rcv = conn.writeAndRead(pyld)

	codecLookup := codecs.CodecLookup{}

	for len(rcv) > 4 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.CommandDataDescription:
			protocol.PopUint32(&bts)              // message length
			protocol.PopUint16(&bts)              // number of headers is always 0
			protocol.PopUint8(&bts)               // cardianlity
			protocol.PopUUID(&bts)                // argument descriptor ID
			descriptor := protocol.PopBytes(&bts) // argument descriptor
			for k, v := range codecs.Pop(&descriptor) {
				codecLookup[k] = v
			}
			protocol.PopUUID(&bts)               // result descriptor ID
			descriptor = protocol.PopBytes(&bts) // argument descriptor

			for k, v := range codecs.Pop(&descriptor) {
				codecLookup[k] = v
			}
		case message.ErrorResponse:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint8(&bts)  // severity
			protocol.PopUint32(&bts) // code
			message := protocol.PopString(&bts)
			panic(message)
		}
	}

	argumentCodec := codecLookup[argumentCodecID]
	resultCodec := codecLookup[resultCodecID]

	msg = []byte{message.Execute, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0)       // no headers
	protocol.PushBytes(&msg, []byte{}) // no statement name
	argumentCodec.Encode(&msg, args)
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
			out = append(out, resultCodec.Decode(&bts))
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

// Connect establishes a connection to an EdgeDB server.
func Connect(config ConnConfig) (conn *Conn, err error) {
	tcpConn, err := net.Dial("tcp", "127.0.0.1:5656")
	if err != nil {
		return conn, fmt.Errorf("tcp connection error while connecting: %v", err)
	}
	conn = &Conn{tcpConn, nil}

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

	rcv := conn.writeAndRead(msg)

	var secret []byte

	for len(rcv) > 0 {
		bts := protocol.PopMessage(&rcv)
		mType := protocol.PopUint8(&bts)

		switch mType {
		case message.ServerHandshake:
			// todo close the connection if protocol version can't be supported
			// https://edgedb.com/docs/internals/protocol/overview#connection-phase
		case message.ServerKeyData:
			secret = bts[5:]
		case message.ReadyForCommand:
			return &Conn{tcpConn, secret}, nil
		case message.ErrorResponse:
			protocol.PopUint32(&bts) // message length
			protocol.PopUint8(&bts)  // severity
			protocol.PopUint32(&bts) // code
			message := protocol.PopString(&bts)
			panic(message)
		}
	}
	return conn, nil
}
