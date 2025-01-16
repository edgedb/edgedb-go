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

package codecs

import (
	"fmt"

	"github.com/geldata/gel-go/internal"
	"github.com/geldata/gel-go/internal/buff"
	"github.com/geldata/gel-go/internal/descriptor"
	types "github.com/geldata/gel-go/internal/geltypes"
)

func buildArgEncoder(
	desc descriptor.Descriptor,
	version internal.ProtocolVersion,
) (Encoder, error) {
	fields := make([]*EncoderField, len(desc.Fields))

	for i, field := range desc.Fields {
		encoder, err := BuildEncoder(field.Desc, version)
		if err != nil {
			return nil, err
		}

		fields[i] = &EncoderField{
			name:     field.Name,
			encoder:  encoder,
			required: field.Required,
		}
	}

	if len(desc.Fields) > 0 && desc.Fields[0].Name != "0" {
		return &kwargsEncoder{desc.ID, fields}, nil
	}

	return &argsEncoder{desc.ID, fields}, nil
}

func buildArgEncoderV2(
	desc *descriptor.V2,
	version internal.ProtocolVersion,
) (Encoder, error) {
	fields := make([]*EncoderField, len(desc.Fields))

	for i, field := range desc.Fields {
		encoder, err := BuildEncoderV2(&field.Desc, version)
		if err != nil {
			return nil, err
		}

		fields[i] = &EncoderField{
			name:     field.Name,
			encoder:  encoder,
			required: field.Required,
		}
	}

	if len(desc.Fields) > 0 && !(desc.Fields[0].Name == "0" ||
		desc.Fields[0].Name == "1") {
		return &kwargsEncoder{desc.ID, fields}, nil
	}

	return &argsEncoder{desc.ID, fields}, nil
}

type argsEncoder struct {
	id     types.UUID
	fields []*EncoderField
}

func (c *argsEncoder) DescriptorID() types.UUID { return c.id }

func (c *argsEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	_ bool,
) error {
	in, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("expected %v to be []interface{} got %T", path, val)
	}

	if len(in) != len(c.fields) {
		return fmt.Errorf(
			"expected %v arguments got %v", len(c.fields), len(in),
		)
	}

	w.BeginBytes()

	elmCount := len(c.fields)
	w.PushUint32(uint32(elmCount))

	var err error
	for i, field := range c.fields {
		w.PushUint32(0) // reserved
		err = field.encoder.Encode(w, in[i], path.AddIndex(i), field.required)
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}

type kwargsEncoder struct {
	id     types.UUID
	fields []*EncoderField
}

func (c *kwargsEncoder) DescriptorID() types.UUID { return c.id }

func (c *kwargsEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	_ bool,
) error {
	args, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("expected %v to be []interface{} got %T", path, val)
	}

	if len(args) != 1 {
		return fmt.Errorf(
			"wrong number of arguments, expected 1 got: %v", len(args),
		)
	}

	in, ok := args[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf(
			"expected %v to be map[string]interface{} got %T", path, args[0],
		)
	}

	elmCount := len(c.fields)
	w.BeginBytes()
	w.PushUint32(uint32(elmCount))

	var err error
	for _, field := range c.fields {
		w.PushUint32(0) // reserved
		err = field.encoder.Encode(
			w,
			in[field.name],
			path.AddField(field.name),
			field.required,
		)

		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
