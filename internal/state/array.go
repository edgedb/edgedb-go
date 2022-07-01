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

func buildArrayCodec(
	desc descriptor.Descriptor,
	path codecs.Path,
) (Codec, error) {
	child, err := BuildCodec(desc.Fields[0].Desc, path)
	if err != nil {
		return nil, err
	}

	return &arrayOrSetCodec{desc.ID, child}, nil
}

type arrayOrSetCodec struct {
	id    edgedbtypes.UUID
	child Codec
}

func (c *arrayOrSetCodec) DescriptorID() edgedbtypes.UUID { return c.id }

func (c *arrayOrSetCodec) Decode(
	r *buff.Reader,
	path codecs.Path,
) (interface{}, error) {
	if r.PopUint32() == 0 { // number of dimensions is 1 or 0
		r.Discard(8) // skip 2 reserved fields
		return nil, nil
	}

	r.Discard(8) // skip 2 reserved fields
	upper := int32(r.PopUint32())
	lower := int32(r.PopUint32())
	elmCount := int(upper - lower + 1)
	result := make([]interface{}, elmCount)

	for i := 0; i < elmCount; i++ {
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

func (c *arrayOrSetCodec) Encode(
	w *buff.Writer,
	path codecs.Path,
	val interface{},
) error {
	in, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("expected %v to be a slice got: %T", path, val)
	}

	elmCount := len(in)

	w.BeginBytes()
	w.PushUint32(1)                // number of dimensions
	w.PushUint32(0)                // reserved
	w.PushUint32(0)                // reserved
	w.PushUint32(uint32(elmCount)) // dimension.upper
	w.PushUint32(1)                // dimension.lower

	for i := 0; i < elmCount; i++ {
		err := c.child.Encode(w, path.AddIndex(i), in[i])
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
