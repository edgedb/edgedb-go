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

package protocol

import (
	"encoding/binary"

	"github.com/edgedb/edgedb-go/edgedb/types"
)

// PopUint8 removes a uint8 from the buffer.
func PopUint8(bts *[]byte) uint8 {
	val := (*bts)[0]
	*bts = (*bts)[1:]
	return val
}

// PushUint8 adds a uint8 to the buffer.
func PushUint8(bts *[]byte, val uint8) {
	*bts = append(*bts, val)
}

// PopUint16 removes a uint16 from the buffer.
func PopUint16(bts *[]byte) uint16 {
	val := binary.BigEndian.Uint16((*bts)[:2])
	*bts = (*bts)[2:]
	return val
}

// PushUint16 adds a uint16 to the buffer
func PushUint16(bts *[]byte, val uint16) {
	n := len(*bts)
	*bts = append(*bts, 0, 0)
	slot := (*bts)[n:]
	binary.BigEndian.PutUint16(slot, val)
}

// PopUint32 removes a uint32 from the buffer.
func PopUint32(bts *[]byte) uint32 {
	val := binary.BigEndian.Uint32(*bts)
	*bts = (*bts)[4:]
	return val
}

// PeekUint32 reads a uint32 from the buffer without removing it.
func PeekUint32(bts *[]byte) uint32 {
	return binary.BigEndian.Uint32(*bts)
}

// PushUint32 adds a uint32 to the buffer.
func PushUint32(bts *[]byte, val uint32) {
	n := len(*bts)
	*bts = append(*bts, 0, 0, 0, 0)
	slot := (*bts)[n:]
	binary.BigEndian.PutUint32(slot, val)
}

// PopUint64 removes a uint64 from the buffer.
func PopUint64(bts *[]byte) uint64 {
	val := binary.BigEndian.Uint64(*bts)
	*bts = (*bts)[8:]
	return val
}

// PushUint64 adds a uint64 to the buffer.
func PushUint64(bts *[]byte, val uint64) {
	tmp := make([]byte, 8)
	binary.BigEndian.PutUint64(tmp, val)
	*bts = append(*bts, tmp...)
}

// PopBytes removes a bytes string from the buffer.
func PopBytes(bts *[]byte) []byte {
	n := PopUint32(bts)
	out := (*bts)[:n]
	*bts = (*bts)[n:]
	return out
}

// PushBytes adds a byte string to the buffer.
func PushBytes(bts *[]byte, val []byte) {
	PushUint32(bts, uint32(len(val)))
	*bts = append(*bts, val...)
}

// PopString removes a string from the buffer.
func PopString(bts *[]byte) string {
	return string(PopBytes(bts))
}

// PushString adds a string to the buffer.
func PushString(bts *[]byte, val string) {
	PushUint32(bts, uint32(len(val)))
	*bts = append(*bts, val...)
}

// PopUUID removes a UUID from the buffer.
func PopUUID(bts *[]byte) types.UUID {
	var id types.UUID
	copy(id[:], (*bts)[:16])
	*bts = (*bts)[16:]
	return id
}

// PopMessage removes a message from the buffer.
func PopMessage(bts *[]byte) []byte {
	n := 1 + binary.BigEndian.Uint32((*bts)[1:5])
	msg := (*bts)[:n]
	*bts = (*bts)[n:]
	return msg
}

// PutMsgLength sets the message length bytes
// only call this after the message is complete
func PutMsgLength(msg []byte) {
	// bytes [1:5] are the length of the message
	// excluding the initial message type byte
	// https://www.edgedb.com/docs/internals/protocol/messages
	binary.BigEndian.PutUint32(msg[1:5], uint32(len(msg[1:])))
}
