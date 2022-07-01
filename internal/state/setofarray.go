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
	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/descriptor"
)

func buildSetOfArrayCodec(
	desc descriptor.Descriptor,
	path codecs.Path,
) (Codec, error) {
	child, err := BuildCodec(desc.Fields[0].Desc, path)
	if err != nil {
		return nil, err
	}

	return &setOfArrayCodec{arrayOrSetCodec{desc.ID, child}}, nil
}

type setOfArrayCodec struct {
	arrayOrSetCodec
}

func (c *setOfArrayCodec) Decode(
	r *buff.Reader,
	path codecs.Path,
) (interface{}, error) {
	if r.PopUint32() == 0 {
		r.Discard(8) // skip 2 reserved fields
		return nil, nil
	}

	r.Discard(8) // skip 2 reserved fields
	upper := int32(r.PopUint32())
	lower := int32(r.PopUint32())
	n := int(upper - lower + 1)
	result := make([]interface{}, n)

	for i := 0; i < n; i++ {
		r.Discard(12)
		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

		val, err := c.child.Decode(r.PopSlice(elmLen), path.AddIndex(i))
		if err != nil {
			return nil, err
		}

		result[i] = val
	}

	return result, nil
}
