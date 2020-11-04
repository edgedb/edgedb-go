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

	"github.com/edgedb/edgedb-go/edgedb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetObjectType(t *testing.T) {
	codec := &Object{fields: []*objectField{
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

func TestDecodeObject(t *testing.T) {
	codec := &Object{fields: []*objectField{
		{index: []int{0}, codec: &Str{}},
		{index: []int{1}, codec: &Int32{}},
		{index: []int{2}, codec: &Int64{}},
	}}

	bts := []byte{
		0, 0, 0, 36, // data length
		0, 0, 0, 2, // element count
		// field 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4, // data length
		102, 111, 117, 114, // utf-8 data
		// field 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4, // data length
		0, 0, 0, 4, // int32
		// field 2
		0, 0, 0, 0, // reserved
		0xff, 0xff, 0xff, 0xff, // data length (-1)
	}

	type SomeThing struct {
		A string
		B int32
		C int64
	}

	var result SomeThing
	val := reflect.ValueOf(&result).Elem()
	codec.Decode(&bts, val)

	expected := SomeThing{A: "four", B: 4, C: 0}
	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func BenchmarkDecodeObject(b *testing.B) {
	data := []byte{
		0, 0, 0, 36, // data length
		0, 0, 0, 2, // element count
		// field 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4, // data length
		102, 111, 117, 114, // utf-8 data
		// field 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4, // data length
		0, 0, 0, 4, // int32
		// field 2
		0, 0, 0, 0, // reserved
		0xff, 0xff, 0xff, 0xff, // data length (-1)
	}

	type SomeThing struct {
		A string
		B int32
		C int64
	}

	var result SomeThing
	val := reflect.ValueOf(&result).Elem()
	codec := &Object{fields: []*objectField{
		{index: []int{0}, codec: &Str{}},
		{index: []int{1}, codec: &Int32{}},
		{index: []int{2}, codec: &Int64{}},
	}}

	var buf []byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = data
		codec.Decode(&buf, val) // nolint
	}
}
