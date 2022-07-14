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

package state

import (
	"fmt"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	"github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

// Codec is a state descriptor codec.
type Codec interface {
	Encode(*buff.Writer, codecs.Path, interface{}) error
	Decode(*buff.Reader, codecs.Path) (interface{}, error)
	DescriptorID() edgedbtypes.UUID
}

type codecField struct {
	codec Codec
	name  string
}

// BuildCodec builds a state descriptor codec.
func BuildCodec(desc descriptor.Descriptor, path codecs.Path) (Codec, error) {
	switch desc.Type {
	case descriptor.Set:
		if desc.Fields[0].Desc.Type == descriptor.Array {
			return buildSetOfArrayCodec(desc, path)
		}

		// sets are encoded the same as arrays
		fallthrough
	case descriptor.Array:
		return buildArrayCodec(desc, path)
	case descriptor.Object, descriptor.NamedTuple:
		return buildObjectOrNamedTupleCodec(desc, path)
	case descriptor.BaseScalar:
		return buildBaseScalarCodec(desc, path)
	case descriptor.Tuple:
		return buildTupleCodec(desc, path)
	case descriptor.Enum:
		return &strCodec{desc.ID}, nil
	case descriptor.InputShape:
		return buildSparceObjectCodec(desc, path)
	default:
		return nil, fmt.Errorf(
			"building state codec: unexpected descriptor type 0x%x", desc.Type)
	}
}

func buildBaseScalarCodec(
	desc descriptor.Descriptor,
	path codecs.Path,
) (Codec, error) {
	switch desc.ID {
	case codecs.StrID:
		return &strCodec{desc.ID}, nil
	case codecs.Int64ID:
		return &int64Codec{}, nil
	case codecs.BoolID:
		return &boolCodec{}, nil
	case codecs.DurationID:
		return &durationCodec{}, nil
	case codecs.MemoryID:
		return &memoryCodec{}, nil
	default:
		return nil, fmt.Errorf(
			"building state codec: unexpected scalar type ID: %v", desc.ID)
	}
}
