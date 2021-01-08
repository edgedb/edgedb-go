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

package marshal

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SomeStruct struct {
	First  string
	Second int
	Third  []byte `edgedb:"First"`
}

func TestStructFieldTagPrefered(t *testing.T) {
	typ := reflect.TypeOf(SomeStruct{})
	field, ok := StructField(typ, "First")
	require.True(t, ok)
	assert.Equal(t, "Third", field.Name)
}

func TestStructFieldByName(t *testing.T) {
	typ := reflect.TypeOf(SomeStruct{})
	field, ok := StructField(typ, "Second")
	require.True(t, ok)
	assert.Equal(t, "Second", field.Name)
}

func TestStructFieldMissingField(t *testing.T) {
	typ := reflect.TypeOf(SomeStruct{})
	_, ok := StructField(typ, "Fourth")
	require.False(t, ok)
}

func TestValueOfNonPointer(t *testing.T) {
	var thing string
	_, err := ValueOf(thing)
	expected := errors.New(
		"the \"out\" argument must be a pointer, got string",
	)
	assert.Equal(t, expected, err)
}

func TestValueOfPointerToNil(t *testing.T) {
	thing := (*int64)(nil)
	_, err := ValueOf(thing)
	expected := errors.New(
		"the \"out\" argument must point to a valid value, got <nil>",
	)
	assert.Equal(t, expected, err)
}

func TestValueOfPointer(t *testing.T) {
	var thing string
	val, err := ValueOf(&thing)
	require.Nil(t, err)
	val.SetString("hello")
	assert.Equal(t, "hello", thing)
}

func TestValueOfSliceNonPointer(t *testing.T) {
	var thing []int
	_, err := ValueOfSlice(thing)
	expected := errors.New(
		"the \"out\" argument must be a pointer, got []int",
	)
	assert.Equal(t, expected, err)
}

func TestValueOfSliceNonSlice(t *testing.T) {
	var thing int
	_, err := ValueOfSlice(&thing)
	expected := errors.New(
		"the \"out\" argument must be a pointer to a slice, got *int",
	)
	assert.Equal(t, expected, err)
}

func TestValueOfSlice(t *testing.T) {
	var thing []byte
	val, err := ValueOfSlice(&thing)
	require.Nil(t, err)
	val.SetBytes([]byte{1, 2, 3})
	assert.Equal(t, []byte{1, 2, 3}, thing)
}
