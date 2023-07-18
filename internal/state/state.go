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

package state

import (
	"fmt"

	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/descriptor"
)

type encoderField struct {
	codec codecs.Encoder
	name  string
}

// BuildEncoder builds a state descriptor codec.
func BuildEncoder(
	desc descriptor.Descriptor,
	path codecs.Path,
) (codecs.Encoder, error) {
	switch desc.Type {
	case descriptor.Set:
		if desc.Fields[0].Desc.Type == descriptor.Array {
			return buildSetOfArrayCodec(desc, path)
		}

		// sets are encoded the same as arrays
		fallthrough
	case descriptor.Array:
		return buildArrayEncoder(desc, path)
	case descriptor.Object, descriptor.NamedTuple:
		return buildObjectOrNamedTupleEncoder(desc, path)
	case descriptor.BaseScalar:
		return codecs.BuildScalarEncoder(desc)
	case descriptor.Tuple:
		return buildTupleEncoder(desc, path)
	case descriptor.Enum:
		return &codecs.StrCodec{ID: desc.ID}, nil
	case descriptor.InputShape:
		return buildSparceObjectEncoder(desc, path)
	default:
		return nil, fmt.Errorf(
			"building state codec: unexpected descriptor type 0x%x", desc.Type)
	}
}

// BuildEncoderV2 builds a state descriptor codec.
func BuildEncoderV2(
	desc *descriptor.V2,
	path codecs.Path,
) (codecs.Encoder, error) {
	switch desc.Type {
	case descriptor.Set:
		if desc.Fields[0].Desc.Type == descriptor.Array {
			return buildSetOfArrayCodecV2(desc, path)
		}

		// sets are encoded the same as arrays
		fallthrough
	case descriptor.Array:
		return buildArrayEncoderV2(desc, path)
	case descriptor.Object, descriptor.NamedTuple:
		return buildObjectOrNamedTupleEncoderV2(desc, path)
	case descriptor.Scalar:
		return codecs.BuildScalarEncoderV2(desc)
	case descriptor.Tuple:
		return buildTupleEncoderV2(desc, path)
	case descriptor.Enum:
		return &codecs.StrCodec{ID: desc.ID}, nil
	case descriptor.InputShape:
		return buildSparceObjectEncoderV2(desc, path)
	default:
		return nil, fmt.Errorf(
			"building state codec: unexpected descriptor type 0x%x", desc.Type)
	}
}
