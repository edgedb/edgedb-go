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
	"reflect"
	"unsafe"

	"github.com/geldata/gel-go/internal"
	"github.com/geldata/gel-go/internal/buff"
	"github.com/geldata/gel-go/internal/descriptor"
	types "github.com/geldata/gel-go/internal/geltypes"
	"github.com/geldata/gel-go/internal/introspect"
)

func buildNamedTupleEncoder(
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
			name:    field.Name,
			encoder: encoder,
		}
	}

	return &namedTupleEncoder{desc.ID, fields}, nil
}

type namedTupleEncoder struct {
	id     types.UUID
	fields []*EncoderField
}

func (c *namedTupleEncoder) DescriptorID() types.UUID { return c.id }

func (c *namedTupleEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	_ bool,
) error {
	args, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("expected %v to be []interface{} got %T", path, val)
	}

	elmCount := len(c.fields)

	w.BeginBytes()
	w.PushUint32(uint32(elmCount))

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

	var err error
	for _, field := range c.fields {
		w.PushUint32(0) // reserved
		err = field.encoder.Encode(
			w,
			in[field.name],
			path.AddField(field.name),
			true,
		)

		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}

func buildNamedTupleDecoder(
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
		sf, ok := introspect.StructField(typ, field.Name)
		if !ok {
			return nil, fmt.Errorf(
				"%v struct is missing field %q", typ, field.Name,
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

	decoder := namedTupleDecoder{desc.ID, fields}

	if reflect.PointerTo(typ).Implements(optionalUnmarshalerType) {
		return &optionalNamedTupleDecoder{decoder, typ}, nil
	}

	return &decoder, nil
}

func buildNamedTupleDecoderV2(
	desc *descriptor.V2,
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
		sf, ok := introspect.StructField(typ, field.Name)
		if !ok {
			return nil, fmt.Errorf(
				"%v struct is missing field %q", typ, field.Name,
			)
		}

		child, err := BuildDecoderV2(
			&field.Desc,
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

	decoder := namedTupleDecoder{desc.ID, fields}

	if reflect.PointerTo(typ).Implements(optionalUnmarshalerType) {
		return &optionalNamedTupleDecoder{decoder, typ}, nil
	}

	return &decoder, nil
}

type namedTupleDecoder struct {
	id     types.UUID
	fields []*DecoderField
}

func (c *namedTupleDecoder) DescriptorID() types.UUID { return c.id }

func (c *namedTupleDecoder) Decode(r *buff.Reader, out unsafe.Pointer) error {
	elmCount := int(int32(r.PopUint32()))
	if elmCount != len(c.fields) {
		return fmt.Errorf(
			"wrong number of elements expected %v got %v",
			len(c.fields), elmCount)
	}

	for _, field := range c.fields {
		r.Discard(4) // reserved

		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

		err := field.decoder.Decode(
			r.PopSlice(elmLen),
			pAdd(out, field.offset),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

type optionalNamedTupleDecoder struct {
	namedTupleDecoder
	typ reflect.Type
}

func (c *optionalNamedTupleDecoder) DecodeMissing(out unsafe.Pointer) {
	val := reflect.NewAt(c.typ, out)
	method := val.MethodByName("SetMissing")
	method.Call([]reflect.Value{trueValue})
}

func (c *optionalNamedTupleDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	val := reflect.NewAt(c.typ, out)
	method := val.MethodByName("SetMissing")
	method.Call([]reflect.Value{falseValue})
	return c.namedTupleDecoder.Decode(r, out)
}
