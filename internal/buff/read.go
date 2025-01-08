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
	"fmt"

	types "github.com/edgedb/edgedb-go/internal/geltypes"
	"github.com/edgedb/edgedb-go/internal/soc"
)

// Reader is a buffer reader.
type Reader struct {
	toBeDeserialized chan *soc.Data

	data    *soc.Data
	Err     error
	Buf     []byte
	MsgType uint8
}

// NewReader returns a new Reader.
func NewReader(toBeDeserialized chan *soc.Data) *Reader {
	return &Reader{toBeDeserialized: toBeDeserialized}
}

// SimpleReader creates a new reader that operates on a single []byte.
func SimpleReader(buf []byte) *Reader {
	r := &Reader{Buf: buf[:len(buf):len(buf)]}
	return r
}

// Next advances the reader to the next message.
// Next returns false when the reader doesn't own any socket data
// and a signal is received on doneReadingSignal,
// or an error is encountered while reading.
//
// Callers must continue to call Next until it returns false.
//
// Next() panics if called on a reader created with SimpleReader().
func (r *Reader) Next(doneReadingSignal chan struct{}) bool {
	if r.toBeDeserialized == nil {
		panic("called next on a simple reader")
	}

	if len(r.Buf) > 0 {
		r.Err = fmt.Errorf(
			"cannot finish: unread data in buffer (message type: 0x%x)",
			r.MsgType,
		)
		return false
	}

	if r.data != nil && len(r.data.Buf) == 0 {
		r.data.Release()
		r.data = nil
	}

	r.MsgType = 0

	if r.data == nil {
		select {
		case <-doneReadingSignal:
			return false
		case r.data = <-r.toBeDeserialized:
			if r.data.Err != nil {
				r.Err = r.data.Err
				r.data.Release()
				r.data = nil
				return false
			}
		}
	}

	// put message type and length into r.Buf
	r.Err = r.feed(5)
	if r.Err != nil {
		return false
	}

	r.MsgType = r.PopUint8()
	msgLen := int(r.PopUint32()) - 4

	r.Err = r.feed(msgLen)
	if r.Err != nil {
		return false
	}

	r.Buf = r.Buf[:msgLen:msgLen]
	return true
}

func min(x, y int) int {
	if x < y {
		return x
	}

	return y
}

func (r *Reader) feed(n int) error {
	if r.data != nil && len(r.data.Buf) == 0 {
		r.data.Release()
		r.data = nil
	}

	if n == 0 {
		return nil
	}

	if r.data == nil {
		r.data = <-r.toBeDeserialized

		if r.data.Err != nil {
			e := r.data.Err
			r.data.Release()
			r.data = nil
			return e
		}
	}

	m := min(n, len(r.data.Buf))
	r.Buf = r.data.Buf[:m]
	r.data.Buf = r.data.Buf[m:]

	for len(r.Buf) < n {
		previous := r.data
		r.data = <-r.toBeDeserialized

		if r.data.Err != nil {
			previous.Release()
			e := r.data.Err
			r.data.Release()
			r.data = nil
			return e
		}

		m := min(n-len(r.Buf), len(r.data.Buf))
		r.Buf = append(r.Buf, r.data.Buf[:m]...)
		r.data.Buf = r.data.Buf[m:]
		previous.Release()
	}

	return nil
}

// Discard skips n bytes.
func (r *Reader) Discard(n int) {
	r.Buf = r.Buf[n:]
}

// DiscardMessage discards all remaining bytes in the current message.
func (r *Reader) DiscardMessage() {
	r.Buf = nil
}

// PopSlice returns a SimpleReader
// populated with the first n bytes from the buffer
// and discards those bytes.
func (r *Reader) PopSlice(n uint32) *Reader {
	s := SimpleReader(r.Buf[:n])
	r.Buf = r.Buf[n:]
	return s
}

// PopUint8 returns the next byte and advances the buffer.
func (r *Reader) PopUint8() uint8 {
	val := r.Buf[0]
	r.Buf = r.Buf[1:]
	return val
}

// PopUint16 reads a uint16 and advances the buffer.
func (r *Reader) PopUint16() uint16 {
	val := binary.BigEndian.Uint16(r.Buf[:2])
	r.Buf = r.Buf[2:]
	return val
}

// PopUint32 reads a uint32 and advances the buffer.
func (r *Reader) PopUint32() uint32 {
	val := binary.BigEndian.Uint32(r.Buf[:4])
	r.Buf = r.Buf[4:]
	return val
}

// PopUint64 reads a uint64 and advances the buffer.
func (r *Reader) PopUint64() uint64 {
	val := binary.BigEndian.Uint64(r.Buf[:8])
	r.Buf = r.Buf[8:]
	return val
}

// PopUUID reads a types.UUID and advances the buffer.
func (r *Reader) PopUUID() types.UUID {
	var id types.UUID
	copy(id[:], r.Buf[:16])
	r.Buf = r.Buf[16:]
	return id
}

// PopBytes reads a []byte and advances the buffer.
// The returned slice is owned by the buffer.
func (r *Reader) PopBytes() []byte {
	n := int(r.PopUint32())
	val := r.Buf[:n]
	r.Buf = r.Buf[n:]
	return val
}

// PopString reads a string and advances the buffer.
func (r *Reader) PopString() string {
	n := int(r.PopUint32())
	val := string(r.Buf[:n])
	r.Buf = r.Buf[n:]
	return val
}
