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
)

func TestReset(t *testing.T) {
	buf := Writer{
		bts:  []byte{1, 2, 3},
		bPos: []int{1, 2, 3},
		mPos: 27,
	}

	buf.Reset()

	assert.Equal(t, []byte{}, buf.bts)
	assert.Equal(t, []int{}, buf.bPos)
	assert.Equal(t, 0, buf.mPos)
}

func TestPushUint8(t *testing.T) {
	buf := Writer{}
	buf.PushUint8(0xff)
	assert.Equal(t, []byte{0xff}, buf.bts)
}

func BenchmarkPushUint8(b *testing.B) {
	buf := Writer{}
	var n uint8 = 0xff

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushUint8(n)
	}
}

func TestPushUint16(t *testing.T) {
	buf := Writer{}
	buf.PushUint16(0xffff)
	assert.Equal(t, []byte{0xff, 0xff}, buf.bts)
}

func BenchmarkPushUint16(b *testing.B) {
	buf := Writer{}
	var n uint16 = 0xffff

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushUint16(n)
	}
}

func TestPushUint32(t *testing.T) {
	buf := Writer{}
	buf.PushUint32(0xffffffff)
	assert.Equal(t, []byte{0xff, 0xff, 0xff, 0xff}, buf.bts)
}

func BenchmarkPushUint32(b *testing.B) {
	buf := Writer{}
	var n uint32 = 0xffffffff

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushUint32(n)
	}
}

func TestPushUint64(t *testing.T) {
	buf := Writer{}
	buf.PushUint64(0xffffffffffffffff)
	expected := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	assert.Equal(t, expected, buf.bts)
}

func BenchmarkPushUint64(b *testing.B) {
	buf := Writer{}
	var n uint64 = 0xffffffffffffffff

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushUint64(n)
	}
}

func TestPushUUID(t *testing.T) {
	buf := Writer{}
	buf.PushUUID(types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1})
	expected := []byte{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	assert.Equal(t, expected, buf.bts)
}

func TestPushBytes(t *testing.T) {
	buf := Writer{}
	buf.PushBytes([]byte{7, 5})
	assert.Equal(t, []byte{0, 0, 0, 2, 7, 5}, buf.bts)
}

func BenchmarkPushBytes(b *testing.B) {
	buf := Writer{}
	data := []byte{1, 2, 3, 4}

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushBytes(data)
	}
}

func TestPushString(t *testing.T) {
	buf := Writer{}
	buf.PushString("hello")
	assert.Equal(t, []byte{0, 0, 0, 5, 0x68, 0x65, 0x6c, 0x6c, 0x6f}, buf.bts)
}

func BenchmarkPushString(b *testing.B) {
	buf := Writer{}
	data := "hello"

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.PushString(data)
	}
}

func TestBeginBytes(t *testing.T) {
	buf := Writer{}
	buf.BeginMessage(message.Sync)

	buf.BeginBytes()
	assert.Equal(t, []int{5}, buf.bPos)
	assert.Equal(t, []byte{message.Sync, 0, 0, 0, 0, 0, 0, 0, 0}, buf.bts)
}

func TestBeginBytesWithoutMessage(t *testing.T) {
	buf := Writer{}
	assert.Panics(t, func() { buf.BeginBytes() })
}

func TestEndBytes(t *testing.T) {
	buf := Writer{}
	buf.BeginMessage(message.Sync)
	buf.BeginBytes()
	buf.PushUint32(9)
	buf.EndBytes()

	expected := []byte{message.Sync, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 9}
	assert.Equal(t, expected, buf.bts)
	assert.Equal(t, []int{}, buf.bPos)
}

func TestEndBytesWithoutBeginingBytes(t *testing.T) {
	buf := Writer{}
	assert.Panics(t, func() { buf.EndBytes() })
}

func TestBeginMessage(t *testing.T) {
	buf := Writer{}
	buf.BeginMessage(message.Sync)
	assert.Equal(t, []byte{message.Sync, 0, 0, 0, 0}, buf.bts)
}

func BenchmarkBeginMessage(b *testing.B) {
	buf := Writer{}

	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.BeginMessage(message.Sync)
	}
}

func TestBeginMessageWithoutEndingPreviousMessage(t *testing.T) {
	buf := Writer{}
	buf.BeginMessage(message.Sync)
	assert.Panics(t, func() { buf.BeginMessage(message.Sync) })
}

func TestEndMessage(t *testing.T) {
	buf := Writer{}
	buf.BeginMessage(message.Sync)
	buf.EndMessage()
	assert.Equal(t, []byte{message.Sync, 0, 0, 0, 4}, buf.bts)
}

func BenchmarkEndMessage(b *testing.B) {
	buf := Writer{}
	buf.BeginMessage(message.Sync)
	data := buf.bts

	for i := 0; i < b.N; i++ {
		buf.bts = data
		buf.mPos = 1
		buf.EndMessage()
	}
}

func TestEndMessageWithoutBegining(t *testing.T) {
	buf := Writer{}
	assert.Panics(t, func() { buf.EndMessage() })
}

func TestEndMessageWithUnfinishedBytes(t *testing.T) {
	buf := Writer{}
	buf.BeginMessage(message.Sync)
	buf.BeginBytes()
	assert.Panics(t, func() { buf.EndMessage() })
}

func TestUnwrap(t *testing.T) {
	buf := Writer{bts: []byte{1, 2, 3, 4}}
	bts := buf.Unwrap()
	assert.Equal(t, []byte{1, 2, 3, 4}, *bts)

	*bts = []byte{7}
	assert.Equal(t, []byte{7}, buf.bts)
}

func BenchmarkUnwrap(b *testing.B) {
	buf := Writer{}
	buf.BeginMessage(message.Sync)
	buf.EndMessage()

	for i := 0; i < b.N; i++ {
		buf.Unwrap()
	}
}

func TestUnwrapWithoutEndingMessage(t *testing.T) {
	buf := Writer{}
	buf.BeginMessage(message.Sync)
	assert.Panics(t, func() { buf.Unwrap() })
}

func TestNext(t *testing.T) {
	buf := Writer{bts: []byte{0, 0, 0}}
	assert.Equal(t, true, buf.Next())

	buf.mPos = 3
	assert.Equal(t, false, buf.Next())
}

func TestPopMessage(t *testing.T) {
	buf := Writer{}

	buf.BeginMessage(message.Sync)
	buf.PushUint8(1)
	buf.EndMessage()

	buf.BeginMessage(message.Sync)
	buf.PushUint8(2)
	buf.EndMessage()

	msg := buf.PopMessage()
	assert.Equal(t, []byte{1}, msg.Bts)

	msg = buf.PopMessage()
	assert.Equal(t, []byte{2}, msg.Bts)

	assert.Panics(t, func() { buf.PopMessage() })
}
