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
	"reflect"
	"runtime/debug"
	"testing"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArraySetType(t *testing.T) {
	codec := &Array{child: &Int64{typ: int64Type}}
	err := codec.setType(reflect.TypeOf([]int64{}))
	require.Nil(t, err)

	assert.Equal(t, 8, codec.step)
}

func TestDecodeArray(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 56, // data length
		0, 0, 0, 1, // dimension count
		0, 0, 0, 0, // reserved
		0, 0, 0, 0x14, // reserved
		0, 0, 0, 3, // dimension.upper
		0, 0, 0, 1, // dimension.lower
		// element 0
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 3, // ing64
		// element 1
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 5, // int64
		// element 2
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 8, // int64
	})

	var result []int64

	codec := &Array{child: &Int64{typ: int64Type}}
	err := codec.setType(reflect.TypeOf(result))
	require.Nil(t, err)
	codec.Decode(r, unsafe.Pointer(&result))

	// force garbage collection to be sure that
	// references are durable.
	debug.FreeOSMemory()

	expected := []int64{3, 5, 8}
	assert.Equal(t, expected, result)
}

func TestEncodeArray(t *testing.T) {
	arrays := []interface{}{
		[]interface{}{int64(3), int64(5), int64(8)},
		[]int64{3, 5, 8},
	}

	for _, a := range arrays {
		w := buff.NewWriter()
		w.BeginMessage(0xff)

		codec := &Array{child: &Int64{}}
		err := codec.Encode(w, a)
		require.Nil(t, err)
		w.EndMessage()

		expected := []byte{
			0xff,          // message type
			0, 0, 0, 0x40, // message length
			0, 0, 0, 0x38, // data length
			0, 0, 0, 1, // dimension count
			0, 0, 0, 0, // reserved
			0, 0, 0, 0, // reserved
			0, 0, 0, 3, // dimension.upper
			0, 0, 0, 1, // dimension.lower
			// element 0
			0, 0, 0, 8, // data length
			0, 0, 0, 0, 0, 0, 0, 3, // ing64
			// element 1
			0, 0, 0, 8, // data length
			0, 0, 0, 0, 0, 0, 0, 5, // int64
			// element 2
			0, 0, 0, 8, // data length
			0, 0, 0, 0, 0, 0, 0, 8, // int64
		}

		conn := &writeFixture{}
		require.Nil(t, w.Send(conn))
		assert.Equal(t, expected, conn.written)
	}
}

func TestEncodeArrayWrongType(t *testing.T) {
	w := buff.NewWriter()
	w.BeginMessage(0xff)

	codec := &Array{child: &Int64{typ: reflect.TypeOf(int64(1))}}
	err := codec.Encode(w, "hello")
	assert.EqualError(t, err, "expected []int64 got: string")
}
