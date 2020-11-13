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

	"github.com/edgedb/edgedb-go/types"
)

// Message is a Reader with a Type attribute.
type Message struct {
	Bts  []byte
	Type uint8
}

// NewMessage returns a new Message.
func NewMessage(bts []byte) *Message {
	return &Message{Bts: bts}
}

// Finish asserts that the message has been fully read.
// It panics if it has not.
func (b *Message) Finish() {
	if len(b.Bts) > 0 {
		panic(fmt.Sprintf(
			"cannot finish: unread data in buffer (message type: 0x%x)",
			b.Type,
		))
	}
}

// Len returns the number of bytes remaining to be read.
func (b *Message) Len() int {
	return len(b.Bts)
}

// Discard skips the next n bytes.
func (b *Message) Discard(n int) {
	if len(b.Bts) < n {
		panic("buffer overread")
	}

	b.Bts = b.Bts[n:]
}

// PopUint8 returns the next byte and advances the buffer.
func (b *Message) PopUint8() uint8 {
	if len(b.Bts) < 1 {
		panic("buffer overread")
	}

	val := b.Bts[0]
	b.Bts = b.Bts[1:]
	return val
}

// PopUint16 reads a uint16 and advances the buffer.
func (b *Message) PopUint16() uint16 {
	if len(b.Bts) < 2 {
		panic("buffer overread")
	}

	val := binary.BigEndian.Uint16(b.Bts)
	b.Bts = b.Bts[2:]
	return val
}

// PopUint32 reads a uint32 and advances the buffer.
func (b *Message) PopUint32() uint32 {
	val := b.PeekUint32()
	b.Bts = b.Bts[4:]
	return val
}

// PeekUint32 reads a uint32 but does not advance the buffer.
func (b *Message) PeekUint32() uint32 {
	if len(b.Bts) < 4 {
		panic("buffer overread")
	}

	return binary.BigEndian.Uint32(b.Bts)
}

// PopUint64 reads a uint64 and advances the buffer.
func (b *Message) PopUint64() uint64 {
	if len(b.Bts) < 8 {
		panic("buffer overread")
	}

	val := binary.BigEndian.Uint64(b.Bts)
	b.Bts = b.Bts[8:]
	return val
}

// PopUUID reads a types.UUID and advances the buffer.
func (b *Message) PopUUID() types.UUID {
	if len(b.Bts) < 16 {
		panic("buffer overread")
	}

	var id types.UUID
	copy(id[:], b.Bts[:16])
	b.Bts = b.Bts[16:]
	return id
}

// PopBytes reads a []byte and advances the buffer.
// The returned slice is owned by the buffer.
func (b *Message) PopBytes() []byte {
	n := int(b.PopUint32())

	if len(b.Bts) < n {
		panic("buffer overread")
	}

	val := b.Bts[:n]
	b.Bts = b.Bts[n:]
	return val
}

// PopString reads a string and advances the buffer.
func (b *Message) PopString() string {
	return string(b.PopBytes())
}
