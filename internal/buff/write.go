// This source file is part of the EdgeDB open source project.
//
// Copyright EdgeDB Inc. and the EdgeDB authors.
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

	types "github.com/geldata/gel-go/internal/geltypes"
)

// Writer is a write buffer.
type Writer struct {
	buf     []byte
	msgPos  int
	bytePos []int
}

// NewWriter returns a new Writer.
func NewWriter(alocatedMemory []byte) *Writer {
	return &Writer{buf: alocatedMemory[:0]}
}

// Unwrap returns the underlying []byte.
func (w *Writer) Unwrap() []byte {
	if w.msgPos != 0 {
		panic("cannot send: the previous message is not finished")
	}

	if len(w.buf) == 0 {
		panic("cannot send: no data")
	}

	buf := w.buf
	w.buf = nil
	return buf
}

// PushUint8 writes a uint8 to the buffer.
func (w *Writer) PushUint8(val uint8) {
	w.buf = append(w.buf, val)
}

// PushUint16 writes a uint16 to the buffer.
func (w *Writer) PushUint16(val uint16) {
	n := len(w.buf)
	w.buf = append(w.buf, 0, 0)
	binary.BigEndian.PutUint16(w.buf[n:n+2], val)
}

// PushUint32 writes a uint32 to the buffer.
func (w *Writer) PushUint32(val uint32) {
	n := len(w.buf)
	w.buf = append(w.buf, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(w.buf[n:n+4], val)
}

// PushUint64 writes a uint64 to the buffer.
func (w *Writer) PushUint64(val uint64) {
	n := len(w.buf)
	w.buf = append(w.buf, 0, 0, 0, 0, 0, 0, 0, 0)
	binary.BigEndian.PutUint64(w.buf[n:n+8], val)
}

// PushUUID writes a types.UUID to the buffer.
func (w *Writer) PushUUID(val types.UUID) {
	w.buf = append(w.buf, val[:]...)
}

// PushBytes writes []byte to the buffer.
func (w *Writer) PushBytes(val []byte) {
	w.buf = append(w.buf, val...)
}

// PushString writes a string to the buffer.
func (w *Writer) PushString(val string) {
	w.PushUint32(uint32(len(val)))
	w.PushBytes([]byte(val))
}

// BeginBytes allocates space for `data_length` in the buffer.
// May be called multiple times to create nested bytes blocks.
// Calling EndBytes once for each BeginBytes call is required
// before ending a message.
// BeginBytes panics if BeginMessage was not called first.
func (w *Writer) BeginBytes() {
	if w.msgPos == 0 {
		panic("cannot begin bytes: no current message")
	}

	n := len(w.buf)
	w.buf = append(w.buf, 0, 0, 0, 0)
	w.bytePos = append(w.bytePos, n)
}

// EndBytes sets the `data_length` allocated by BeginBytes
// to the number of bytes that were written since the last BeginBytes call.
// EndBytes panics if BeginBytes was not called first.
func (w *Writer) EndBytes() {
	n := len(w.bytePos)
	if n < 1 {
		panic("cannot end bytes: no bytes in progress")
	}

	pos := w.bytePos[n-1]
	w.bytePos = w.bytePos[:n-1]

	byteLen := uint32(len(w.buf) - pos - 4)
	binary.BigEndian.PutUint32(w.buf[pos:], byteLen)
}

// BeginMessage writes mType to the buffer
// and allocates space for message length.
// BeginMessage panics if the EndMessage was not called
// for the previous message.
func (w *Writer) BeginMessage(mType uint8) {
	if w.msgPos != 0 {
		panic("cannot begin message: the previous message is not finished")
	}

	w.msgPos = 1 + len(w.buf)
	w.buf = append(w.buf, mType, 0, 0, 0, 0)
}

// EndMessage sets the `message_length` allocated by BeginMessage.
// EndMessage panics if BeginMessage was not called first
// or if BeginBytes was not followed by EndBytes.
func (w *Writer) EndMessage() {
	if w.msgPos == 0 {
		panic("cannot end message: no current message")
	}

	if len(w.bytePos) != 0 {
		panic("cannot end message: bytes in progress")
	}

	msgLen := uint32(len(w.buf) - w.msgPos)
	binary.BigEndian.PutUint32(w.buf[w.msgPos:], msgLen)
	w.msgPos = 0
}
