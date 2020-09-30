package protocol

import (
	"encoding/binary"

	"github.com/fmoor/edgedb-golang/edgedb/types"
)

func PopUint8(bts *[]byte) uint8 {
	val := (*bts)[0]
	*bts = (*bts)[1:]
	return val
}

func PushUint8(bts *[]byte, val uint8) {
	*bts = append(*bts, val)
}

func PopUint16(bts *[]byte) uint16 {
	val := binary.BigEndian.Uint16((*bts)[:2])
	*bts = (*bts)[2:]
	return val
}

func PushUint16(bts *[]byte, val uint16) {
	segment := make([]byte, 2)
	binary.BigEndian.PutUint16(segment, val)
	*bts = append(*bts, segment...)
}

func PopUint32(bts *[]byte) uint32 {
	val := binary.BigEndian.Uint32(*bts)
	*bts = (*bts)[4:]
	return val
}

func PeekUint32(bts *[]byte) uint32 {
	return binary.BigEndian.Uint32(*bts)
}

func PushUint32(bts *[]byte, val uint32) {
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, val)
	*bts = append(*bts, tmp...)
}

func PopUint64(bts *[]byte) uint64 {
	val := binary.BigEndian.Uint64(*bts)
	*bts = (*bts)[8:]
	return val
}

func PushUint64(bts *[]byte, val uint64) {
	tmp := make([]byte, 8)
	binary.BigEndian.PutUint64(tmp, val)
	*bts = append(*bts, tmp...)
}

func PopBytes(bts *[]byte) []byte {
	n := PopUint32(bts)
	out := make([]byte, n)
	copy(out, (*bts)[:n])
	*bts = (*bts)[n:]
	return out
}

func PushBytes(bts *[]byte, val []byte) {
	PushUint32(bts, uint32(len(val)))
	*bts = append(*bts, val...)
}

func PopString(bts *[]byte) string {
	return string(PopBytes(bts))
}

func PushString(bts *[]byte, val string) {
	PushUint32(bts, uint32(len(val)))
	*bts = append(*bts, val...)
}

func PopUUID(bts *[]byte) types.UUID {
	var id types.UUID
	copy(id[:], (*bts)[:16])
	*bts = (*bts)[16:]
	return id
}

func PopMessage(bts *[]byte) []byte {
	n := 1 + binary.BigEndian.Uint32((*bts)[1:5])
	msg := make([]byte, n)
	copy(msg, *bts)
	*bts = (*bts)[n:]
	return msg
}

// PutMsgLength sets the message length bytes
// only call this after the message is complete
func PutMsgLength(msg []byte) {
	// bytes [1:5] are the length of the message excluding the initial message type byte
	// https://www.edgedb.com/docs/internals/protocol/messages
	binary.BigEndian.PutUint32(msg[1:5], uint32(len(msg[1:])))
}
