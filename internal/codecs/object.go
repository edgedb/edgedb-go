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
	"github.com/edgedb/edgedb-go/internal/marshal"
	"github.com/edgedb/edgedb-go/types"
)

func popObjectCodec(
	r *buff.Reader,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []*objectField{}

	elmCount := int(r.PopUint16())
	for i := 0; i < elmCount; i++ {
		flags := r.PopUint8()
		name := r.PopString()
		index := r.PopUint16()

		field := &objectField{
			isImplicit:     flags&0b1 != 0,
			isLinkProperty: flags&0b10 != 0,
			isLink:         flags&0b100 != 0,
			name:           name,
			codec:          codecs[index],
		}

		fields = append(fields, field)
	}

	return &Object{id: id, fields: fields}
}

type objectField struct {
	name           string
	offset         uintptr
	codec          Codec
	isImplicit     bool
	isLinkProperty bool
	isLink         bool
}

// Object is an EdgeDB object type codec.
type Object struct {
	id         types.UUID
	fields     []*objectField
	typ        reflect.Type
	useReflect bool
}

// ID returns the descriptor id.
func (c *Object) ID() types.UUID {
	return c.id
}

func (c *Object) setDefaultType() {
	for _, field := range c.fields {
		field.codec.setDefaultType()
	}

	c.typ = reflect.TypeOf(map[string]interface{}{})
	c.useReflect = true
}

func (c *Object) setType(typ reflect.Type) (bool, error) {
	if typ.Kind() != reflect.Struct {
		return false, fmt.Errorf("expected Struct got %v", typ.Kind())
	}

	for _, field := range c.fields {
		if field.name == "__tid__" {
			continue
		}

		f, ok := marshal.StructField(typ, field.name)
		if !ok {
			return false, fmt.Errorf(
				"%v struct is missing field %q",
				typ, field.name,
			)
		}

		useReflect, err := field.codec.setType(f.Type)
		if err != nil {
			return false, err
		}

		field.offset = f.Offset
		c.useReflect = c.useReflect || useReflect
	}

	c.typ = typ
	return c.useReflect, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Object) Type() reflect.Type {
	return c.typ
}

// Decode an object
func (c *Object) Decode(r *buff.Reader, out reflect.Value) {
	if c.useReflect {
		c.DecodeReflect(r, out)
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodeReflect decodes an object into a reflect.Value.
func (c *Object) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Sprintf(
			"object codec unexpected type: expected %v, but got %v",
			c.typ,
			out.Type(),
		))
	}

	switch out.Kind() {
	case reflect.Struct:
		c.decodeReflectStruct(r, out)
	case reflect.Map:
		c.decodeReflectMap(r, out)
	default:
		panic(fmt.Sprintf("object codec can not decode into %v", out.Kind()))
	}
}

func (c *Object) decodeReflectStruct(r *buff.Reader, out reflect.Value) {
	r.Discard(8) // data length & element count

	for _, field := range c.fields {
		r.Discard(4) // reserved

		switch int32(r.PeekUint32()) {
		case -1:
			// element length -1 means missing field
			// https://www.edgedb.com/docs/internals/protocol/dataformats
			r.Discard(4)
		default:
			if field.name == "__tid__" {
				r.Discard(20)
				break
			}

			field.codec.DecodeReflect(r, out.FieldByName(field.name))
		}
	}
}

func (c *Object) decodeReflectMap(r *buff.Reader, out reflect.Value) {
	r.Discard(8) // data length & element count
	out.Set(reflect.MakeMapWithSize(c.typ, len(c.fields)))

	for _, field := range c.fields {
		r.Discard(4) // reserved

		switch int32(r.PeekUint32()) {
		case -1:
			// element length -1 means missing field
			// https://www.edgedb.com/docs/internals/protocol/dataformats
			r.Discard(4)
		default:
			if field.name == "__tid__" {
				r.Discard(20)
				break
			}

			val := reflect.New(field.codec.Type()).Elem()
			field.codec.DecodeReflect(r, val)
			out.SetMapIndex(reflect.ValueOf(field.name), val)
		}
	}
}

// DecodePtr decodes an object into an unsafe.Pointer.
func (c *Object) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	r.Discard(8) // data length & element count

	for _, field := range c.fields {
		r.Discard(4) // reserved

		switch int32(r.PeekUint32()) {
		case -1:
			// element length -1 means missing field
			// https://www.edgedb.com/docs/internals/protocol/dataformats
			r.Discard(4)
		default:
			if field.name == "__tid__" {
				r.Discard(20)
				break
			}

			p := pAdd(out, field.offset)
			field.codec.DecodePtr(r, p)
		}
	}
}

// Encode an object
func (c *Object) Encode(buf *buff.Writer, val interface{}) error {
	panic("objects can't be query parameters")
}
