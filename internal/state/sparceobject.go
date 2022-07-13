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

func buildSparceObjectCodec(
	desc descriptor.Descriptor,
	path codecs.Path,
) (Codec, error) {
	fields := make([]*codecField, len(desc.Fields))
	for i, field := range desc.Fields {
		child, err := BuildCodec(field.Desc, path.AddField(field.Name))
		if err != nil {
			return nil, err
		}

		fields[i] = &codecField{
			name:  field.Name,
			codec: child,
		}
	}

	return &sparceObjectCodec{desc.ID, fields}, nil
}

type sparceObjectCodec struct {
	id     edgedbtypes.UUID
	fields []*codecField
}

func (c *sparceObjectCodec) DescriptorID() edgedbtypes.UUID { return c.id }

func (c *sparceObjectCodec) Decode(
	r *buff.Reader,
	path codecs.Path,
) (interface{}, error) {
	elmCount := int(r.PopUint32())
	if elmCount > len(c.fields) {
		return nil, fmt.Errorf(
			"too many object fields: expected at most %v, got %v",
			len(c.fields), elmCount)
	}

	result := make(map[string]interface{}, elmCount)
	for i := 0; i < elmCount; i++ {
		field := c.fields[r.PopUint32()]
		val, err := field.codec.Decode(r.PopSlice(r.PopUint32()), path)
		if err != nil {
			return nil, err
		}

		result[field.name] = val
	}

	return result, nil
}

func (c *sparceObjectCodec) Encode(
	w *buff.Writer,
	path codecs.Path,
	val interface{},
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
		if !ok {
			continue
		}

		w.PushUint32(uint32(i))
		err = field.codec.Encode(w, path.AddField(field.name), fieldValue)
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
