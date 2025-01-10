// This source file is part of the EdgeDB open source project.
//
// Copyright EdgeDB Inc. and the EdgeDB authors.
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

package introspect

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
	Third  []byte `gel:"First"`
}

type InlinedSomeStruct struct {
	SomeStruct `gel:"$inline"`
	Zebra      string
}

type InnerOne struct {
	One         string `gel:"one"`
	OnePointOne string `gel:"one_point_one"`
}

type InnerTwo struct {
	Two string `gel:"two"`
}

type InnerThree struct {
	Three string `gel:"three"`
}

type InlinedMultipleStructs struct {
	InnerOne   `gel:"$inline"`
	InnerTwo   `gel:"$inline"`
	InnerThree `gel:"$inline"`
}

type InlinedMultipleLayers struct {
	InlinedSomeStruct      `gel:"$inline"`
	InlinedMultipleStructs `gel:"$inline"`
}

func checkInlinedSomeStruct(t *testing.T, typ reflect.Type, offset uintptr) {
	// Nested field
	field, ok := StructField(typ, "First")
	require.True(t, ok)
	assert.Equal(t, "Third", field.Name)
	assert.Equal(t, offset+16+8, field.Offset)

	field, ok = StructField(typ, "Second")
	require.True(t, ok)
	assert.Equal(t, "Second", field.Name)
	assert.Equal(t, offset+16, field.Offset)

	// Top level field
	field, ok = StructField(typ, "Zebra")
	require.True(t, ok)
	assert.Equal(t, "Zebra", field.Name)
	assert.Equal(t, offset+16+8+24, field.Offset)
}

func checkInlinedMultipleStructs(
	t *testing.T, typ reflect.Type, offset uintptr,
) {
	field, ok := StructField(typ, "one")
	require.True(t, ok)
	assert.Equal(t, "One", field.Name)
	assert.Equal(t, offset+16*0, field.Offset)

	field, ok = StructField(typ, "one_point_one")
	require.True(t, ok)
	assert.Equal(t, "OnePointOne", field.Name)
	assert.Equal(t, offset+16*1, field.Offset)

	field, ok = StructField(typ, "two")
	require.True(t, ok)
	assert.Equal(t, "Two", field.Name)
	assert.Equal(t, offset+16*2, field.Offset)

	field, ok = StructField(typ, "three")
	require.True(t, ok)
	assert.Equal(t, "Three", field.Name)
	assert.Equal(t, offset+16*3, field.Offset)
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

func TestStructFieldTagInline(t *testing.T) {
	typ := reflect.TypeOf(InlinedSomeStruct{})
	checkInlinedSomeStruct(t, typ, 0)
}

func TestStructFieldTagInlineNestedMultiple(t *testing.T) {
	typ := reflect.TypeOf(InlinedMultipleStructs{})
	checkInlinedMultipleStructs(t, typ, 0)
}

func TestStructFieldTagInlineThreeLayers(t *testing.T) {
	typ := reflect.TypeOf(InlinedMultipleLayers{})
	checkInlinedSomeStruct(t, typ, 0)
	checkInlinedMultipleStructs(t, typ, 16+8+24+16)
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
	require.NoError(t, err)
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
	require.NoError(t, err)
	val.SetBytes([]byte{1, 2, 3})
	assert.Equal(t, []byte{1, 2, 3}, thing)
}

type EdgeDBTag struct {
	FieldName string `edgedb:"tag_name"`
}

func TestEdgeDBTagAccepted(t *testing.T) {
	typ := reflect.TypeOf(EdgeDBTag{})
	field, ok := StructField(typ, "tag_name")
	require.True(t, ok)
	assert.Equal(t, "FieldName", field.Name)
}
