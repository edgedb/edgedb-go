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

package codecs

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
	"github.com/edgedb/edgedb-go/internal/marshal"
)

func buildTupleEncoder(desc descriptor.Descriptor) (Encoder, error) {
	fields := make([]*EncoderField, len(desc.Fields))

	for i, field := range desc.Fields {
		encoder, err := BuildEncoder(field.Desc)
		if err != nil {
			return nil, err
		}

		fields[i] = &EncoderField{
			encoder: encoder,
		}
	}

	return &tupleEncoder{desc.ID, fields}, nil
}

type tupleEncoder struct {
	id     types.UUID
	fields []*EncoderField
}

func (c *tupleEncoder) DescriptorID() types.UUID { return c.id }

func (c *tupleEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	in, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("expected %v to be []interface{} got %T", path, val)
	}

	if len(in) != len(c.fields) {
		return fmt.Errorf(
			"expected %v to be []interface{} with len=%v, got len=%v",
			path, len(c.fields), len(in),
		)
	}

	w.BeginBytes()

	elmCount := len(c.fields)
	w.PushUint32(uint32(elmCount))

	var err error
	for i, field := range c.fields {
		w.PushUint32(0) // reserved
		err = field.encoder.Encode(w, in[i], path.AddIndex(i))
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}

func buildTupleDecoder(
	desc descriptor.Descriptor,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf(
			"expected %v to be a struct got %v", path, typ.Kind(),
		)
	}

	fields := make([]*DecoderField, len(desc.Fields))

	for i, field := range desc.Fields {
		sf, ok := marshal.StructField(typ, field.Name)
		if !ok {
			return nil, fmt.Errorf(
				"expected %v to have a field with the tag `edgedb:\"%v\"`",
				typ, field.Name,
			)
		}

		child, err := BuildDecoder(
			field.Desc,
			sf.Type,
			path.AddField(field.Name),
		)

		if err != nil {
			return nil, err
		}

		fields[i] = &DecoderField{
			name:    field.Name,
			offset:  sf.Offset,
			decoder: child,
		}
	}

	return &tupleDecoder{desc.ID, fields}, nil
}

type tupleDecoder struct {
	id     types.UUID
	fields []*DecoderField
}

func (c *tupleDecoder) DescriptorID() types.UUID { return c.id }

func (c *tupleDecoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	elmCount := int(int32(r.PopUint32()))
	if elmCount != len(c.fields) {
		panic(fmt.Sprintf(
			"wrong number of elements, expected %v got %v",
			len(c.fields), elmCount,
		))
	}

	for _, field := range c.fields {
		r.Discard(4) // reserved

		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

		field.decoder.Decode(r.PopSlice(elmLen), pAdd(out, field.offset))
	}
}
