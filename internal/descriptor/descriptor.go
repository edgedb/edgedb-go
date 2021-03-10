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

package descriptor

import (
	"fmt"
	"strconv"

	"github.com/edgedb/edgedb-go/internal/buff"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

// Type represents a descriptor type.
type Type uint8

const (
	// Set represents the set descriptor type.
	Set Type = iota

	// Object represents the object descriptor type.
	Object

	// BaseScalar represents the base scalar descriptor type.
	BaseScalar

	// Scalar represents the scalar descriptor type.
	Scalar

	// Tuple represents the tuple descriptor type.
	Tuple

	// NamedTuple represents the named tuple descriptor type.
	NamedTuple

	// Array represents the array descriptor type.
	Array

	// Enum represents the enum descriptor type.
	Enum
)

// Descriptor is a type descriptor
// https://www.edgedb.com/docs/internals/protocol/typedesc
type Descriptor struct {
	Type   Type
	ID     types.UUID
	Fields []*Field
}

// Field represents the child of a descriptor
type Field struct {
	Name string
	Desc Descriptor
}

// Pop builds a descriptor tree from a describe statement type description.
func Pop(r *buff.Reader) Descriptor {
	descriptors := []Descriptor{}

	for len(r.Buf) > 0 {
		typ := Type(r.PopUint8())
		id := r.PopUUID()
		var desc Descriptor

		switch typ {
		case Set:
			fields := []*Field{{"", descriptors[r.PopUint16()]}}
			desc = Descriptor{Set, id, fields}
		case Object:
			fields := objectFields(r, descriptors)
			desc = Descriptor{Object, id, fields}
		case BaseScalar:
			desc = Descriptor{BaseScalar, id, nil}
		case Scalar:
			desc = descriptors[r.PopUint16()]
		case Tuple:
			fields := tupleFields(r, descriptors)
			desc = Descriptor{Tuple, id, fields}
		case NamedTuple:
			fields := namedTupleFields(r, descriptors)
			desc = Descriptor{typ, id, fields}
		case Array:
			fields := []*Field{{"", descriptors[r.PopUint16()]}}
			assertArrayDimensions(r)
			desc = Descriptor{typ, id, fields}
		case Enum:
			discardEnumMemberNames(r)
			desc = Descriptor{typ, id, nil}
		default:
			if 0x80 <= typ && typ <= 0xff {
				// ignore unknown type annotations
				r.PopBytes()
				break
			}

			panic(fmt.Sprintf("unknown descriptor type 0x%x", typ))
		}

		descriptors = append(descriptors, desc)
	}

	return descriptors[len(descriptors)-1]
}

func objectFields(r *buff.Reader, descriptors []Descriptor) []*Field {
	n := int(r.PopUint16())
	fields := make([]*Field, n)

	for i := 0; i < n; i++ {
		r.Discard(1) // flags
		fields[i] = &Field{
			Name: r.PopString(),
			Desc: descriptors[r.PopUint16()],
		}
	}

	return fields
}

func tupleFields(r *buff.Reader, descriptors []Descriptor) []*Field {
	n := int(r.PopUint16())
	fields := make([]*Field, n)

	for i := 0; i < n; i++ {
		fields[i] = &Field{
			Name: strconv.Itoa(i),
			Desc: descriptors[r.PopUint16()],
		}
	}

	return fields
}

func namedTupleFields(r *buff.Reader, descriptors []Descriptor) []*Field {
	n := int(r.PopUint16())
	fields := make([]*Field, n)

	for i := 0; i < n; i++ {
		fields[i] = &Field{
			Name: r.PopString(),
			Desc: descriptors[r.PopUint16()],
		}
	}

	return fields
}

func assertArrayDimensions(r *buff.Reader) {
	n := int(r.PopUint16()) // number of array dimensions
	if n == 0 {
		panic("too few array dimensions: expected at least 1, got 0")
	}

	r.Discard(4 * n) // array dimension
}

func discardEnumMemberNames(r *buff.Reader) {
	n := int(r.PopUint16())
	for i := 0; i < n; i++ {
		r.PopBytes() // enumeration member name
	}
}
