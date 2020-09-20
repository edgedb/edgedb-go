package protocol

import (
	"encoding/binary"
	"fmt"
)

func PopMessage(bts *[]byte) []byte {
	n := 1 + Uint32((*bts)[1:5])
	msg := make([]byte, n)
	copy(msg, *bts)
	*bts = (*bts)[n:]
	return msg
}

func PopUint8(bts *[]byte) uint8 {
	val := (*bts)[0]
	*bts = (*bts)[1:]
	return val
}

func PushUint8(bts *[]byte, val uint8) {
	*bts = append(*bts, val)
}

// Uint16 decode uint16
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

func Uint32(bts []byte) uint32 {
	return binary.BigEndian.Uint32(bts[:4])
}

// Uint32 decode uint32
func PopUint32(bts *[]byte) uint32 {
	val := binary.BigEndian.Uint32(*bts)
	*bts = (*bts)[4:]
	return val
}

func PutUint32(bts []byte, val uint32) {
	binary.BigEndian.PutUint32(bts, val)
}

// Uint64 decode uint64
func PopUint64(bts *[]byte) uint64 {
	val := binary.BigEndian.Uint64(*bts)
	*bts = (*bts)[8:]
	return val
}

func PutUint64(bts []byte, val uint64) {
	binary.BigEndian.PutUint64(bts, val)
}

// Int32 decode int32
func PopInt32(bts *[]byte) int32 {
	return int32(PopUint32(bts))
}

func PutInt32(bts []byte, val int32) {
	PutUint32(bts, uint32(val))
}

// Int64 decode int64
func PopInt64(bts *[]byte) int64 {
	return int64(PopUint64(bts))
}

func PutInt64(bts []byte, val int64) {
	PutUint64(bts, uint64(val))
}

// Bytes decode []byte
func PopBytes(bts *[]byte) ([]byte, int) {
	n := PopUint32(bts)
	out := make([]byte, n)
	copy(out, (*bts)[:n])
	*bts = (*bts)[n:]
	return out, int(4 + n)
}

func PutBytes(bts []byte, val []byte) {
	PutUint32(bts[:4], uint32(len(val)))
	copy(bts[4:], val)
}

// String decode a string
func PopString(bts *[]byte) (string, int) {
	out, n := PopBytes(bts)
	return string(out), n
}

func PutString(bts []byte, val string) {
	PutBytes(bts, []byte(val))
}

type UUID string

func PopUUID(bts *[]byte) UUID {
	b := *bts
	val := fmt.Sprintf("%x-%x-%x-%x-%x-%x",
		b[:4], b[4:6], b[6:8], b[8:10], b[10:12], b[12:16])
	*bts = b[16:]
	return UUID(val)
}
