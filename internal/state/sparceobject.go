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
	"strings"

	"github.com/geldata/gel-go/internal/buff"
	"github.com/geldata/gel-go/internal/codecs"
	"github.com/geldata/gel-go/internal/descriptor"
	"github.com/geldata/gel-go/internal/geltypes"
)

func buildSparceObjectEncoder(
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

	return &sparceObjectEncoder{desc.ID, fields}, nil
}

func buildSparceObjectEncoderV2(
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

	return &sparceObjectEncoder{desc.ID, fields}, nil
}

type sparceObjectEncoder struct {
	id     geltypes.UUID
	fields []*encoderField
}

func (c *sparceObjectEncoder) DescriptorID() geltypes.UUID { return c.id }

func (c *sparceObjectEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path codecs.Path,
	_ bool,
) error {
	in, ok := val.(map[string]interface{})
	if !ok {
		return fmt.Errorf(
			"expected %v to be map[string]interface{} got %T", path, val)
	}

	elmCount := len(in)
	w.BeginBytes()
	w.PushUint32(uint32(elmCount))

	var err error
	seen := 0
	for i, field := range c.fields {
		fieldValue, ok := in[field.name]
		if !ok && strings.HasPrefix(field.name, "default::") {
			fieldValue, ok = in[strings.TrimPrefix(field.name, "default::")]
		}
		if !ok {
			continue
		}

		w.PushUint32(uint32(i))
		err = field.codec.Encode(
			w,
			fieldValue,
			path.AddField(field.name),
			false,
		)
		if err != nil {
			return err
		}
		seen++
	}

	if seen != elmCount {
		missing := make(map[string]struct{}, elmCount)
		for name := range in {
			missing[name] = struct{}{}
		}

		for _, field := range c.fields {
			if _, ok := in[field.name]; ok {
				delete(missing, field.name)
			}
		}

		for name := range missing {
			return fmt.Errorf(
				"found unknown state value %v",
				path.AddField(name),
			)
		}
	}
	w.EndBytes()
	return nil
}
