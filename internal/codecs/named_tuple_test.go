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
	"github.com/edgedb/edgedb-go/internal/message"
	"github.com/edgedb/edgedb-go/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamedTupleSetType(t *testing.T) {
	type Thing struct {
		Bool  bool       `edgedb:"bool"`
		Small int16      `edgedb:"small"`
		Med   int32      `edgedb:"med"`
		Large int64      `edgedb:"large"`
		Name  string     `edgedb:"name"`
		ID    types.UUID `edgedb:"id"`
	}

	codec := &NamedTuple{fields: []*objectField{
		{name: "bool", codec: &Bool{typ: boolType}},
		{name: "small", codec: &Int16{typ: int16Type}},
		{name: "med", codec: &Int32{typ: int32Type}},
		{name: "large", codec: &Int64{typ: int64Type}},
		{name: "name", codec: &Str{typ: strType}},
	}}
	useReflect, err := codec.setType(reflect.TypeOf(Thing{}))
	require.Nil(t, err)
	require.False(t, useReflect)

	assert.Equal(t, uintptr(0), codec.fields[0].offset)
	assert.Equal(t, uintptr(2), codec.fields[1].offset)
	assert.Equal(t, uintptr(4), codec.fields[2].offset)
	assert.Equal(t, uintptr(8), codec.fields[3].offset)
	assert.Equal(t, uintptr(16), codec.fields[4].offset)
}

func TestNamedTupleDecodePtr(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 5,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 6,
	})

	type SomeThing struct {
		A int32
		B int32
	}

	var result SomeThing

	codec := &NamedTuple{fields: []*objectField{
		{name: "A", codec: &Int32{typ: int32Type}},
		{name: "B", codec: &Int32{typ: int32Type}},
	}}
	useReflect, err := codec.setType(reflect.TypeOf(result))
	require.Nil(t, err)
	require.False(t, useReflect)
	codec.DecodePtr(r, unsafe.Pointer(&result))

	// force garbage collection to be sure that
	// references are durable.
	debug.FreeOSMemory()

	expected := SomeThing{A: 5, B: 6}
	assert.Equal(t, expected, result)
}

func BenchmarkNamedTupleDecodePtr(b *testing.B) {
	data := []byte{
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 5,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 6,
	}
	r := buff.SimpleReader(repeatingBenchmarkData(b.N, data))

	type SomeThing struct {
		A int32
		B int32
	}

	var result SomeThing
	ptr := unsafe.Pointer(&result)
	codec := &NamedTuple{fields: []*objectField{
		{offset: 0, codec: &Int32{}},
		{offset: 4, codec: &Int32{}},
	}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		codec.DecodePtr(r, ptr)
	}
}

func TestNamedTupleDecodeReflect(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 5,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 6,
	})

	type SomeThing struct {
		A int32
		B int32
	}

	var result SomeThing

	codec := &NamedTuple{fields: []*objectField{
		{name: "A", codec: &Int32{typ: int32Type}},
		{name: "B", codec: &Int32{typ: int32Type}},
	}}
	useReflect, err := codec.setType(reflect.TypeOf(result))
	require.Nil(t, err)
	require.False(t, useReflect)
	codec.DecodeReflect(r, reflect.ValueOf(&result).Elem())

	// force garbage collection to be sure that
	// references are durable.
	debug.FreeOSMemory()

	expected := SomeThing{A: 5, B: 6}
	assert.Equal(t, expected, result)
}

func TestNamedTupleDecodeReflectMap(t *testing.T) {
	r := buff.SimpleReader([]byte{
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 5,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 6,
	})

	codec := &NamedTuple{fields: []*objectField{
		{name: "A", codec: &Int32{typ: int32Type}},
		{name: "B", codec: &Int32{typ: int32Type}},
	}}
	codec.setDefaultType()

	var result map[string]interface{}
	codec.DecodeReflect(r, reflect.ValueOf(&result).Elem())

	// force garbage collection to be sure that
	// references are durable.
	debug.FreeOSMemory()

	expected := map[string]interface{}{"A": int32(5), "B": int32(6)}
	assert.Equal(t, int32(5), result["A"])
	assert.Equal(t, expected, result)
}

func TestEncodeNamedTuple(t *testing.T) {
	codec := &NamedTuple{fields: []*objectField{
		{name: "a", codec: &Int32{}},
		{name: "b", codec: &Int32{}},
	}}

	w := buff.NewWriter([]byte{})
	w.BeginMessage(message.Sync)
	err := codec.Encode(w, []interface{}{map[string]interface{}{
		"a": int32(5),
		"b": int32(6),
	}})
	require.Nil(t, err)
	w.EndMessage()

	conn := &writeFixture{}
	require.Nil(t, w.Send(conn))

	expected := []byte{
		message.Sync,
		0, 0, 0, 0x24,
		0, 0, 0, 28, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 5,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 6,
	}

	assert.Equal(t, expected, conn.written)
}
