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
	"testing"

	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetNamedTupleType(t *testing.T) {
	codec := &NamedTuple{fields: []*objectField{
		{name: "id", codec: &UUID{t: uuidType}},
		{name: "name", codec: &Str{t: strType}},
		{name: "count", codec: &Int64{t: int64Type}},
	}}

	type Thing struct {
		ID    types.UUID `edgedb:"id"`
		Name  string     `edgedb:"name"`
		Count int64      `edgedb:"count"`
	}

	err := codec.setType(reflect.TypeOf(Thing{}))
	require.Nil(t, err)

	assert.Equal(t, []int{0}, codec.fields[0].index)
	assert.Equal(t, []int{1}, codec.fields[1].index)
	assert.Equal(t, []int{2}, codec.fields[2].index)
}

func TestDecodeNamedTuple(t *testing.T) {
	msg := buff.NewMessage([]byte{
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

	codec := &NamedTuple{fields: []*objectField{
		{index: []int{0}, codec: &Int32{}},
		{index: []int{1}, codec: &Int32{}},
	}}

	type SomeThing struct {
		A int32
		B int32
	}

	var result SomeThing
	val := reflect.ValueOf(&result).Elem()
	codec.Decode(msg, val)

	expected := SomeThing{A: 5, B: 6}

	assert.Equal(t, expected, result)
}

func BenchmarkDecodeNamedTuple(b *testing.B) {
	data := []byte{
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

	type SomeThing struct {
		A int32
		B int32
	}

	var result SomeThing
	val := reflect.ValueOf(&result).Elem()
	codec := &NamedTuple{fields: []*objectField{
		{index: []int{0}, codec: &Int32{}},
		{index: []int{1}, codec: &Int32{}},
	}}

	var msg *buff.Message
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg = buff.NewMessage(data)
		codec.Decode(msg, val)
	}
}

func TestEncodeNamedTuple(t *testing.T) {
	codec := &NamedTuple{fields: []*objectField{
		{name: "a", codec: &Int32{}},
		{name: "b", codec: &Int32{}},
	}}

	bts := buff.NewWriter(nil)
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
