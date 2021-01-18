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

package codecs

import (
	"testing"
	"time"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeUUID(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 16, // data length
		0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8,
	})

	var result types.UUID
	(&UUID{}).DecodePtr(r, unsafe.Pointer(&result))

	expected := types.UUID{0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8}
	assert.Equal(t, expected, result)
}

func BenchmarkDecodeUUID(b *testing.B) {
	data := []byte{
		0, 0, 0, 16, // data length
		0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8,
	}
	r := buff.SimpleReader(data)

	var result types.UUID
	ptr := unsafe.Pointer(&result)
	codec := &UUID{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		codec.DecodePtr(r, ptr)
	}
}

func TestEncodeUUID(t *testing.T) {
	w := buff.NewWriter([]byte{})
	err := (&UUID{}).Encode(w, types.UUID{
		0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8,
	})
	require.Nil(t, err)

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 16, // data length
		0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8,
	}

	assert.Equal(t, expected, conn.written)
}

func BenchmarkEncodeUUID(b *testing.B) {
	w := buff.NewWriter([]byte{})
	id := types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	codec := &UUID{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = codec.Encode(w, id)
	}
}

func TestDecodeString(t *testing.T) {
	data := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}
	r := buff.SimpleReader(data)

	var result string
	(&Str{}).DecodePtr(r, unsafe.Pointer(&result))

	assert.Equal(t, "hello", result)

	// make sure that the string value is not tied to the buffer.
	data[5] = 0
	assert.Equal(t, "hello", result)
}

func BenchmarkDecodeString(b *testing.B) {
	data := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}
	r := buff.SimpleReader(data)

	var result string
	ptr := unsafe.Pointer(&result)
	codec := &Str{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		codec.DecodePtr(r, ptr)
	}
}

func TestEncodeString(t *testing.T) {
	w := buff.NewWriter([]byte{})
	err := (&Str{}).Encode(w, "hello")
	require.Nil(t, err)

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	assert.Equal(t, expected, conn.written)
}

func TestDecodeBytes(t *testing.T) {
	data := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}
	r := buff.SimpleReader(data)

	var result []byte
	(&Bytes{}).DecodePtr(r, unsafe.Pointer(&result))

	expected := []byte{104, 101, 108, 108, 111}
	assert.Equal(t, expected, result)

	// assert that memory is not shared with the buffer
	data[5] = 0
	assert.Equal(t, expected, result)
}

func BenchmarkDecodeBytes(b *testing.B) {
	data := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}
	r := buff.SimpleReader(data)

	var result []byte
	ptr := unsafe.Pointer(&result)
	codec := &Bytes{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		codec.DecodePtr(r, ptr)
	}
}

func TestEncodeBytes(t *testing.T) {
	w := buff.NewWriter([]byte{})
	err := (&Bytes{}).Encode(w, []byte{104, 101, 108, 108, 111})
	require.Nil(t, err)

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	assert.Equal(t, expected, conn.written)
}

func TestDecodeInt16(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 2, // data length
		0, 7, // int16
	})

	var result int16
	(&Int16{}).DecodePtr(r, unsafe.Pointer(&result))

	assert.Equal(t, int16(7), result)
}

func BenchmarkDecodeInt16(b *testing.B) {
	data := []byte{
		0, 0, 0, 2, // data length
		1, 2, // int16
	}
	r := buff.SimpleReader(data)

	var result int16
	ptr := unsafe.Pointer(&result)
	codec := &Int16{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		codec.DecodePtr(r, ptr)
	}
}

func TestEncodeInt16(t *testing.T) {
	w := buff.NewWriter([]byte{})
	err := (&Int16{}).Encode(w, int16(7))
	require.Nil(t, err)

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 2, // data length
		0, 7, // int16
	}

	assert.Equal(t, expected, conn.written)
}

func TestDecodeInt32(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 4, // data length
		0, 0, 0, 7, // int32
	})

	var result int32
	(&Int32{}).DecodePtr(r, unsafe.Pointer(&result))

	assert.Equal(t, int32(7), result)
}

func BenchmarkDecodeInt32(b *testing.B) {
	data := []byte{
		0, 0, 0, 4, // data length
		1, 2, 3, 4, // int32
	}
	r := buff.SimpleReader(data)

	var result int32
	ptr := unsafe.Pointer(&result)
	codec := &Int32{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		codec.DecodePtr(r, ptr)
	}
}

func TestEncodeInt32(t *testing.T) {
	w := buff.NewWriter([]byte{})
	err := (&Int32{}).Encode(w, int32(7))
	require.Nil(t, err)

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 4, // data length
		0, 0, 0, 7, // int32
	}

	assert.Equal(t, expected, conn.written)
}

func TestDecodeInt64(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 8, // data length
		1, 2, 3, 4, 5, 6, 7, 8, // int64
	})

	var result int64
	(&Int64{}).DecodePtr(r, unsafe.Pointer(&result))

	assert.Equal(t, int64(72623859790382856), result)
}

func BenchmarkDecodeInt64(b *testing.B) {
	data := []byte{
		0, 0, 0, 8, // data length
		1, 2, 3, 4, 5, 6, 7, 8, // int64
	}
	r := buff.SimpleReader(data)

	var result int64
	ptr := unsafe.Pointer(&result)
	codec := &Int64{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		codec.DecodePtr(r, ptr)
	}
}

func TestEncodeInt64(t *testing.T) {
	w := buff.NewWriter([]byte{})
	err := (&Int64{}).Encode(w, int64(27))
	require.Nil(t, err)

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 27, // int64
	}

	assert.Equal(t, expected, conn.written)
}

func TestDecodeFloat32(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 4, // data length
		0xc2, 0, 0, 0,
	})

	var result float32
	(&Float32{}).DecodePtr(r, unsafe.Pointer(&result))

	assert.Equal(t, float32(-32), result)
}

func BenchmarkDecodeFloat32(b *testing.B) {
	data := []byte{
		0, 0, 0, 4, // data length
		0xc2, 0, 0, 0,
	}
	r := buff.SimpleReader(data)

	var result float32
	ptr := unsafe.Pointer(&result)
	codec := &Float32{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		codec.DecodePtr(r, ptr)
	}
}

func TestEncodeFloat32(t *testing.T) {
	w := buff.NewWriter([]byte{})
	err := (&Float32{}).Encode(w, float32(-32))
	require.Nil(t, err)

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 4, // data length
		0xc2, 0, 0, 0,
	}

	assert.Equal(t, expected, conn.written)
}

func TestDecodeFloat64(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 8, // data length
		0xc0, 0x50, 0, 0, 0, 0, 0, 0,
	})

	var result float64
	(&Float64{}).DecodePtr(r, unsafe.Pointer(&result))

	assert.Equal(t, float64(-64), result)
}

func BenchmarkDecodeFloat64(b *testing.B) {
	data := []byte{
		0, 0, 0, 8, // data length
		0xc0, 0x50, 0, 0, 0, 0, 0, 0,
	}
	r := buff.SimpleReader(data)

	var result float64
	ptr := unsafe.Pointer(&result)
	codec := &Float64{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		codec.DecodePtr(r, ptr)
	}
}

func TestEncodeFloat64(t *testing.T) {
	w := buff.NewWriter([]byte{})
	err := (&Float64{}).Encode(w, float64(-64))
	require.Nil(t, err)

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 8, // data length
		0xc0, 0x50, 0, 0, 0, 0, 0, 0,
	}

	assert.Equal(t, expected, conn.written)
}

func TestDecodeBool(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 1, // data length
		1,
	})

	var result bool
	(&Bool{}).DecodePtr(r, unsafe.Pointer(&result))

	assert.Equal(t, true, result)
}

func BenchmarkDecodeBool(b *testing.B) {
	data := []byte{
		0, 0, 0, 1, // data length
		1,
	}
	r := buff.SimpleReader(data)

	var result bool
	ptr := unsafe.Pointer(&result)
	codec := &Bool{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Buf = data
		codec.DecodePtr(r, ptr)
	}
}

func TestEncodeBool(t *testing.T) {
	w := buff.NewWriter([]byte{})
	err := (&Bool{}).Encode(w, true)
	require.Nil(t, err)

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 1, // data length
		1,
	}

	assert.Equal(t, expected, conn.written)
}

func TestDecodeDateTime(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 8, // data length
		0xff, 0xfc, 0xa2, 0xfe, 0xc4, 0xc8, 0x20, 0x0,
	})

	var result time.Time
	(&DateTime{}).DecodePtr(r, unsafe.Pointer(&result))

	expected := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, result)
}

func TestEncodeDateTime(t *testing.T) {
	w := buff.NewWriter([]byte{})
	err := (&DateTime{}).Encode(w, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))
	require.Nil(t, err)

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 8, // data length
		0xff, 0xfc, 0xa2, 0xfe, 0xc4, 0xc8, 0x20, 0x0,
	}

	assert.Equal(t, expected, conn.written)
}

func TestDecodeDuration(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 0x10, // data length
		0, 0, 0, 0, 0, 0xf, 0x42, 0x40,
		0, 0, 0, 0, // reserved
		0, 0, 0, 0, // reserved
	})

	var result time.Duration
	(&Duration{}).DecodePtr(r, unsafe.Pointer(&result))

	assert.Equal(t, time.Duration(1_000_000_000), result)
}

func TestEncodeDuration(t *testing.T) {
	w := buff.NewWriter([]byte{})
	err := (&Duration{}).Encode(w, time.Duration(1_000_000_000))
	require.Nil(t, err)

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 0x10, // data length
		0, 0, 0, 0, 0, 0xf, 0x42, 0x40,
		0, 0, 0, 0, // reserved
		0, 0, 0, 0, // reserved
	}

	assert.Equal(t, expected, conn.written)
}

func TestDecodeJSON(t *testing.T) {
	t.SkipNow()

	r := buff.SimpleReader([]byte{
		0, 0, 0, 0x12, // data length
		1, // json format
		0x7b, 0x22, 0x68, 0x65,
		0x6c, 0x6c, 0x6f, 0x22,
		0x3a, 0x22, 0x77, 0x6f,
		0x72, 0x6c, 0x64, 0x22,
		0x7d,
	})

	var result interface{}
	(&JSON{}).DecodePtr(r, unsafe.Pointer(&result))
	expected := map[string]interface{}{"hello": "world"}

	assert.Equal(t, expected, result)
}

func TestEncodeJSON(t *testing.T) {
	w := buff.NewWriter([]byte{})
	(&JSON{}).Encode(w, map[string]string{"hello": "world"})

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0, 0, 0, 0x12, // data length
		1, // json format
		0x7b, 0x22, 0x68, 0x65,
		0x6c, 0x6c, 0x6f, 0x22,
		0x3a, 0x22, 0x77, 0x6f,
		0x72, 0x6c, 0x64, 0x22,
		0x7d,
	}

	assert.Equal(t, expected, conn.written)
}
