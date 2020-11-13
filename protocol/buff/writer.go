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

	"github.com/edgedb/edgedb-go/types"
)

// Writer is a write only buffer.
type Writer struct {
	bts  []byte
	mPos int
	bPos []int
}

// NewWriter returns a new Writer.
func NewWriter(bts []byte) *Writer {
	return &Writer{bts: bts}
}

// Reset empties the buffer.
func (b *Writer) Reset() {
	b.bts = b.bts[:0]
	b.bPos = b.bPos[:0]
	b.mPos = 0
}

// PushUint8 writes a uint8 to the buffer.
func (b *Writer) PushUint8(val uint8) {
	b.bts = append(b.bts, val)
}

// PushUint16 writes a uint16 to the buffer.
func (b *Writer) PushUint16(val uint16) {
	n := len(b.bts)
	b.bts = append(b.bts, 0, 0)
	binary.BigEndian.PutUint16(b.bts[n:], val)
}

// PushUint32 writes a uint32 to the buffer.
func (b *Writer) PushUint32(val uint32) {
	n := len(b.bts)
	b.bts = append(b.bts, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(b.bts[n:], val)
}

// PushUint64 writes a uint64 to the buffer.
func (b *Writer) PushUint64(val uint64) {
	tmp := make([]byte, 8)
	binary.BigEndian.PutUint64(tmp, val)
	b.bts = append(b.bts, tmp...)
}

// PushUUID writes a types.UUID to the buffer.
func (b *Writer) PushUUID(val types.UUID) {
	b.bts = append(b.bts, val[:]...)
}

// PushBytes writes a []byte to the buffer.
func (b *Writer) PushBytes(val []byte) {
	b.PushUint32(uint32(len(val)))
	b.bts = append(b.bts, val...)
}

// PushString writes a string to the buffer.
func (b *Writer) PushString(val string) {
	b.PushUint32(uint32(len(val)))
	b.bts = append(b.bts, val...)
}

// BeginBytes allocates space for `data_length` in the buffer.
// May be called multiple times to create nested bytes blocks.
// Calling EndBytes once for each BeginBytes call is required
// before ending a message.
// BeginBytes panics if BeginMessage was not called first.
func (b *Writer) BeginBytes() {
	if b.mPos <= 0 {
		panic("cannot begin bytes: no current message")
	}

	b.bPos = append(b.bPos, len(b.bts))
	b.bts = append(b.bts, 0, 0, 0, 0)
}

// EndBytes sets the `data_length` allocated by BeginBytes
// to the number of bytes that were written since the last BeginBytes call.
// EndBytes panics if BeginBytes was not called first.
func (b *Writer) EndBytes() {
	n := len(b.bPos)
	if n < 1 {
		panic("cannot end bytes: no bytes in progress")
	}

	p := b.bPos[n-1]
	b.bPos = b.bPos[:n-1]
	length := uint32(len(b.bts) - p - 4)
	binary.BigEndian.PutUint32(b.bts[p:p+4], length)
}

// BeginMessage writes mType to the buffer
// and allocates space for message length.
// BeginMessage panics if the EndMessage was not called
// for the previous message.
func (b *Writer) BeginMessage(mType uint8) {
	if b.mPos > 0 {
		panic("cannot begin message: the previous message is not finished")
	}

	b.mPos = 1 + len(b.bts)
	b.bts = append(b.bts, mType, 0, 0, 0, 0)
}

// EndMessage sets the `message_length` allocated by BeginMessage.
// EndMessage panics if BeginMessage was not called first
// or if BeginBytes was not followed by EndBytes.
func (b *Writer) EndMessage() {
	// todo check unfinished bytes
	if b.mPos <= 0 {
		panic("cannot end message: no current message")
	}

	if len(b.bPos) > 0 {
		panic("cannot end message: bytes in progress")
	}

	length := uint32(len(b.bts) - b.mPos)
	binary.BigEndian.PutUint32(b.bts[b.mPos:b.mPos+4], length)
	b.mPos = 0
}

// Unwrap returns a pointer to the buffers *[]byte.
func (b *Writer) Unwrap() *[]byte {
	if b.mPos > 0 {
		panic("cannot unwrap: the previous message is not finished")
	}

	return &b.bts
}

// Next returns true if the buffer is not fully read.
func (b *Writer) Next() bool {
	return b.mPos < len(b.bts)
}

// PopMessage returns a Message with the next message
// excluding the message length and advances the buffer.
func (b *Writer) PopMessage() *Message {
	if len(b.bts) < b.mPos+5 {
		panic("buffer overread")
	}

	pos := 1 + int(binary.BigEndian.Uint32(b.bts[b.mPos+1:b.mPos+5]))

	if len(b.bts) < b.mPos+pos {
		panic("buffer overread")
	}

	msg := &Message{
		bts:  b.bts[b.mPos+5 : b.mPos+pos],
		Type: b.bts[b.mPos],
	}
	b.mPos += pos
	return msg
}
