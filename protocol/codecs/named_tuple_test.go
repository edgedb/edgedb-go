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

	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/types"
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
		{name: "id", codec: &UUID{typ: uuidType}},
	}}
	err := codec.setType(reflect.TypeOf(Thing{}))
	require.Nil(t, err)

	assert.Equal(t, uintptr(0), codec.fields[0].offset)
	assert.Equal(t, uintptr(2), codec.fields[1].offset)
	assert.Equal(t, uintptr(4), codec.fields[2].offset)
	assert.Equal(t, uintptr(8), codec.fields[3].offset)
	assert.Equal(t, uintptr(16), codec.fields[4].offset)
	assert.Equal(t, uintptr(32), codec.fields[5].offset)
}

func TestDecodeNamedTuple(t *testing.T) {
	buf := buff.New([]byte{
		0,
		0, 0, 0, 36,
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
	})
	buf.Next()

	type SomeThing struct {
		A int32
		B int32
	}

	var result SomeThing

	codec := &NamedTuple{fields: []*objectField{
		{name: "A", codec: &Int32{typ: int32Type}},
		{name: "B", codec: &Int32{typ: int32Type}},
	}}
	err := codec.setType(reflect.TypeOf(result))
	require.Nil(t, err)
	codec.Decode(buf, unsafe.Pointer(&result))

	// force garbage collection to be sure that
	// references are durable.
	debug.FreeOSMemory()

	expected := SomeThing{A: 5, B: 6}
	assert.Equal(t, expected, result)
}

func BenchmarkDecodeNamedTuple(b *testing.B) {
	data := []byte{
		0xa,
		0, 0, 0, 32,
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
	buf := buff.New(data)
	buf.Next()

	type SomeThing struct {
		A int32
		B int32
	}

	var result SomeThing
	ptr := unsafe.Pointer(&result)
	codec := &NamedTuple{fields: []*objectField{
		{offset: 0, codec: &Int32{}},
		// todo fix offsets
		{offset: 0, codec: &Int32{}},
	}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Msg = data[5:]
		codec.Decode(buf, ptr)
	}
}

func TestEncodeNamedTuple(t *testing.T) {
	codec := &NamedTuple{fields: []*objectField{
		{name: "a", codec: &Int32{}},
		{name: "b", codec: &Int32{}},
	}}

	bts := buff.New(nil)
	bts.BeginMessage(0xff)
	codec.Encode(bts, []interface{}{map[string]interface{}{
		"a": int32(5),
		"b": int32(6),
	}})
	bts.EndMessage()

	expected := []byte{
		0xff,          // message type
		0, 0, 0, 0x24, // message length
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

	assert.Equal(t, expected, *bts.Unwrap())
}
