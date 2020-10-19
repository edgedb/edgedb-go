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

	"github.com/edgedb/edgedb-go/edgedb/types"
	"github.com/stretchr/testify/assert"
)

func TestDecodeUUID(t *testing.T) {
	bts := []byte{
		0, 0, 0, 16, // data length
		0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8,
	}

	id := (&UUID{}).Decode(&bts)

	expected := types.UUID{0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8}
	assert.Equal(t, expected, id)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeUUID(t *testing.T) {
	bts := []byte{}
	(&UUID{}).Encode(&bts, types.UUID{
		0, 1, 2, 3, 3, 2, 1, 0,
		8, 7, 6, 5, 5, 6, 7, 8,
	})

	expected := []byte{
		0, 0, 0, 16, // data length
		0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeString(t *testing.T) {
	bts := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	result := (&String{}).Decode(&bts)

	assert.Equal(t, "hello", result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeString(t *testing.T) {
	bts := []byte{}
	(&String{}).Encode(&bts, "hello")

	expected := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeBytes(t *testing.T) {
	bts := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	result := (&Bytes{}).Decode(&bts)
	expected := []byte{104, 101, 108, 108, 111}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeBytes(t *testing.T) {
	bts := []byte{}
	(&Bytes{}).Encode(&bts, []byte{104, 101, 108, 108, 111})

	expected := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeInt16(t *testing.T) {
	bts := []byte{
		0, 0, 0, 2, // data length
		0, 7, // int16
	}

	result := (&Int16{}).Decode(&bts)

	assert.Equal(t, int16(7), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeInt16(t *testing.T) {
	bts := []byte{}
	(&Int16{}).Encode(&bts, int16(7))

	expected := []byte{
		0, 0, 0, 2, // data length
		0, 7, // int16
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeInt32(t *testing.T) {
	bts := []byte{
		0, 0, 0, 4, // data length
		0, 0, 0, 7, // int32
	}

	result := (&Int32{}).Decode(&bts)

	assert.Equal(t, int32(7), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeInt32(t *testing.T) {
	bts := []byte{}
	(&Int32{}).Encode(&bts, int32(7))

	expected := []byte{
		0, 0, 0, 4, // data length
		0, 0, 0, 7, // int32
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeInt64(t *testing.T) {
	bts := []byte{
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 7, // int64
	}

	result := (&Int64{}).Decode(&bts)

	assert.Equal(t, int64(7), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeInt64(t *testing.T) {
	bts := []byte{}
	(&Int64{}).Encode(&bts, int64(27))

	expected := []byte{
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 27, // int64
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeFloat32(t *testing.T) {
	bts := []byte{
		0, 0, 0, 4, // data length
		0xc2, 0, 0, 0,
	}

	result := (&Float32{}).Decode(&bts)

	assert.Equal(t, float32(-32), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeFloat32(t *testing.T) {
	bts := []byte{}
	(&Float32{}).Encode(&bts, float32(-32))

	expected := []byte{
		0, 0, 0, 4, // data length
		0xc2, 0, 0, 0,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeFloat64(t *testing.T) {
	bts := []byte{
		0, 0, 0, 8, // data length
		0xc0, 0x50, 0, 0, 0, 0, 0, 0,
	}

	result := (&Float64{}).Decode(&bts)

	assert.Equal(t, float64(-64), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeFloat64(t *testing.T) {
	bts := []byte{}
	(&Float64{}).Encode(&bts, float64(-64))

	expected := []byte{
		0, 0, 0, 8, // data length
		0xc0, 0x50, 0, 0, 0, 0, 0, 0,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeBool(t *testing.T) {
	bts := []byte{
		0, 0, 0, 1, // data length
		1,
	}

	result := (&Bool{}).Decode(&bts)

	assert.Equal(t, true, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeBool(t *testing.T) {
	bts := []byte{}
	(&Bool{}).Encode(&bts, true)

	expected := []byte{
		0, 0, 0, 1, // data length
		1,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeDateTime(t *testing.T) {
	bts := []byte{
		0, 0, 0, 8, // data length
		0xff, 0xfc, 0xa2, 0xfe, 0xc4, 0xc8, 0x20, 0x0,
	}

	result := (&DateTime{}).Decode(&bts)
	expected := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeDateTime(t *testing.T) {
	bts := []byte{}
	(&DateTime{}).Encode(&bts, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))

	expected := []byte{
		0, 0, 0, 8, // data length
		0xff, 0xfc, 0xa2, 0xfe, 0xc4, 0xc8, 0x20, 0x0,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeDuration(t *testing.T) {
	bts := []byte{
		0, 0, 0, 0x10, // data length
		0, 0, 0, 0, 0, 0xf, 0x42, 0x40,
		0, 0, 0, 0, // reserved
		0, 0, 0, 0, // reserved
	}

	result := (&Duration{}).Decode(&bts)

	assert.Equal(t, time.Duration(1_000_000_000), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeDuration(t *testing.T) {
	bts := []byte{}
	(&Duration{}).Encode(&bts, time.Duration(1_000_000_000))

	expected := []byte{
		0, 0, 0, 0x10, // data length
		0, 0, 0, 0, 0, 0xf, 0x42, 0x40,
		0, 0, 0, 0, // reserved
		0, 0, 0, 0, // reserved
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeJSON(t *testing.T) {
	bts := []byte{
		0, 0, 0, 0x12, // data length
		1, // json format
		0x7b, 0x22, 0x68, 0x65,
		0x6c, 0x6c, 0x6f, 0x22,
		0x3a, 0x22, 0x77, 0x6f,
		0x72, 0x6c, 0x64, 0x22,
		0x7d,
	}

	result := (&JSON{}).Decode(&bts)
	expected := map[string]interface{}{"hello": "world"}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeJSON(t *testing.T) {
	bts := []byte{}
	(&JSON{}).Encode(&bts, map[string]string{"hello": "world"})

	expected := []byte{
		0, 0, 0, 0x12, // data length
		1, // json format
		0x7b, 0x22, 0x68, 0x65,
		0x6c, 0x6c, 0x6f, 0x22,
		0x3a, 0x22, 0x77, 0x6f,
		0x72, 0x6c, 0x64, 0x22,
		0x7d,
	}

	assert.Equal(t, expected, bts)
}
