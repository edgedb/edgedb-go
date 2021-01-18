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

	"github.com/edgedb/edgedb-go/internal/message"
	"github.com/edgedb/edgedb-go/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestPushUint8(t *testing.T) {
	w := NewWriter([]byte{})
	w.PushUint8(0xff)
	assert.Equal(t, []byte{0xff}, w.buf)
}

func BenchmarkPushUint8(b *testing.B) {
	w := newBenchmarkWriter(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.PushUint8(0xff)
	}
}

func TestPushUint16(t *testing.T) {
	w := NewWriter([]byte{})
	w.PushUint16(0xffff)
	assert.Equal(t, []byte{0xff, 0xff}, w.buf)
}

func BenchmarkPushUint16(b *testing.B) {
	w := newBenchmarkWriter(2 * b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.PushUint16(0xffff)
	}
}

func TestPushUint32(t *testing.T) {
	w := NewWriter([]byte{})
	w.PushUint32(0xffffffff)
	assert.Equal(t, []byte{0xff, 0xff, 0xff, 0xff}, w.buf)
}

func BenchmarkPushUint32(b *testing.B) {
	w := newBenchmarkWriter(4 * b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.PushUint32(0xffffffff)
	}
}

func TestPushUint64(t *testing.T) {
	w := NewWriter([]byte{})
	w.PushUint64(0xffffffffffffffff)

	expected := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	assert.Equal(t, expected, w.buf)
}

func BenchmarkPushUint64(b *testing.B) {
	w := newBenchmarkWriter(8 * b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.PushUint64(0xffffffffffffffff)
	}
}

func TestPushUUID(t *testing.T) {
	w := NewWriter([]byte{})
	w.PushUUID(types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1})

	expected := []byte{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	assert.Equal(t, expected, w.buf)
}

func BenchmarkPushUUID(b *testing.B) {
	w := NewWriter([]byte{})
	id := types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.buf = w.buf[:0]
		w.PushUUID(id)
	}
}

func TestPushBytes(t *testing.T) {
	w := NewWriter([]byte{})
	w.PushBytes([]byte{7, 5})

	assert.Equal(t, []byte{0, 0, 0, 2, 7, 5}, w.buf)
}

func BenchmarkPushBytes(b *testing.B) {
	w := NewWriter([]byte{})
	bytes := []byte{1, 2, 3, 4}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.buf = w.buf[:0]
		w.PushBytes(bytes)
	}
}

func TestPushString(t *testing.T) {
	w := NewWriter([]byte{})
	w.PushString("hello")

	expected := []byte{0, 0, 0, 5, 0x68, 0x65, 0x6c, 0x6c, 0x6f}
	assert.Equal(t, expected, w.buf)
}

func BenchmarkPushString(b *testing.B) {
	w := newBenchmarkWriter(8 * b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.PushString("abcd")
	}
}

func TestBeginBytes(t *testing.T) {
	w := NewWriter([]byte{})

	msg := "cannot begin bytes: no current message"
	assert.PanicsWithValue(t, msg, func() { w.BeginBytes() })

	w.BeginMessage(message.Sync)
	w.BeginBytes()

	expected := []byte{message.Sync, 0, 0, 0, 0, 0, 0, 0, 0}
	assert.Equal(t, expected, w.buf)
}

func BenchmarkBeginBytes(b *testing.B) {
	w := newBenchmarkWriter(5 + 4*b.N)
	w.BeginMessage(message.Sync)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.BeginBytes()
	}
}

func TestEndBytes(t *testing.T) {
	w := NewWriter([]byte{})
	noBytesMsg := "cannot end bytes: no bytes in progress"
	assert.PanicsWithValue(t, noBytesMsg, func() { w.EndBytes() })

	w.BeginMessage(message.Sync)
	assert.PanicsWithValue(t, noBytesMsg, func() { w.EndBytes() })

	w.BeginBytes()
	w.PushUint32(9)
	w.EndBytes()

	assert.PanicsWithValue(t, noBytesMsg, func() { w.EndBytes() })
	expected := []byte{message.Sync, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 9}
	assert.Equal(t, expected, w.buf)
}

func BenchmarkBeginAndEndBytes(b *testing.B) {
	w := newBenchmarkWriter(5 + 4*b.N)
	w.BeginMessage(message.Sync)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.BeginBytes()
		w.EndBytes()
	}
}

func TestBeginMessage(t *testing.T) {
	w := NewWriter([]byte{})
	w.BeginMessage(message.Sync)

	msg := "cannot begin message: the previous message is not finished"
	assert.PanicsWithValue(t, msg, func() { w.BeginMessage(message.Sync) })
	assert.Equal(t, []byte{message.Sync, 0, 0, 0, 0}, w.buf)
}

func BenchmarkBeginAndEndMessage(b *testing.B) {
	w := newBenchmarkWriter(5 * b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.BeginMessage(message.Sync)
		w.EndMessage()
	}
}

func TestEndMessage(t *testing.T) {
	w := NewWriter([]byte{})
	noMsgMsg := "cannot end message: no current message"
	assert.PanicsWithValue(t, noMsgMsg, func() { w.EndMessage() })

	w.BeginMessage(message.Sync)
	w.BeginBytes()

	noBytesMsg := "cannot end message: bytes in progress"
	assert.PanicsWithValue(t, noBytesMsg, func() { w.EndMessage() })

	w.EndBytes()
	w.EndMessage()

	expected := []byte{message.Sync, 0, 0, 0, 8, 0, 0, 0, 0}
	assert.Equal(t, expected, w.buf)
	assert.PanicsWithValue(t, noMsgMsg, func() { w.EndMessage() })
}

func TestOnlySendsWhatWasPushed(t *testing.T) {
	w := NewWriter([]byte{})
	w.PushString("hello")

	f := &writerFixture{}
	assert.Nil(t, w.Send(f))

	expected := []byte{0, 0, 0, 5, 0x68, 0x65, 0x6c, 0x6c, 0x6f}
	assert.Equal(t, expected, f.written)
}

func TestSendsAllChuncks(t *testing.T) {
	w := NewWriter([]byte{})
	w.PushUint32(1)
	w.PushUint32(2)
	w.PushUint32(3)

	f := &writerFixture{}
	assert.Nil(t, w.Send(f))

	expected := []uint8{
		0x0, 0x0, 0x0, 0x1,
		0x0, 0x0, 0x0, 0x2,
		0x0, 0x0, 0x0, 0x3,
	}

	assert.Equal(t, expected, f.written)
}

func TestSendWithoutEndingMessage(t *testing.T) {
	w := NewWriter([]byte{})
	w.BeginMessage(message.Sync)
	assert.Panics(t, func() { _ = w.Send(&writerFixture{}) })
}
