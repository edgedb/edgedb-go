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

	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/types"
	"github.com/stretchr/testify/assert"
)

func TestDecodeUUID(t *testing.T) {
	buf := buff.New([]byte{
		0,
		0, 0, 0, 24,
		0, 0, 0, 16, // data length
		0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8,
	})
	buf.Next()

	codec := &UUID{}

	var result types.UUID
	codec.Decode(buf, unsafe.Pointer(&result))

	expected := types.UUID{0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8}
	assert.Equal(t, expected, result)
}

func BenchmarkDecodeUUID(b *testing.B) {
	data := []byte{
		0,
		0, 0, 0, 24,
		0, 0, 0, 16, // data length
		0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8,
	}
	buf := buff.New(data)
	buf.Next()

	var result types.UUID
	ptr := unsafe.Pointer(&result)
	codec := &UUID{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data[5:]
		codec.Decode(buf, ptr)
	}
}

func TestEncodeUUID(t *testing.T) {
	buf := buff.New(nil)
	(&UUID{}).Encode(buf, types.UUID{
		0, 1, 2, 3, 3, 2, 1, 0,
		8, 7, 6, 5, 5, 6, 7, 8,
	})

	expected := []byte{
		0, 0, 0, 16, // data length
		0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8,
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func BenchmarkEncodeUUID(b *testing.B) {
	codec := &UUID{}
	id := types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	buf := buff.New(make([]byte, 2000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		codec.Encode(buf, id)
	}
}

func TestDecodeString(t *testing.T) {
	data := []byte{
		0,
		0, 0, 0, 13,
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}
	buf := buff.New(data)
	buf.Next()

	var result string
	(&Str{}).Decode(buf, unsafe.Pointer(&result))

	assert.Equal(t, "hello", result)

	// make sure that the string value is not tied to the buffer.
	data[5] = 0
	assert.Equal(t, "hello", result)
}

func BenchmarkDecodeString(b *testing.B) {
	data := []byte{
		0,
		0, 0, 0, 13,
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}
	buf := buff.New(data)
	buf.Next()

	var result string
	ptr := unsafe.Pointer(&result)
	codec := &Str{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data[5:]
		codec.Decode(buf, ptr)
	}
}

func TestEncodeString(t *testing.T) {
	buf := buff.New(nil)
	(&Str{}).Encode(buf, "hello")

	expected := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func TestDecodeBytes(t *testing.T) {
	data := []byte{
		0,
		0, 0, 0, 13,
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}
	buf := buff.New(data)
	buf.Next()

	codec := Bytes{}

	var result []byte
	codec.Decode(buf, unsafe.Pointer(&result))

	expected := []byte{104, 101, 108, 108, 111}

	assert.Equal(t, expected, result)

	// assert that memory is not shared with the buffer
	data[5] = 0
	assert.Equal(t, expected, result)
}

func BenchmarkDecodeBytes(b *testing.B) {
	data := []byte{
		0,
		0, 0, 0, 13,
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}
	buf := buff.New(data)
	buf.Next()

	var result []byte
	ptr := unsafe.Pointer(&result)
	codec := &Bytes{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data[5:]
		codec.Decode(buf, ptr)
	}
}

func TestEncodeBytes(t *testing.T) {
	buf := buff.New(nil)
	(&Bytes{}).Encode(buf, []byte{104, 101, 108, 108, 111})

	expected := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func TestDecodeInt16(t *testing.T) {
	buf := buff.New([]byte{
		0,
		0, 0, 0, 10,
		0, 0, 0, 2, // data length
		0, 7, // int16
	})
	buf.Next()

	var result int16
	codec := Int16{}
	codec.Decode(buf, unsafe.Pointer(&result))

	assert.Equal(t, int16(7), result)
}

func BenchmarkDecodeInt16(b *testing.B) {
	data := []byte{
		0,
		0, 0, 0, 10,
		0, 0, 0, 2, // data length
		1, 2, // int16
	}
	buf := buff.New(data)
	buf.Next()

	var result int16
	ptr := unsafe.Pointer(&result)
	codec := &Int16{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data[5:]
		codec.Decode(buf, ptr)
	}
}

func TestEncodeInt16(t *testing.T) {
	buf := buff.New(nil)
	(&Int16{}).Encode(buf, int16(7))

	expected := []byte{
		0, 0, 0, 2, // data length
		0, 7, // int16
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func TestDecodeInt32(t *testing.T) {
	buf := buff.New([]byte{
		0,
		0, 0, 0, 12,
		0, 0, 0, 4, // data length
		0, 0, 0, 7, // int32
	})
	buf.Next()

	var result int32
	(&Int32{}).Decode(buf, unsafe.Pointer(&result))

	assert.Equal(t, int32(7), result)
}

func BenchmarkDecodeInt32(b *testing.B) {
	data := []byte{
		0,
		0, 0, 0, 12,
		0, 0, 0, 4, // data length
		1, 2, 3, 4, // int32
	}
	buf := buff.New(data)
	buf.Next()

	var result int32
	ptr := unsafe.Pointer(&result)
	codec := &Int32{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data[5:]
		codec.Decode(buf, ptr)
	}
}

func TestEncodeInt32(t *testing.T) {
	buf := buff.New(nil)
	(&Int32{}).Encode(buf, int32(7))

	expected := []byte{
		0, 0, 0, 4, // data length
		0, 0, 0, 7, // int32
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func TestDecodeInt64(t *testing.T) {
	buf := buff.New([]byte{
		0,
		0, 0, 0, 16,
		0, 0, 0, 8, // data length
		1, 2, 3, 4, 5, 6, 7, 8, // int64
	})
	buf.Next()

	var result int64
	(&Int64{}).Decode(buf, unsafe.Pointer(&result))

	assert.Equal(t, int64(72623859790382856), result)
}

func BenchmarkDecodeInt64(b *testing.B) {
	data := []byte{
		0,
		0, 0, 0, 16,
		0, 0, 0, 8, // data length
		1, 2, 3, 4, 5, 6, 7, 8, // int64
	}
	buf := buff.New(data)
	buf.Next()

	var result int64
	ptr := unsafe.Pointer(&result)
	codec := &Int64{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data[5:]
		codec.Decode(buf, ptr)
	}
}

func TestEncodeInt64(t *testing.T) {
	buf := buff.New(nil)
	(&Int64{}).Encode(buf, int64(27))

	expected := []byte{
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 27, // int64
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func TestDecodeFloat32(t *testing.T) {
	buf := buff.New([]byte{
		0,
		0, 0, 0, 12,
		0, 0, 0, 4, // data length
		0xc2, 0, 0, 0,
	})
	buf.Next()

	var result float32
	codec := &Float32{}
	codec.Decode(buf, unsafe.Pointer(&result))

	assert.Equal(t, float32(-32), result)
}

func BenchmarkDecodeFloat32(b *testing.B) {
	data := []byte{
		0,
		0, 0, 0, 12,
		0, 0, 0, 4, // data length
		0xc2, 0, 0, 0,
	}
	buf := buff.New(data)
	buf.Next()

	var result float32
	ptr := unsafe.Pointer(&result)
	codec := &Float32{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data[5:]
		codec.Decode(buf, ptr)
	}
}

func TestEncodeFloat32(t *testing.T) {
	buf := buff.New(nil)
	(&Float32{}).Encode(buf, float32(-32))

	expected := []byte{
		0, 0, 0, 4, // data length
		0xc2, 0, 0, 0,
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func TestDecodeFloat64(t *testing.T) {
	buf := buff.New([]byte{
		0,
		0, 0, 0, 16,
		0, 0, 0, 8, // data length
		0xc0, 0x50, 0, 0, 0, 0, 0, 0,
	})
	buf.Next()

	var result float64
	codec := &Float64{}
	codec.Decode(buf, unsafe.Pointer(&result))

	assert.Equal(t, float64(-64), result)
}

func BenchmarkDecodeFloat64(b *testing.B) {
	data := []byte{
		0,
		0, 0, 0, 16,
		0, 0, 0, 8, // data length
		0xc0, 0x50, 0, 0, 0, 0, 0, 0,
	}
	buf := buff.New(data)
	buf.Next()

	var result float64
	ptr := unsafe.Pointer(&result)
	codec := &Float64{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data[5:]
		codec.Decode(buf, ptr)
	}
}

func TestEncodeFloat64(t *testing.T) {
	buf := buff.New(nil)
	(&Float64{}).Encode(buf, float64(-64))

	expected := []byte{
		0, 0, 0, 8, // data length
		0xc0, 0x50, 0, 0, 0, 0, 0, 0,
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func TestDecodeBool(t *testing.T) {
	buf := buff.New([]byte{
		0,
		0, 0, 0, 9,
		0, 0, 0, 1, // data length
		1,
	})
	buf.Next()

	var result bool
	codec := &Bool{}
	codec.Decode(buf, unsafe.Pointer(&result))

	assert.Equal(t, true, result)
}

func BenchmarkDecodeBool(b *testing.B) {
	data := []byte{
		0,
		0, 0, 0, 9,
		0, 0, 0, 1, // data length
		1,
	}
	buf := buff.New(data)
	buf.Next()

	var result bool
	ptr := unsafe.Pointer(&result)
	codec := &Bool{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data[5:]
		codec.Decode(buf, ptr)
	}
}

func TestEncodeBool(t *testing.T) {
	buf := buff.New(nil)
	(&Bool{}).Encode(buf, true)

	expected := []byte{
		0, 0, 0, 1, // data length
		1,
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func TestDecodeDateTime(t *testing.T) {
	buf := buff.New([]byte{
		0,
		0, 0, 0, 16,
		0, 0, 0, 8, // data length
		0xff, 0xfc, 0xa2, 0xfe, 0xc4, 0xc8, 0x20, 0x0,
	})
	buf.Next()

	var result time.Time
	codec := &DateTime{}
	codec.Decode(buf, unsafe.Pointer(&result))

	expected := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, result)
}

func TestEncodeDateTime(t *testing.T) {
	buf := buff.New(nil)
	(&DateTime{}).Encode(buf, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))

	expected := []byte{
		0, 0, 0, 8, // data length
		0xff, 0xfc, 0xa2, 0xfe, 0xc4, 0xc8, 0x20, 0x0,
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func TestDecodeDuration(t *testing.T) {
	buf := buff.New([]byte{
		0,
		0, 0, 0, 24,
		0, 0, 0, 0x10, // data length
		0, 0, 0, 0, 0, 0xf, 0x42, 0x40,
		0, 0, 0, 0, // reserved
		0, 0, 0, 0, // reserved
	})
	buf.Next()

	var result time.Duration
	codec := &Duration{}
	codec.Decode(buf, unsafe.Pointer(&result))

	assert.Equal(t, time.Duration(1_000_000_000), result)
}

func TestEncodeDuration(t *testing.T) {
	buf := buff.New(nil)
	(&Duration{}).Encode(buf, time.Duration(1_000_000_000))

	expected := []byte{
		0, 0, 0, 0x10, // data length
		0, 0, 0, 0, 0, 0xf, 0x42, 0x40,
		0, 0, 0, 0, // reserved
		0, 0, 0, 0, // reserved
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func TestDecodeJSON(t *testing.T) {
	// todo
	t.SkipNow()

	buf := buff.New([]byte{
		0, 0, 0, 0x12, // data length
		1, // json format
		0x7b, 0x22, 0x68, 0x65,
		0x6c, 0x6c, 0x6f, 0x22,
		0x3a, 0x22, 0x77, 0x6f,
		0x72, 0x6c, 0x64, 0x22,
		0x7d,
	})

	var result interface{}
	(&JSON{}).Decode(buf, unsafe.Pointer(&result))
	expected := map[string]interface{}{"hello": "world"}

	assert.Equal(t, expected, result)
}

func TestEncodeJSON(t *testing.T) {
	buf := buff.New(nil)
	(&JSON{}).Encode(buf, map[string]string{"hello": "world"})

	expected := []byte{
		0, 0, 0, 0x12, // data length
		1, // json format
		0x7b, 0x22, 0x68, 0x65,
		0x6c, 0x6c, 0x6f, 0x22,
		0x3a, 0x22, 0x77, 0x6f,
		0x72, 0x6c, 0x64, 0x22,
		0x7d,
	}

	assert.Equal(t, expected, *buf.Unwrap())
}
