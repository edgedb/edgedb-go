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
	"fmt"
	"strconv"

	"github.com/geldata/gel-go/internal"
	"github.com/geldata/gel-go/internal/buff"
	"github.com/geldata/gel-go/internal/geltypes"
)

// V2 is a type descriptor
// https://www.edgedb.com/docs/internals/protocol/typedesc
type V2 struct {
	Type          Type
	ID            geltypes.UUID
	Name          string
	SchemaDefined bool
	Ancestors     []*FieldV2
	Fields        []*FieldV2
}

// FieldV2 represents the child of a descriptor
type FieldV2 struct {
	Name     string
	Desc     V2
	Required bool
	Union    bool
}

// PopV2 builds a descriptor tree from a describe statement type description.
func PopV2(
	r *buff.Reader,
	_ internal.ProtocolVersion,
) (V2, error) {
	if len(r.Buf) == 0 {
		return V2{Type: Tuple, ID: IDZero}, nil
	}
	descriptorsV2 := []V2{}
	for len(r.Buf) > 0 {
		r.PopUint32()
		t := r.PopUint8()
		typ := Type(t)
		id := r.PopUUID()
		var desc V2

		switch typ {
		case Set:
			fields := []*FieldV2{{
				Desc: descriptorsV2[r.PopUint16()],
			}}
			desc = V2{Set, id, "", false, nil, fields}
		case Object:
			r.PopUint8()  // schema_defined
			r.PopUint16() // type
			fields, err := objectFields2pX(r, descriptorsV2, false)
			if err != nil {
				return V2{}, err
			}
			desc = V2{Object, id, "", true, nil, fields}
		case Scalar:
			name := r.PopString()
			r.PopUint8() // schema_defined
			ancestors := scalarFields2pX(r, descriptorsV2, false)
			desc = V2{Scalar, id, name, true, ancestors, nil}
		case Tuple:
			name := r.PopString()
			r.PopUint8() // schema_defined
			ancestors, fields := tupleFields2pX(r, descriptorsV2)
			desc = V2{Tuple, id, name, true, ancestors, fields}
		case NamedTuple:
			name := r.PopString()
			r.PopUint8() // schema_defined
			ancestors, fields := namedTupleFields2pX(r, descriptorsV2)
			desc = V2{Tuple, id, name, true, ancestors, fields}
		case Array:
			name := r.PopString()
			r.PopUint8() // schema_defined
			ancestors := scalarFields2pX(r, descriptorsV2, false)
			fields := []*FieldV2{{
				Desc: descriptorsV2[r.PopUint16()],
			}}
			err := assertArrayDimensions(r)
			if err != nil {
				return V2{}, err
			}
			desc = V2{Array, id, name, true, ancestors, fields}
		case Enum:
			name := r.PopString()
			r.PopUint8() // schema_defined
			ancestors := scalarFields2pX(r, descriptorsV2, false)
			discardEnumMemberNames(r)
			desc = V2{Enum, id, name, true, ancestors, nil}
		case InputShape:
			fields, err := objectFields2pX(r, descriptorsV2, true)
			if err != nil {
				return V2{}, err
			}
			desc = V2{InputShape, id, "", true, nil, fields}
		case Range:
			name := r.PopString()
			r.PopUint8() // schema_defined
			ancestors := scalarFields2pX(r, descriptorsV2, false)
			fields := []*FieldV2{{
				Desc: descriptorsV2[r.PopUint16()],
			}}
			desc = V2{Range, id, name, true, ancestors, fields}
		case ObjectShape:
			name := r.PopString()
			r.PopUint8() // schema_defined
			desc = V2{ObjectShape, id, name, true, nil, nil}
		case Compound:
			name := r.PopString()
			r.PopUint8() // schema_defined
			t := r.PopUint8()
			var unionOperation bool
			switch t {
			case 0x01:
				unionOperation = true
			case 0x02:
				unionOperation = false
			default:
				return V2{}, fmt.Errorf("unexpected operation type: %v", t)
			}
			fields := scalarFields2pX(r, descriptorsV2, unionOperation)
			desc = V2{Compound, id, name, true, nil, fields}
		case MultiRange:
			name := r.PopString()
			r.PopUint8() // schema_defined
			ancestors := scalarFields2pX(r, descriptorsV2, false)
			fields := []*FieldV2{{
				Desc: V2{
					Type: Range,
					Fields: []*FieldV2{{
						Desc: descriptorsV2[r.PopUint16()],
					}},
				},
			}}
			desc = V2{MultiRange, id, name, true, ancestors, fields}
		case SQLRecord:
			fields := sqlRecordFields(r, descriptorsV2)
			desc = V2{SQLRecord, id, "", false, nil, fields}
		default:
			if 0x80 <= typ {
				// ignore unknown type annotations
				r.PopBytes()
				break
			}
			return V2{}, fmt.Errorf(
				"poping descriptor: unknown descriptor type 0x%x", typ)
		}

		descriptorsV2 = append(descriptorsV2, desc)
	}

	return descriptorsV2[len(descriptorsV2)-1], nil
}

func scalarFields2pX(r *buff.Reader, descriptors []V2, union bool) []*FieldV2 {
	n := int(r.PopUint16())
	fields := make([]*FieldV2, n)
	for i := 0; i < n; i++ {
		fields[i] = &FieldV2{
			Desc:  descriptors[r.PopUint16()],
			Union: union,
		}
	}

	return fields
}

func tupleFields2pX(r *buff.Reader, descriptors []V2) (
	[]*FieldV2, []*FieldV2,
) {
	n := int(r.PopUint16())
	ancestors := make([]*FieldV2, n)

	for i := 0; i < n; i++ {
		ancestors[i] = &FieldV2{
			Name: strconv.Itoa(i),
			Desc: descriptors[r.PopUint16()],
		}
	}

	n = int(r.PopUint16())
	fields := make([]*FieldV2, n)

	for i := 0; i < n; i++ {
		fields[i] = &FieldV2{
			Name: strconv.Itoa(i),
			Desc: descriptors[r.PopUint16()],
		}
	}

	return ancestors, fields
}

func namedTupleFields2pX(r *buff.Reader, descriptors []V2) (
	[]*FieldV2, []*FieldV2,
) {
	n := int(r.PopUint16())
	ancestors := make([]*FieldV2, n)

	for i := 0; i < n; i++ {
		ancestors[i] = &FieldV2{
			Name: strconv.Itoa(i),
			Desc: descriptors[r.PopUint16()],
		}
	}

	n = int(r.PopUint16())
	fields := make([]*FieldV2, n)

	for i := 0; i < n; i++ {
		fields[i] = &FieldV2{
			Name: r.PopString(),
			Desc: descriptors[r.PopUint16()],
		}
	}

	return ancestors, fields
}

func objectFields2pX(r *buff.Reader, descriptors []V2, input bool) (
	[]*FieldV2, error,
) {
	n := int(r.PopUint16())
	fields := make([]*FieldV2, n)

	for i := 0; i < n; i++ {
		var required bool
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
		fields[i] = &FieldV2{
			Name:     r.PopString(),
			Desc:     descriptors[r.PopUint16()],
			Required: required,
		}
		if !input {
			r.PopUint16() // source_type
		}
	}

	return fields, nil
}

func sqlRecordFields(
	r *buff.Reader,
	descriptors []V2,
) []*FieldV2 {
	n := int(r.PopUint16())
	fields := make([]*FieldV2, n)

	for i := 0; i < n; i++ {
		fields[i] = &FieldV2{
			Name:     r.PopString(),
			Desc:     descriptors[r.PopUint16()],
			Required: true,
		}
	}

	return fields
}
