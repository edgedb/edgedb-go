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
	"testing"

	"github.com/edgedb/edgedb-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessagePreventsOverread(t *testing.T) {
	data := make([]byte, 10)
	msg := NewMessage(data[:2])

	assert.Panics(t, func() { _ = msg.Bts[:4] })
}

func TestDiscard(t *testing.T) {
	data := []byte{0, 0, 0, 0, 0}
	msg := NewMessage(data[:2])
	msg.Discard(2)
	require.Equal(t, []byte{}, msg.Bts)

	assert.Panics(t, func() { msg.Discard(2) })
}

func BenchmarkDiscard(b *testing.B) {
	data := []byte{0xff, 0xff, 0xff, 0xff}
	msg := NewMessage(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.Bts = data
		msg.Discard(4)
	}
}

func TestPopUint8(t *testing.T) {
	data := []byte{0xff, 0, 0}
	msg := NewMessage(data[:1])
	var expected uint8 = 0xff
	require.Equal(t, expected, msg.PopUint8())
	require.Equal(t, []byte{}, msg.Bts)

	assert.Panics(t, func() { msg.PopUint8() })
}

func BenchmarkPopUint8(b *testing.B) {
	data := []byte{0xff}
	msg := NewMessage(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.Bts = data
		msg.PopUint8()
	}
}

func TestPopUint16(t *testing.T) {
	data := []byte{0xff, 0xff, 0, 0, 0, 0}
	msg := NewMessage(data[:2])
	var expected uint16 = 0xffff
	require.Equal(t, expected, msg.PopUint16())
	require.Equal(t, []byte{}, msg.Bts)

	assert.Panics(t, func() { msg.PopUint16() })
}

func BenchmarkPopUint16(b *testing.B) {
	data := []byte{0xff, 0xff}
	msg := NewMessage(nil)

	for i := 0; i < b.N; i++ {
		msg.Bts = data
		msg.PopUint16()
	}
}

func TestPopUint32(t *testing.T) {
	data := []byte{0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0}
	msg := NewMessage(data[:4])
	var expected uint32 = 0xffffffff
	require.Equal(t, expected, msg.PopUint32())
	require.Equal(t, []byte{}, msg.Bts)

	assert.Panics(t, func() { msg.PopUint32() })
}

func BenchmarkPopUint32(b *testing.B) {
	data := []byte{0xff, 0xff, 0xff, 0xff}
	msg := NewMessage(nil)

	for i := 0; i < b.N; i++ {
		msg.Bts = data
		msg.PopUint32()
	}
}

func TestPeekUint32(t *testing.T) {
	msg := NewMessage([]byte{0xff, 0xff, 0xff, 0xff})
	assert.Equal(t, uint32(0xffffffff), msg.PeekUint32())
	assert.Equal(t, []byte{0xff, 0xff, 0xff, 0xff}, msg.Bts)
}

func BenchmarkPeekUint32(b *testing.B) {
	msg := NewMessage([]byte{0xff, 0xff, 0xff, 0xff})

	for i := 0; i < b.N; i++ {
		msg.PeekUint32()
	}
}

func TestPopUint64(t *testing.T) {
	data := []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0, 0, 0, 0, 0, 0, 0, 0,
	}
	msg := NewMessage(data[:8])

	var expected uint64 = 0xffffffffffffffff
	require.Equal(t, expected, msg.PopUint64())
	require.Equal(t, []byte{}, msg.Bts)

	assert.Panics(t, func() { msg.PopUint64() })
}

func BenchmarkPopUint64(b *testing.B) {
	data := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	msg := NewMessage(nil)

	for i := 0; i < b.N; i++ {
		msg.Bts = data
		msg.PopUint64()
	}
}

func TestPopUUID(t *testing.T) {
	data := []byte{
		1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	msg := NewMessage(data[:16])

	expected := types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	require.Equal(t, expected, msg.PopUUID())
	require.Equal(t, []byte{}, msg.Bts)

	assert.Panics(t, func() { msg.PopUUID() })
}

func BenchmarkPopUUID(b *testing.B) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	msg := NewMessage(nil)

	for i := 0; i < b.N; i++ {
		msg.Bts = data
		msg.PopUUID()
	}
}

func TestPopBytes(t *testing.T) {
	data := []byte{
		0, 0, 0, 4, 1, 2, 3, 5,
		0, 0, 0, 4, 0, 0, 0, 0,
	}
	msg := NewMessage(data[:8])

	require.Equal(t, []byte{1, 2, 3, 5}, msg.PopBytes())
	require.Equal(t, []byte{}, msg.Bts)

	assert.Panics(t, func() { msg.PopBytes() })
}

func BenchmarkPopBytes(b *testing.B) {
	data := []byte{0, 0, 0, 4, 1, 2, 3, 5}
	msg := NewMessage(nil)

	for i := 0; i < b.N; i++ {
		msg.Bts = data
		msg.PopBytes()
	}
}

func TestPopString(t *testing.T) {
	data := []byte{
		0, 0, 0, 5, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
		0, 0, 0, 4, 0, 0, 0, 0,
	}
	msg := NewMessage(data[:9])
	require.Equal(t, "hello", msg.PopString())
	require.Equal(t, []byte{}, msg.Bts)

	assert.Panics(t, func() { msg.PopString() })
}

func BenchmarkPopString(b *testing.B) {
	data := []byte{0, 0, 0, 5, 0x68, 0x65, 0x6c, 0x6c, 0x6f}
	msg := NewMessage(data)

	for i := 0; i < b.N; i++ {
		msg.Bts = data
		msg.PopString()
	}
}

func TestFinish(t *testing.T) {
	data := []byte{0xff, 0, 0, 0, 0}
	msg := &Message{Bts: data[:1], Type: 0xa}
	assert.PanicsWithValue(
		t,
		"cannot finish: unread data in buffer (message type: 0xa)",
		func() { msg.Finish() },
	)

	msg.PopUint8()
	msg.Finish()
}
