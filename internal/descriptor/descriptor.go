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

package descriptor

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/geldata/gel-go/internal"
	"github.com/geldata/gel-go/internal/buff"
	"github.com/geldata/gel-go/internal/geltypes"
)

// IDZero is descriptor ID 00000000-0000-0000-0000-000000000000
// https://www.edgedb.com/docs/internals/protocol/typedesc#type-descriptors
var IDZero = geltypes.UUID{}

//go:generate go run golang.org/x/tools/cmd/stringer@v0.25.0 -type Type

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

	// InputShape represents the input shape descriptor type.
	InputShape

	// Range represents the range descriptor type.
	Range

	// ObjectShape represents the object shape descriptor type.
	ObjectShape

	// Compound represents the compound descriptor type.
	Compound

	// MultiRange represents the multi range descriptor type.
	MultiRange

	// SQLRecord represents the SQL record descriptor type.
	SQLRecord
)

// Descriptor is a type descriptor
// https://www.edgedb.com/docs/internals/protocol/typedesc
type Descriptor struct {
	Type   Type
	ID     geltypes.UUID
	Fields []*Field
}

// Field represents the child of a descriptor
type Field struct {
	Name     string
	Desc     Descriptor
	Required bool
}

// Pop builds a descriptor tree from a describe statement type description.
func Pop(
	r *buff.Reader,
	version internal.ProtocolVersion,
) (Descriptor, error) {
	if len(r.Buf) == 0 {
		return Descriptor{Type: Tuple, ID: IDZero}, nil
	}

	descriptors := []Descriptor{}
	for len(r.Buf) > 0 {
		typ := Type(r.PopUint8())
		id := r.PopUUID()
		var desc Descriptor

		switch typ {
		case Set:
			fields := []*Field{{
				Desc: descriptors[r.PopUint16()],
			}}
			desc = Descriptor{Set, id, fields}
		case Object, InputShape:
			fields, err := objectFields(r, descriptors, version)
			if err != nil {
				return Descriptor{}, err
			}
			desc = Descriptor{typ, id, fields}
		case BaseScalar:
			desc = Descriptor{BaseScalar, id, nil}
		case Scalar:
			desc = Descriptor{Scalar, id, []*Field{{
				Desc: descriptors[r.PopUint16()],
			}}}
		case Tuple:
			fields := tupleFields(r, descriptors)
			desc = Descriptor{Tuple, id, fields}
		case NamedTuple:
			fields := namedTupleFields(r, descriptors)
			desc = Descriptor{typ, id, fields}
		case Array:
			fields := []*Field{{
				Desc: descriptors[r.PopUint16()],
			}}
			err := assertArrayDimensions(r)
			if err != nil {
				return Descriptor{}, err
			}
			desc = Descriptor{typ, id, fields}
		case Enum:
			discardEnumMemberNames(r)
			desc = Descriptor{typ, id, nil}
		case Range:
			desc = Descriptor{typ, id, []*Field{{
				Desc: descriptors[r.PopUint16()],
			}}}
		default:

			if 0x80 <= typ {
				// ignore unknown type annotations
				r.PopBytes()
				break
			}

			return Descriptor{}, fmt.Errorf(
				"poping descriptor: unknown descriptor type 0x%x", typ)
		}

		descriptors = append(descriptors, desc)
	}

	return descriptors[len(descriptors)-1], nil
}

func objectFields(
	r *buff.Reader,
	descriptors []Descriptor,
	version internal.ProtocolVersion,
) ([]*Field, error) {
	n := int(r.PopUint16())
	fields := make([]*Field, n)

	for i := 0; i < n; i++ {
		var required bool
		if version.GTE(internal.ProtocolVersion{Major: 0, Minor: 11}) {
			r.Discard(4) // flags
			card := r.PopUint8()
			switch card {
			case 0x6f, 0x6d:
				required = false
			case 0x41, 0x4d:
				required = true
			default:
				return nil, fmt.Errorf("unexpected cardinality: %v", card)
			}
		} else {
			r.Discard(1) // flags

			// Preserve backward compatibility with old behavior. If the
			// protocol version does not support the cardinality flag assume
			// all fields are required.
			required = true
		}

		fields[i] = &Field{
			Name:     r.PopString(),
			Desc:     descriptors[r.PopUint16()],
			Required: required,
		}
	}

	return fields, nil
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

func assertArrayDimensions(r *buff.Reader) error {
	n := int(r.PopUint16()) // number of array dimensions
	if n == 0 {
		return errors.New(
			"too few array dimensions: expected at least 1, got 0")
	}

	r.Discard(4 * n) // array dimension
	return nil
}

func discardEnumMemberNames(r *buff.Reader) {
	n := int(r.PopUint16())
	for i := 0; i < n; i++ {
		r.PopBytes() // enumeration member name
	}
}
