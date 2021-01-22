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
	"github.com/edgedb/edgedb-go/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTupleSetType(t *testing.T) {
	codec := &Tuple{fields: []Codec{
		&Int64{typ: int64Type},
		&Int32{typ: int32Type},
	}}
	useReflect, err := codec.setType(reflect.TypeOf([]interface{}{}))
	require.Nil(t, err)
	require.False(t, useReflect)
}

func TestTupleDecodePtr(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 32, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 2,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4, // data length
		0, 0, 0, 3,
	})

	var result []interface{}

	codec := &Tuple{fields: []Codec{
		&Int64{typ: int64Type},
		&Int32{typ: int32Type},
	}}
	useReflect, err := codec.setType(reflect.TypeOf(result))
	require.Nil(t, err)
	require.False(t, useReflect)
	codec.DecodePtr(r, unsafe.Pointer(&result))

	// force garbage collection to be sure that
	// references are durable.
	debug.FreeOSMemory()

	expected := []interface{}{int64(2), int32(3)}
	assert.Equal(t, expected, result)
}

func TestTupleDecodeReflect(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 32, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 2,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4, // data length
		0, 0, 0, 3,
	})

	var result []interface{}

	codec := &Tuple{fields: []Codec{
		&Int64{typ: int64Type},
		&Int32{typ: int32Type},
	}}
	useReflect, err := codec.setType(reflect.TypeOf(result))
	require.Nil(t, err)
	require.False(t, useReflect)
	codec.DecodeReflect(r, reflect.ValueOf(&result).Elem())

	// force garbage collection to be sure that
	// references are durable.
	debug.FreeOSMemory()

	expected := []interface{}{int64(2), int32(3)}
	assert.Equal(t, expected, result)
}

func TestEncodeNullTuple(t *testing.T) {
	w := buff.NewWriter([]byte{})
	w.BeginMessage(0xff)
	err := (&Tuple{}).Encode(w, []interface{}{})
	require.Nil(t, err)
	w.EndMessage()

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0xff,
		0, 0, 0, 12,
		0, 0, 0, 4, // data length
		0, 0, 0, 0, // number of elements
	}

	assert.Equal(t, expected, conn.written)
}

func TestTupleEncodeWrongNumberOfArgs(t *testing.T) {
	w := buff.NewWriter([]byte{})
	w.BeginMessage(0xff)

	codec := &Tuple{fields: []Codec{
		&Int64{},
		&Int64{},
	}}

	err := codec.Encode(w, []interface{}{int64(2), int64(3), int64(4)})
	assert.EqualError(t, err, "expected 2 elements in the tuple, got 3")
}

func TestEncodeTuple(t *testing.T) {
	w := buff.NewWriter([]byte{})
	w.BeginMessage(0xff)

	codec := &Tuple{fields: []Codec{&Int64{}, &Int64{}}}
	err := codec.Encode(w, []interface{}{int64(2), int64(3)})
	require.Nil(t, err)
	w.EndMessage()

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		0xff,
		0, 0, 0, 44,
		0, 0, 0, 36, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 2,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 3,
	}

	assert.Equal(t, expected, conn.written)
}

func BenchmarkEncodeTuple(b *testing.B) {
	codec := Tuple{fields: []Codec{&UUID{}}}
	id := types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	ids := []interface{}{id}
	w := buff.NewWriter([]byte{})
	w.BeginMessage(0xff)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = codec.Encode(w, ids)
	}
}
