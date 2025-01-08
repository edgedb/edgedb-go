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

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/geltypes"
	"github.com/edgedb/edgedb-go/internal/introspect"
)

var optionalTypeNameLookup = map[reflect.Type]string{
	reflect.TypeOf(&BoolCodec{}):          "gel.OptionalBool",
	reflect.TypeOf(&BytesCodec{}):         "gel.OptionalBytes",
	reflect.TypeOf(&DateTimeCodec{}):      "gel.OptionalDateTime",
	reflect.TypeOf(&LocalDateTimeCodec{}): "gel.OptionalLocalDateTime",
	reflect.TypeOf(&LocalDateCodec{}):     "gel.OptionalLocalDate",
	reflect.TypeOf(&LocalTimeCodec{}):     "gel.OptionalLocalTime",
	reflect.TypeOf(&DurationCodec{}):      "gel.OptionalDuration",
	reflect.TypeOf(
		&RelativeDurationCodec{}): "gel.OptionalRelativeDuration",
	reflect.TypeOf(&namedTupleDecoder{}): "gel.Optional",
	reflect.TypeOf(&Int16Codec{}):        "gel.OptionalInt16",
	reflect.TypeOf(&Int32Codec{}):        "gel.OptionalInt32",
	reflect.TypeOf(&Int64Codec{}):        "gel.OptionalInt64",
	reflect.TypeOf(&Float32Codec{}):      "gel.OptionalFloat32",
	reflect.TypeOf(&Float64Codec{}):      "gel.OptionalFloat64",
	reflect.TypeOf(&BigIntCodec{}):       "gel.OptionalBigInt",
	reflect.TypeOf(&objectDecoder{}):     "gel.Optional",
	reflect.TypeOf(&StrCodec{}):          "gel.OptionalStr",
	reflect.TypeOf(&tupleDecoder{}):      "gel.Optional",
	reflect.TypeOf(&UUIDCodec{}):         "gel.OptionalUUID",
}

func buildObjectDecoder(
	desc descriptor.Descriptor,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf(
			"expected %v to be a Struct got %v", path, typ.Kind(),
		)
	}

	fields := make([]*DecoderField, len(desc.Fields))

	for i, field := range desc.Fields {
		sf, ok := introspect.StructField(typ, field.Name)
		if !ok {
			return nil, fmt.Errorf(
				"expected %v to have a field named %q", path, field.Name,
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

		if !field.Required {
			if _, isOptional := child.(OptionalDecoder); !isOptional {
				typeName, ok := optionalTypeNameLookup[reflect.TypeOf(child)]
				if !ok {
					typeName = "OptionalUnmarshaler interface"
				}
				return nil, fmt.Errorf("expected %v at %v.%v to be %v "+
					"because the field is not required",
					sf.Type, path, field.Name, typeName)
			}
		}

		fields[i] = &DecoderField{
			name:    field.Name,
			offset:  sf.Offset,
			decoder: child,
		}
	}

	decoder := objectDecoder{desc.ID, fields}

	if reflect.PointerTo(typ).Implements(optionalUnmarshalerType) {
		return &optionalObjectDecoder{decoder, typ}, nil
	}

	return &decoder, nil
}

func buildObjectDecoderV2(
	desc *descriptor.V2,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf(
			"expected %v to be a Struct got %v", path, typ.Kind(),
		)
	}

	fields := make([]*DecoderField, len(desc.Fields))

	for i, field := range desc.Fields {
		sf, ok := introspect.StructField(typ, field.Name)
		if !ok {
			return nil, fmt.Errorf(
				"expected %v to have a field named %q", path, field.Name,
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

		if !field.Required {
			if _, isOptional := child.(OptionalDecoder); !isOptional {
				typeName, ok := optionalTypeNameLookup[reflect.TypeOf(child)]
				if !ok {
					typeName = "OptionalUnmarshaler interface"
				}
				return nil, fmt.Errorf("expected %v at %v.%v to be %v "+
					"because the field is not required",
					sf.Type, path, field.Name, typeName)
			}
		}

		fields[i] = &DecoderField{
			name:    field.Name,
			offset:  sf.Offset,
			decoder: child,
		}
	}

	decoder := objectDecoder{desc.ID, fields}

	if reflect.PointerTo(typ).Implements(optionalUnmarshalerType) {
		return &optionalObjectDecoder{decoder, typ}, nil
	}

	return &decoder, nil
}

type objectDecoder struct {
	id     types.UUID
	fields []*DecoderField
}

func (c *objectDecoder) DescriptorID() types.UUID { return c.id }

func (c *objectDecoder) Decode(r *buff.Reader, out unsafe.Pointer) error {
	elmCount := int(r.PopUint32())
	if elmCount != len(c.fields) {
		return fmt.Errorf(
			"wrong number of object fields: expected %v, got %v",
			len(c.fields), elmCount)
	}

	for _, field := range c.fields {
		r.Discard(4) // reserved

		p := pAdd(out, field.offset)
		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			// element length -1 means missing field
			// https://www.edgedb.com/docs/internals/protocol/dataformats
			field.decoder.(OptionalDecoder).DecodeMissing(p)
		} else {
			err := field.decoder.Decode(r.PopSlice(elmLen), p)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type optionalObjectDecoder struct {
	objectDecoder
	typ reflect.Type
}

func (c *optionalObjectDecoder) DecodeMissing(out unsafe.Pointer) {
	val := reflect.NewAt(c.typ, out)
	method := val.MethodByName("SetMissing")
	method.Call([]reflect.Value{trueValue})
}

func (c *optionalObjectDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	val := reflect.NewAt(c.typ, out)
	method := val.MethodByName("SetMissing")
	method.Call([]reflect.Value{falseValue})
	return c.objectDecoder.Decode(r, out)
}
