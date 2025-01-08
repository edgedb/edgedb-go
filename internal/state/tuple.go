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

func buildTupleEncoder(
	desc descriptor.Descriptor,
	path codecs.Path,
) (codecs.Encoder, error) {
	fields := make([]*encoderField, len(desc.Fields))

	for i, field := range desc.Fields {
		codec, err := BuildEncoder(field.Desc, path.AddField(field.Name))
		if err != nil {
			return nil, err
		}

		fields[i] = &encoderField{codec: codec}
	}

	return &tupleEncoder{desc.ID, fields}, nil
}

func buildTupleEncoderV2(
	desc *descriptor.V2,
	path codecs.Path,
) (codecs.Encoder, error) {
	fields := make([]*encoderField, len(desc.Fields))

	for i, field := range desc.Fields {
		codec, err := BuildEncoderV2(&field.Desc, path.AddField(field.Name))
		if err != nil {
			return nil, err
		}

		fields[i] = &encoderField{codec: codec}
	}

	return &tupleEncoder{desc.ID, fields}, nil
}

type tupleEncoder struct {
	id     geltypes.UUID
	fields []*encoderField
}

func (c *tupleEncoder) DescriptorID() geltypes.UUID { return c.id }

func (c *tupleEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path codecs.Path,
	_ bool,
) error {
	in, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("expected %v to be []interface{} got %T", path, val)
	}

	elmCount := len(c.fields)
	if len(in) != elmCount {
		return fmt.Errorf(
			"expected %v to have %v elements got %v", path, elmCount, len(in))
	}

	w.BeginBytes()
	w.PushUint32(uint32(elmCount))

	for i, field := range c.fields {
		w.PushUint32(0) // reserved
		err := field.codec.Encode(w, in[i], path.AddIndex(i), false)
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
