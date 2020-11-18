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

	"github.com/edgedb/edgedb-go/protocol/message"
	"github.com/edgedb/edgedb-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNext(t *testing.T) {
	data := []byte{0xa, 0, 0, 0, 8, 1, 2, 3, 4, 0, 0, 0, 0}
	buf := New(data[:9])

	assert.True(t, buf.Next())
	assert.Equal(t, uint8(0xa), buf.MsgType)
	assert.Equal(t, []byte{1, 2, 3, 4}, buf.Msg)
	assert.Panics(t, func() { _ = buf.Msg[:5] })
	assert.PanicsWithValue(
		t,
		"cannot finish: unread data in buffer (message type: 0xa)",
		func() { buf.Next() },
	)

	buf.PopUint32()

	assert.False(t, buf.Next())
	assert.Equal(t, uint8(0), buf.MsgType)
	assert.Equal(t, []byte{}, buf.Msg)
	assert.Panics(t, func() { _ = buf.Msg[:1] })
}

func TestDiscard(t *testing.T) {
	data := []byte{1, 0, 0, 0, 6, 0, 0, 0, 0, 0}
	buf := New(data[:7])
	buf.Next()
	buf.Discard(2)
	require.Equal(t, []byte{}, buf.Msg)

	assert.Panics(t, func() { buf.Discard(2) })
}

func BenchmarkDiscard(b *testing.B) {
	data := []byte{0xff, 0xff, 0xff, 0xff}
	buf := New(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data
		buf.Discard(4)
	}
}

func TestPopUint8(t *testing.T) {
	data := []byte{0, 0, 0, 0, 5, 0xff, 0, 0}
	buf := New(data[:6])
	buf.Next()

	var expected uint8 = 0xff
	require.Equal(t, expected, buf.PopUint8())
	require.Equal(t, []byte{}, buf.Msg)

	assert.Panics(t, func() { buf.PopUint8() })
}

func BenchmarkPopUint8(b *testing.B) {
	data := []byte{0xff}
	buf := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data
		buf.PopUint8()
	}
}

func TestPopUint16(t *testing.T) {
	data := []byte{0, 0, 0, 0, 6, 0xff, 0xff, 0, 0, 0, 0}
	buf := New(data[:7])
	buf.Next()

	var expected uint16 = 0xffff
	require.Equal(t, expected, buf.PopUint16())
	require.Equal(t, []byte{}, buf.Msg)

	assert.Panics(t, func() { buf.PopUint16() })
}

func BenchmarkPopUint16(b *testing.B) {
	data := []byte{0xff, 0xff}
	buf := New(nil)

	for i := 0; i < b.N; i++ {
		buf.Msg = data
		buf.PopUint16()
	}
}

func TestPopUint32(t *testing.T) {
	data := []byte{0, 0, 0, 0, 8, 0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0}
	buf := New(data[:9])
	buf.Next()

	var expected uint32 = 0xffffffff
	require.Equal(t, expected, buf.PopUint32())
	require.Equal(t, []byte{}, buf.Msg)

	assert.Panics(t, func() { buf.PopUint32() })
}

func BenchmarkPopUint32(b *testing.B) {
	data := []byte{0xff, 0xff, 0xff, 0xff}
	buf := New(nil)

	for i := 0; i < b.N; i++ {
		buf.Msg = data
		buf.PopUint32()
	}
}

func TestPeekUint32(t *testing.T) {
	data := []byte{0, 0, 0, 0, 8, 0xff, 0xff, 0xff, 0xff}
	buf := New(data[:9])
	buf.Next()

	assert.Equal(t, uint32(0xffffffff), buf.PeekUint32())
	assert.Equal(t, []byte{0xff, 0xff, 0xff, 0xff}, buf.Msg)
}

func BenchmarkPeekUint32(b *testing.B) {
	data := []byte{0, 0, 0, 0, 8, 0xff, 0xff, 0xff, 0xff}
	buf := New(data[:9])
	buf.Next()

	for i := 0; i < b.N; i++ {
		buf.PeekUint32()
	}
}

func TestPopUint64(t *testing.T) {
	data := []byte{
		0, 0, 0, 0, 12,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0, 0, 0, 0, 0, 0, 0, 0,
	}
	buf := New(data[:13])
	buf.Next()

	var expected uint64 = 0xffffffffffffffff
	require.Equal(t, expected, buf.PopUint64())
	require.Equal(t, []byte{}, buf.Msg)

	assert.Panics(t, func() { buf.PopUint64() })
}

func BenchmarkPopUint64(b *testing.B) {
	data := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	buf := New(nil)

	for i := 0; i < b.N; i++ {
		buf.Msg = data
		buf.PopUint64()
	}
}

func TestPopUUID(t *testing.T) {
	data := []byte{
		0, 0, 0, 0, 20,
		1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
	buf := New(data[:21])
	buf.Next()

	expected := types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	require.Equal(t, expected, buf.PopUUID())
	require.Equal(t, []byte{}, buf.Msg)

	assert.Panics(t, func() { buf.PopUUID() })
}

func BenchmarkPopUUID(b *testing.B) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	buf := New(nil)

	for i := 0; i < b.N; i++ {
		buf.Msg = data
		buf.PopUUID()
	}
}

func TestPopBytes(t *testing.T) {
	data := []byte{
		0, 0, 0, 0, 12,
		0, 0, 0, 4, 1, 2, 3, 5,
		0, 0, 0, 4, 0, 0, 0, 0,
	}
	buf := New(data[:13])
	buf.Next()

	require.Equal(t, []byte{1, 2, 3, 5}, buf.PopBytes())
	require.Equal(t, []byte{}, buf.Msg)

	assert.Panics(t, func() { buf.PopBytes() })
}

func BenchmarkPopBytes(b *testing.B) {
	data := []byte{0, 0, 0, 4, 1, 2, 3, 5}
	buf := New(nil)

	for i := 0; i < b.N; i++ {
		buf.Msg = data
		buf.PopBytes()
	}
}

func TestPopString(t *testing.T) {
	data := []byte{
		0, 0, 0, 0, 13,
		0, 0, 0, 5, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
		0, 0, 0, 4, 0, 0, 0, 0,
	}
	buf := New(data[:14])
	buf.Next()

	require.Equal(t, "hello", buf.PopString())
	require.Equal(t, []byte{}, buf.Msg)

	assert.Panics(t, func() { buf.PopString() })
}

func BenchmarkPopString(b *testing.B) {
	data := []byte{0, 0, 0, 5, 0x68, 0x65, 0x6c, 0x6c, 0x6f}
	buf := New(data)

	for i := 0; i < b.N; i++ {
		buf.Msg = data
		buf.PopString()
	}
}

func TestFinish(t *testing.T) {
	data := []byte{0xa, 0, 0, 0, 5, 0xff, 0, 0, 0, 0}
	buf := New(data[:6])
	buf.Next()

	assert.PanicsWithValue(
		t,
		"cannot finish: unread data in buffer (message type: 0xa)",
		func() { buf.Finish() },
	)

	buf.PopUint8()
	buf.Finish()
}

func TestReset(t *testing.T) {
	buf := New([]byte{1, 2, 3})
	buf.bPos = []int{1, 2, 3}
	buf.mPos = 27
	buf.Msg = buf.payload

	buf.Reset()

	assert.Equal(t, []byte{}, buf.payload)
	assert.Equal(t, []byte{}, buf.Msg)
	assert.Equal(t, []int{}, buf.bPos)
	assert.Equal(t, 0, buf.mPos)
}

func TestPushUint8(t *testing.T) {
	buf := New([]byte{})
	buf.PushUint8(0xff)
	assert.Equal(t, []byte{0xff}, buf.payload)
}

func BenchmarkPushUint8(b *testing.B) {
	buf := New([]byte{})
	var n uint8 = 0xff

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushUint8(n)
	}
}

func TestPushUint16(t *testing.T) {
	buf := New([]byte{})
	buf.PushUint16(0xffff)
	assert.Equal(t, []byte{0xff, 0xff}, buf.payload)
}

func BenchmarkPushUint16(b *testing.B) {
	buf := New([]byte{})
	var n uint16 = 0xffff

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushUint16(n)
	}
}

func TestPushUint32(t *testing.T) {
	buf := New([]byte{})
	buf.PushUint32(0xffffffff)
	assert.Equal(t, []byte{0xff, 0xff, 0xff, 0xff}, buf.payload)
}

func BenchmarkPushUint32(b *testing.B) {
	buf := New([]byte{})
	var n uint32 = 0xffffffff

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushUint32(n)
	}
}

func TestPushUint64(t *testing.T) {
	buf := New([]byte{})
	buf.PushUint64(0xffffffffffffffff)
	expected := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	assert.Equal(t, expected, buf.payload)
}

func BenchmarkPushUint64(b *testing.B) {
	buf := New([]byte{})
	var n uint64 = 0xffffffffffffffff

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushUint64(n)
	}
}

func TestPushUUID(t *testing.T) {
	buf := New([]byte{})
	buf.PushUUID(types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1})
	expected := []byte{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	assert.Equal(t, expected, buf.payload)
}

func TestPushBytes(t *testing.T) {
	buf := New([]byte{})
	buf.PushBytes([]byte{7, 5})
	assert.Equal(t, []byte{0, 0, 0, 2, 7, 5}, buf.payload)
}

func BenchmarkPushBytes(b *testing.B) {
	buf := New([]byte{})
	data := []byte{1, 2, 3, 4}

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushBytes(data)
	}
}

func TestPushString(t *testing.T) {
	buf := New([]byte{})
	buf.PushString("hello")

	expected := []byte{0, 0, 0, 5, 0x68, 0x65, 0x6c, 0x6c, 0x6f}
	assert.Equal(t, expected, buf.payload)
}

func BenchmarkPushString(b *testing.B) {
	buf := New([]byte{})
	data := "hello"

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushString(data)
	}
}

func TestBeginBytes(t *testing.T) {
	buf := New([]byte{})
	buf.BeginMessage(message.Sync)

	buf.BeginBytes()
	assert.Equal(t, []int{5}, buf.bPos)
	assert.Equal(t, []byte{message.Sync, 0, 0, 0, 0, 0, 0, 0, 0}, buf.payload)
}

func TestBeginBytesWithoutMessage(t *testing.T) {
	buf := New([]byte{})
	assert.Panics(t, func() { buf.BeginBytes() })
}

func TestEndBytes(t *testing.T) {
	buf := New([]byte{})
	buf.BeginMessage(message.Sync)
	buf.BeginBytes()
	buf.PushUint32(9)
	buf.EndBytes()

	expected := []byte{message.Sync, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 9}
	assert.Equal(t, expected, buf.payload)
	assert.Equal(t, []int{}, buf.bPos)
}

func TestEndBytesWithoutBeginingBytes(t *testing.T) {
	buf := New([]byte{})
	assert.Panics(t, func() { buf.EndBytes() })
}

func TestBeginMessage(t *testing.T) {
	buf := New([]byte{})
	buf.BeginMessage(message.Sync)
	assert.Equal(t, []byte{message.Sync, 0, 0, 0, 0}, buf.payload)
}

func BenchmarkBeginMessage(b *testing.B) {
	buf := New([]byte{})

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.BeginMessage(message.Sync)
	}
}

func TestBeginMessageWithoutEndingPreviousMessage(t *testing.T) {
	buf := New([]byte{})
	buf.BeginMessage(message.Sync)
	assert.Panics(t, func() { buf.BeginMessage(message.Sync) })
}

func TestEndMessage(t *testing.T) {
	buf := New([]byte{})
	buf.BeginMessage(message.Sync)
	buf.EndMessage()
	assert.Equal(t, []byte{message.Sync, 0, 0, 0, 4}, buf.payload)
}

func BenchmarkEndMessage(b *testing.B) {
	buf := New([]byte{})
	buf.BeginMessage(message.Sync)
	data := buf.payload

	for i := 0; i < b.N; i++ {
		buf.payload = data
		buf.mPos = 1
		buf.EndMessage()
	}
}

func TestEndMessageWithoutBegining(t *testing.T) {
	buf := New([]byte{})
	assert.Panics(t, func() { buf.EndMessage() })
}

func TestEndMessageWithUnfinishedBytes(t *testing.T) {
	buf := New([]byte{})
	buf.BeginMessage(message.Sync)
	buf.BeginBytes()
	assert.Panics(t, func() { buf.EndMessage() })
}

func TestUnwrap(t *testing.T) {
	buf := New([]byte{1, 2, 3, 4})
	bts := buf.Unwrap()
	assert.Equal(t, []byte{1, 2, 3, 4}, *bts)

	*bts = []byte{7}
	assert.Equal(t, []byte{7}, buf.payload)
}

func BenchmarkUnwrap(b *testing.B) {
	buf := New([]byte{})
	buf.BeginMessage(message.Sync)
	buf.EndMessage()

	for i := 0; i < b.N; i++ {
		buf.Unwrap()
	}
}

func TestUnwrapWithoutEndingMessage(t *testing.T) {
	buf := New([]byte{})
	buf.BeginMessage(message.Sync)
	assert.Panics(t, func() { buf.Unwrap() })
}
