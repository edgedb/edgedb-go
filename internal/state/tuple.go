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

func buildTupleCodec(
	desc descriptor.Descriptor,
	path codecs.Path,
) (Codec, error) {
	fields := make([]*codecField, len(desc.Fields))

	for i, field := range desc.Fields {
		codec, err := BuildCodec(field.Desc, path.AddField(field.Name))
		if err != nil {
			return nil, err
		}

		fields[i] = &codecField{codec: codec}
	}

	return &tupleCodec{desc.ID, fields}, nil
}

type tupleCodec struct {
	id     edgedbtypes.UUID
	fields []*codecField
}

func (c *tupleCodec) DescriptorID() edgedbtypes.UUID { return c.id }

func (c *tupleCodec) Decode(
	r *buff.Reader,
	path codecs.Path,
) (interface{}, error) {
	elmCount := int(int32(r.PopUint32()))
	if elmCount != len(c.fields) {
		return nil, fmt.Errorf(
			"wrong number of elements for %v, expected %v got %v",
			path, len(c.fields), elmCount)
	}

	result := make([]interface{}, elmCount)
	for i, field := range c.fields {
		r.Discard(4) // reserved
		val, err := field.codec.Decode(
			r.PopSlice(r.PopUint32()),
			path.AddField(field.name),
		)
		if err != nil {
			return nil, err
		}

		result[i] = val
	}

	return result, nil
}

func (c *tupleCodec) Encode(
	w *buff.Writer,
	path codecs.Path,
	val interface{},
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
		err := field.codec.Encode(w, path.AddIndex(i), in[i])
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
