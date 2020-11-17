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

package buff

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/types"
)

// Buff is a write only buffer.
type Buff struct {
	payload []byte
	mPos    int
	bPos    []int
	Msg     []byte
	msgHdr  *reflect.SliceHeader
	MsgType uint8
}

// New returns a new *Buff.
func New(payload []byte) *Buff {
	b := &Buff{payload: payload}
	b.Msg = b.payload[:0]
	b.msgHdr = (*reflect.SliceHeader)(unsafe.Pointer(&b.Msg))
	b.msgHdr.Cap = 0
	return b
}

// NewReader returns a new *Buff with payload in it's Msg atribute.
// This is useful for parsing cached descriptors.
func NewReader(payload []byte) *Buff {
	b := &Buff{payload: payload}
	b.Msg = b.payload
	return b
}

// Reset empties the buffer.
func (b *Buff) Reset() {
	b.payload = b.payload[:0]
	b.bPos = b.bPos[:0]
	b.mPos = 0
	b.msgHdr.Cap = 0
	b.msgHdr.Len = 0
}

// PushUint8 writes a uint8 to the buffer.
func (b *Buff) PushUint8(val uint8) {
	b.payload = append(b.payload, val)
}

// PushUint16 writes a uint16 to the buffer.
func (b *Buff) PushUint16(val uint16) {
	n := len(b.payload)
	b.payload = append(b.payload, 0, 0)
	binary.BigEndian.PutUint16(b.payload[n:], val)
}

// PushUint32 writes a uint32 to the buffer.
func (b *Buff) PushUint32(val uint32) {
	n := len(b.payload)
	b.payload = append(b.payload, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(b.payload[n:], val)
}

// PushUint64 writes a uint64 to the buffer.
func (b *Buff) PushUint64(val uint64) {
	tmp := make([]byte, 8)
	binary.BigEndian.PutUint64(tmp, val)
	b.payload = append(b.payload, tmp...)
}

// PushUUID writes a types.UUID to the buffer.
func (b *Buff) PushUUID(val types.UUID) {
	b.payload = append(b.payload, val[:]...)
}

// PushBytes writes a []byte to the buffer.
func (b *Buff) PushBytes(val []byte) {
	b.PushUint32(uint32(len(val)))
	b.payload = append(b.payload, val...)
}

// PushString writes a string to the buffer.
func (b *Buff) PushString(val string) {
	b.PushUint32(uint32(len(val)))
	b.payload = append(b.payload, val...)
}

// BeginBytes allocates space for `data_length` in the buffer.
// May be called multiple times to create nested bytes blocks.
// Calling EndBytes once for each BeginBytes call is required
// before ending a message.
// BeginBytes panics if BeginMessage was not called first.
func (b *Buff) BeginBytes() {
	if b.mPos <= 0 {
		panic("cannot begin bytes: no current message")
	}

	b.bPos = append(b.bPos, len(b.payload))
	b.payload = append(b.payload, 0, 0, 0, 0)
}

// EndBytes sets the `data_length` allocated by BeginBytes
// to the number of bytes that were written since the last BeginBytes call.
// EndBytes panics if BeginBytes was not called first.
func (b *Buff) EndBytes() {
	n := len(b.bPos)
	if n < 1 {
		panic("cannot end bytes: no bytes in progress")
	}

	p := b.bPos[n-1]
	b.bPos = b.bPos[:n-1]
	length := uint32(len(b.payload) - p - 4)
	binary.BigEndian.PutUint32(b.payload[p:p+4], length)
}

// BeginMessage writes mType to the buffer
// and allocates space for message length.
// BeginMessage panics if the EndMessage was not called
// for the previous message.
func (b *Buff) BeginMessage(mType uint8) {
	if b.mPos > 0 {
		panic("cannot begin message: the previous message is not finished")
	}

	b.mPos = 1 + len(b.payload)
	b.payload = append(b.payload, mType, 0, 0, 0, 0)
}

// EndMessage sets the `message_length` allocated by BeginMessage.
// EndMessage panics if BeginMessage was not called first
// or if BeginBytes was not followed by EndBytes.
func (b *Buff) EndMessage() {
	// todo check unfinished bytes
	if b.mPos <= 0 {
		panic("cannot end message: no current message")
	}

	if len(b.bPos) > 0 {
		panic("cannot end message: bytes in progress")
	}

	length := uint32(len(b.payload) - b.mPos)
	binary.BigEndian.PutUint32(b.payload[b.mPos:b.mPos+4], length)
	b.mPos = 0
}

// Unwrap returns a pointer to the buffers *[]byte.
func (b *Buff) Unwrap() *[]byte {
	if b.mPos > 0 {
		panic("cannot unwrap: the previous message is not finished")
	}

	return &b.payload
}

// Next returns true if the buffer is not fully read.
func (b *Buff) Next() bool {
	b.Finish()

	if b.mPos < len(b.payload) {
		if len(b.payload) < b.mPos+5 {
			panic("buffer overread")
		}

		pos := 1 + int(binary.BigEndian.Uint32(b.payload[b.mPos+1:b.mPos+5]))

		if len(b.payload) < b.mPos+pos {
			panic("buffer overread")
		}

		b.Msg = b.payload[b.mPos+5 : b.mPos+pos]
		b.msgHdr.Cap = b.msgHdr.Len
		b.MsgType = b.payload[b.mPos]
		b.mPos += pos

		return true
	}

	b.msgHdr.Cap = 0
	b.msgHdr.Len = 0
	b.MsgType = 0
	return false
}

// Finish asserts that the message has been fully read.
// It panics if it has not.
func (b *Buff) Finish() {
	if len(b.Msg) > 0 {
		panic(fmt.Sprintf(
			"cannot finish: unread data in buffer (message type: 0x%x)",
			b.MsgType,
		))
	}
}

// Len returns the number of bytes remaining to be read.
func (b *Buff) Len() int {
	return len(b.Msg)
}

// AssertAllocated panics if there aren't n bytes in the buffer.
func (b *Buff) AssertAllocated(n int) {
	if len(b.Msg) < n {
		panic("buffer overread")
	}
}

// Discard skips the next n bytes.
func (b *Buff) Discard(n int) {
	b.AssertAllocated(n)
	b.Msg = b.Msg[n:]
}

// PopUint8 returns the next byte and advances the buffer.
func (b *Buff) PopUint8() uint8 {
	val := b.Msg[0]
	b.Msg = b.Msg[1:]
	return val
}

// PopUint16 reads a uint16 and advances the buffer.
func (b *Buff) PopUint16() uint16 {
	val := binary.BigEndian.Uint16(b.Msg)
	b.Msg = b.Msg[2:]
	return val
}

// PopUint32 reads a uint32 and advances the buffer.
func (b *Buff) PopUint32() uint32 {
	val := b.PeekUint32()
	b.Msg = b.Msg[4:]
	return val
}

// PeekUint32 reads a uint32 but does not advance the buffer.
func (b *Buff) PeekUint32() uint32 {
	return binary.BigEndian.Uint32(b.Msg)
}

// PopUint64 reads a uint64 and advances the buffer.
func (b *Buff) PopUint64() uint64 {
	val := binary.BigEndian.Uint64(b.Msg[:8])
	b.Msg = b.Msg[8:]
	return val
}

// PopUUID reads a types.UUID and advances the buffer.
func (b *Buff) PopUUID() types.UUID {
	var id types.UUID
	copy(id[:], b.Msg[:16])
	b.Msg = b.Msg[16:]
	return id
}

// PopBytes reads a []byte and advances the buffer.
// The returned slice is owned by the buffer.
func (b *Buff) PopBytes() []byte {
	n := int(b.PopUint32())
	b.AssertAllocated(n)
	val := b.Msg[:n]
	b.Msg = b.Msg[n:]
	return val
}

// PopString reads a string and advances the buffer.
func (b *Buff) PopString() string {
	return string(b.PopBytes())
}
