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

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	"github.com/edgedb/edgedb-go/internal/geltypes"
)

func buildArrayEncoder(
	desc descriptor.Descriptor,
	path codecs.Path,
) (codecs.Encoder, error) {
	child, err := BuildEncoder(desc.Fields[0].Desc, path)
	if err != nil {
		return nil, err
	}

	return &arrayOrSetEncoder{desc.ID, child}, nil
}

func buildArrayEncoderV2(
	desc *descriptor.V2,
	path codecs.Path,
) (codecs.Encoder, error) {
	child, err := BuildEncoderV2(&desc.Fields[0].Desc, path)
	if err != nil {
		return nil, err
	}

	return &arrayOrSetEncoder{desc.ID, child}, nil
}

type arrayOrSetEncoder struct {
	id    geltypes.UUID
	child codecs.Encoder
}

func (c *arrayOrSetEncoder) DescriptorID() geltypes.UUID { return c.id }

func (c *arrayOrSetEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path codecs.Path,
	_ bool,
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
		err := c.child.Encode(w, in[i], path.AddIndex(i), false)
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
