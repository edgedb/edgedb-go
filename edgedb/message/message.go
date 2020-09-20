package message

import (
	"fmt"

	"github.com/fmoor/edgedb-golang/edgedb/protocol"
)

const (
	// PrepareType https://edgedb.com/docs/internals/protocol/messages#prepare
	PrepareType         = 0x50
	CmdCmpltType        = 0x43
	CmdDataDescType     = 0x54
	DescStmtType        = 0x44
	PrepareCmpltType    = 0x31
	RdyForCmdType       = 0x5a
	ServerKeyDataType   = 0x4b
	ServerHandshakeType = 0x76
	SyncType            = 0x53
	ErrorResponseType   = 0x45
	ExecuteType         = 0x45
	DataType            = 0x44

	// IO Formats
	BinaryFormat       = 0x62
	JSONFormat         = 0x6a
	JSONElementsFormat = 0x4a

	// Cardinalities
	NoResult = 0x6e
	One      = 0x6f
	Many     = 0x6d

	// Aspects
	DataDescription = 0x54
)

// Messages
var SyncMsg Message = FromBytes([]byte{SyncType, 0, 0, 0, 0})

// Params ...
type Params map[string]string

// Message for the edgedb server
type Message struct {
	buf []byte
}

// Make creates a new message
func Make(mType uint8) Message {
	return Message{[]byte{mType, 0, 0, 0, 0}}
}

func FromBytes(bts []byte) Message {
	return Message{bts}
}

// PushUint8 appends a byte to the message
func (m *Message) PushUint8(value uint8) {
	m.buf = append(m.buf, value)
}

// PushUint16 appends a uint16 value to the message
func (m *Message) PushUint16(value uint16) {
	protocol.PushUint16(&m.buf, value)
}

// PushUint32 appends a uint32 value to the message
func (m *Message) PushUint32(value uint32) {
	tmp := make([]byte, 4)
	protocol.PutUint32(tmp, value)
	m.buf = append(m.buf, tmp...)
}

// PushString appends a 32bit-length prefixed string to the message
func (m *Message) PushString(str string) {
	m.PushUint32(uint32(len(str)))
	m.buf = append(m.buf, str...)
}

// PushBytes appends 32bit-length prefixed bytes to the message
func (m *Message) PushBytes(bts []byte) {
	m.PushUint32(uint32(len(bts)))
	m.buf = append(m.buf, bts...)
}

// PushParams appends parameters to the message
func (m *Message) PushParams(params map[string]string) {
	m.PushUint16(uint16(len(params)))
	for key, val := range params {
		m.PushString(key)
		m.PushString(val)
	}
}

// ToBytes returns the full message as bytes
func (m *Message) ToBytes() []byte {
	out := make([]byte, len(m.buf))
	copy(out, m.buf)
	length := uint32(len(out[1:]))
	protocol.PutUint32(out[1:5], length)
	fmt.Printf("wrote message %q:\n% x\n", out[0], out)
	return out
}
