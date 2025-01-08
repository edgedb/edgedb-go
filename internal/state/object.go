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

func buildObjectOrNamedTupleEncoder(
	desc descriptor.Descriptor,
	path codecs.Path,
) (codecs.Encoder, error) {
	fields := make([]*encoderField, len(desc.Fields))
	for i, field := range desc.Fields {
		child, err := BuildEncoder(field.Desc, path.AddField(field.Name))
		if err != nil {
			return nil, err
		}

		fields[i] = &encoderField{
			name:  field.Name,
			codec: child,
		}
	}

	return &objectEncoder{desc.ID, fields}, nil
}

func buildObjectOrNamedTupleEncoderV2(
	desc *descriptor.V2,
	path codecs.Path,
) (codecs.Encoder, error) {
	fields := make([]*encoderField, len(desc.Fields))
	for i, field := range desc.Fields {
		child, err := BuildEncoderV2(&field.Desc, path.AddField(field.Name))
		if err != nil {
			return nil, err
		}

		fields[i] = &encoderField{
			name:  field.Name,
			codec: child,
		}
	}

	return &objectEncoder{desc.ID, fields}, nil
}

type objectEncoder struct {
	id     geltypes.UUID
	fields []*encoderField
}

func (c *objectEncoder) DescriptorID() geltypes.UUID { return c.id }

func (c *objectEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path codecs.Path,
	_ bool,
) error {
	in, ok := val.(map[string]interface{})
	if !ok {
		return fmt.Errorf(
			"expected %v to be map[string]interface{} got %T", path, val)
	} else if len(in) != len(c.fields) {
		return fmt.Errorf(
			"expected %v to have %v fields got %v",
			path, len(c.fields), len(in))
	}

	w.BeginBytes()
	w.PushUint32(uint32(len(c.fields)))
	for _, field := range c.fields {
		fieldValue, ok := in[field.name]
		if !ok {
			return fmt.Errorf(
				"expected %v to have a field %q but it is missing",
				path, field.name)
		}

		w.PushUint32(0)
		err := field.codec.Encode(
			w,
			fieldValue,
			path.AddField(field.name),
			false,
		)
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
