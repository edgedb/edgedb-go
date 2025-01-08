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
	"testing"

	types "github.com/edgedb/edgedb-go/internal/geltypes"
	"github.com/edgedb/edgedb-go/internal/soc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNext(t *testing.T) {
	toBeDeserialized := make(chan *soc.Data, 1)
	toBeDeserialized <- &soc.Data{Buf: []byte{0xa, 0, 0, 0, 8, 1, 2, 3, 4}}
	r := NewReader(toBeDeserialized)

	assert.True(t, r.Next(nil))
	assert.Equal(t, uint8(0xa), r.MsgType)

	expected := "cannot finish: unread data in buffer (message type: 0xa)"
	assert.False(t, r.Next(nil))
	assert.EqualError(t, r.Err, expected)

	assert.Equal(t, uint32(0x1020304), r.PopUint32())
	assert.Panics(t, func() { r.Discard(1) })

	doneReadingSignal := make(chan struct{}, 1)
	doneReadingSignal <- struct{}{}
	assert.False(t, r.Next(doneReadingSignal))
	assert.Equal(t, uint8(0), r.MsgType)
	assert.Panics(t, func() { r.Discard(1) })
}

func TestDiscard(t *testing.T) {
	r := SimpleReader([]byte{1, 2, 3, 4})
	r.Discard(2)

	require.Equal(t, uint16(0x304), r.PopUint16())
	assert.Panics(t, func() { r.Discard(1) })
}

func BenchmarkDiscard(b *testing.B) {
	r := SimpleReader(newBenchmarkMessage(4 * b.N))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Discard(4)
	}
}

func TestPopUint8(t *testing.T) {
	r := SimpleReader([]byte{0xff, 1})

	require.Equal(t, uint8(0xff), r.PopUint8())
	require.Equal(t, uint8(1), r.PopUint8())
	assert.Panics(t, func() { r.PopUint8() })
}

func BenchmarkPopUint8(b *testing.B) {
	r := SimpleReader(newBenchmarkMessage(b.N))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.PopUint8()
	}
}

func TestPopUint16(t *testing.T) {
	r := SimpleReader([]byte{0xff, 0xff, 1})

	require.Equal(t, uint16(0xffff), r.PopUint16())
	require.Equal(t, uint8(1), r.PopUint8())
	assert.Panics(t, func() { r.PopUint16() })
}

func BenchmarkPopUint16(b *testing.B) {
	r := SimpleReader(newBenchmarkMessage(2 * b.N))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.PopUint16()
	}
}

func TestPopUint32(t *testing.T) {
	r := SimpleReader(
		[]byte{0xff, 0xff, 0xff, 0xff, 1},
	)

	require.Equal(t, uint32(0xffffffff), r.PopUint32())
	require.Equal(t, uint8(1), r.PopUint8())
	assert.Panics(t, func() { r.PopUint32() })
}

func BenchmarkPopUint32(b *testing.B) {
	r := SimpleReader(newBenchmarkMessage(4 * b.N))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.PopUint32()
	}
}

func TestPopUint64(t *testing.T) {
	r := SimpleReader([]byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		1,
	})

	require.Equal(t, uint64(0xffffffffffffffff), r.PopUint64())
	require.Equal(t, uint8(1), r.PopUint8())
	assert.Panics(t, func() { r.PopUint64() })
}

func BenchmarkPopUint64(b *testing.B) {
	r := SimpleReader(newBenchmarkMessage(8 * b.N))

	for i := 0; i < b.N; i++ {
		r.PopUint64()
	}
}

func TestPopUUID(t *testing.T) {
	r := SimpleReader([]byte{
		1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1,
		1,
	})

	expected := types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	require.Equal(t, expected, r.PopUUID())
	require.Equal(t, uint8(1), r.PopUint8())
	assert.Panics(t, func() { r.PopUUID() })
}

func BenchmarkPopUUID(b *testing.B) {
	data := []byte{
		1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1,
		1,
	}
	r := SimpleReader(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		r.PopUUID()
	}
}

func TestPopBytes(t *testing.T) {
	r := SimpleReader([]byte{
		0, 0, 0, 4, 1, 2, 3, 5,
		6,
	})

	require.Equal(t, []byte{1, 2, 3, 5}, r.PopBytes())
	require.Equal(t, uint8(6), r.PopUint8())
	assert.Panics(t, func() { r.PopBytes() })
}

func BenchmarkPopBytes(b *testing.B) {
	data := []byte{0, 0, 0, 4, 0xff, 0xff, 0xff, 0xff}
	r := SimpleReader(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		r.PopBytes()
	}
}

func TestPopString(t *testing.T) {
	r := SimpleReader([]byte{
		0, 0, 0, 5, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
		1,
	})

	require.Equal(t, "hello", r.PopString())
	require.Equal(t, uint8(1), r.PopUint8())
	assert.Panics(t, func() { r.PopString() })
}

func BenchmarkPopString(b *testing.B) {
	data := []byte{0, 0, 0, 4, 0x30, 0x78, 0x66, 0x66}
	r := SimpleReader(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		r.PopString()
	}
}

func BenchmarkAssignBuf(b *testing.B) {
	data := []byte{0, 0, 0, 4, 0x30, 0x78, 0x66, 0x66}
	r := SimpleReader(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
	}
}
