package edgedb

// todo add context.Context

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/fmoor/edgedb-golang/edgedb/marshal"
	"github.com/fmoor/edgedb-golang/edgedb/protocol"
	"github.com/fmoor/edgedb-golang/edgedb/protocol/aspect"
	"github.com/fmoor/edgedb-golang/edgedb/protocol/cardinality"
	"github.com/fmoor/edgedb-golang/edgedb/protocol/codecs"
	"github.com/fmoor/edgedb-golang/edgedb/protocol/format"
	"github.com/fmoor/edgedb-golang/edgedb/protocol/message"
	"github.com/fmoor/edgedb-golang/edgedb/types"
)

// todo add examples

var (
	// ErrorZeroResults is returned when a query has no results.
	// todo use this in all the query methods
	ErrorZeroResults = errors.New("zero results")
)

// Options for connecting to an EdgeDB server
type Options struct {
	Host     string
	Port     int
	User     string
	Database string
	Password string
	// todo support authentication etc.
}

func (o Options) dialHost() string {
	host := o.Host
	if host == "" {
		host = "localhost"
	}

	port := o.Port
	if port == 0 {
		port = 5656
	}

	return fmt.Sprintf("%v:%v", host, port)
}

// DSN parses a URI string into an Options struct
func DSN(dsn string) Options {
	// todo assert scheme is edgedb
	parsed, err := url.Parse(dsn)
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		panic(err)
	}

	host := strings.Split(parsed.Host, ":")[0]
	db := strings.TrimLeft(parsed.Path, "/")
	password, _ := parsed.User.Password()

	return Options{
		Host:     host,
		Port:     port,
		User:     parsed.User.Username(),
		Database: db,
		Password: password,
	}
}

// Error is returned when something bad happened.
type Error struct {
	Severity int
	Code     int
	Message  string
}

func (e *Error) Error() string {
	return e.Message
}

func decodeError(bts *[]byte) error {
	protocol.PopUint32(bts) // message length

	return &Error{
		Severity: int(protocol.PopUint8(bts)),
		Code:     int(protocol.PopUint32(bts)),
		Message:  protocol.PopString(bts),
	}
}

type queryCodecIDs struct {
	inputID  types.UUID
	outputID types.UUID
}

// Conn client
type Conn struct {
	conn       io.ReadWriteCloser
	secret     []byte
	codecCache codecs.CodecLookup
	queryCache map[string]queryCodecIDs
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

// Transaction creates a new trasaction struct.
func (conn *Conn) Transaction() (Transaction, error) {
	// todo support transaction options
	return Transaction{conn}, nil
}

// RunInTransaction runs a function in a transaction.
// If function returns an error transaction is rolled back,
// otherwise transaction is committed.
func (conn *Conn) RunInTransaction(fn func() error) error {
	// see https://pkg.go.dev/github.com/go-pg/pg/v10#DB.RunInTransaction
	panic("RunInTransaction() not implemented") // todo
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
			break
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
func (conn *Conn) QueryOne(query string, out interface{}, args ...interface{}) error {
	// https://www.edgedb.com/docs/clients/00_python/api/blocking_con#edgedb.BlockingIOConnection.query_one
	// todo assert cardinality
	result, err := conn.query(query, args...)
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
func (conn *Conn) Query(query string, out interface{}, args ...interface{}) error {
	// todo assert that out is a pointer to a slice
	result, err := conn.query(query, args...)
	if err != nil {
		return err
	}

	marshal.Marshal(&out, result)
	return nil
}

// QueryJSON runs a query and return the results as JSON.
func (conn *Conn) QueryJSON(query string, args ...interface{}) ([]byte, error) {
	result, err := conn.query(query, args...)
	if err != nil {
		return []byte{}, err
	}

	// todo set format instead of marshaling json
	return json.Marshal(result)
}

// QueryOneJSON runs a singleton-returning query and return its element in JSON.
// If the query executes successfully but doesn't return a result
// []byte{}, ErrorZeroResults is returned.
func (conn *Conn) QueryOneJSON(query string, args ...interface{}) ([]byte, error) {
	// todo assert cardinally
	result, err := conn.query(query, args...)
	if err != nil {
		return []byte{}, err
	}

	if len(result) == 0 {
		return []byte{}, ErrorZeroResults
	}

	// todo set format instead of marshaling json
	return json.Marshal(result[0])
}

func (conn *Conn) query(query string, args ...interface{}) ([]interface{}, error) {
	_, hasCodecs := conn.queryCache[query]
	if !hasCodecs {
		return conn.execute(query, args...)
	}

	return conn.optimisticExecute(query, args...)
}

func (conn *Conn) optimisticExecute(query string, args ...interface{}) ([]interface{}, error) {
	codecIDs := conn.queryCache[query]
	inputCodec := conn.codecCache[codecIDs.inputID]
	outputCodec := conn.codecCache[codecIDs.outputID]

	msg := []byte{message.OptimisticExecute, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushUint8(&msg, format.Binary)
	protocol.PushUint8(&msg, cardinality.Many) // todo should this be more intelligent?
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

func (conn *Conn) execute(query string, args ...interface{}) ([]interface{}, error) {
	msg := []byte{message.Prepare, 0, 0, 0, 0}
	protocol.PushUint16(&msg, 0) // no headers
	protocol.PushUint8(&msg, format.Binary)
	protocol.PushUint8(&msg, cardinality.Many) // todo should this be more intelligent?
	protocol.PushBytes(&msg, []byte{})         // no statement name
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
			protocol.PopUint32(&bts)               // message length
			protocol.PopUint16(&bts)               // number of headers, assume 0
			protocol.PopUint8(&bts)                // cardianlity
			inputCodecID = protocol.PopUUID(&bts)  // input type id
			outputCodecID = protocol.PopUUID(&bts) // output type id
		case message.ReadyForCommand:
			break
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
	conn.queryCache[query] = queryCodecIDs{inputCodecID, outputCodecID}

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

// Connect establishes a connection to an EdgeDB server.
func Connect(opts Options) (conn *Conn, err error) {
	tcpConn, err := net.Dial("tcp", opts.dialHost())
	if err != nil {
		return nil, fmt.Errorf("tcp connection error while connecting: %v", err)
	}
	conn = &Conn{tcpConn, nil, codecs.CodecLookup{}, map[string]queryCodecIDs{}}

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
			return &Conn{tcpConn, secret, codecs.CodecLookup{}, map[string]queryCodecIDs{}}, nil
		case message.ErrorResponse:
			return nil, decodeError(&bts)
		case message.AuthenticationOK:
			continue
		default:
			panic(fmt.Sprintf("unexpected message type: 0x%x", mType))
		}
	}
	return conn, nil
}

// Transaction represents a transaction or save point block.
// Transactions are created by calling the Conn.Transaction() method.
// Most callers should use `Conn.RunInTransaction()` instead.
type Transaction struct {
	conn *Conn
}

// Start a transaction or save point.
func (tx Transaction) Start() error {
	// todo handle nested blocks and other options.
	return tx.conn.Execute("START TRANSACTION;")
}

// Commit the transaction or save point preserving changes.
func (tx Transaction) Commit() error {
	// todo handle nested blocks etc.
	return tx.conn.Execute("COMMIT;")
}

// RollBack the transaction or save point block discarding changes.
func (tx Transaction) RollBack() error {
	// todo handle nested blocks etc.
	return tx.conn.Execute("ROLLBACK;")
}
