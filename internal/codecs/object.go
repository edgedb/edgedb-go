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
	"github.com/edgedb/edgedb-go/internal/types"
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
	id     types.UUID
	fields []*objectField
	typ    reflect.Type

	// useReflect indicates weather reflection or a known memory layout
	// should be used to deserialize data.
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

func (c *Object) setType(typ reflect.Type, path Path) (bool, error) {
	if typ.Kind() != reflect.Struct {
		return false, fmt.Errorf(
			"expected %v to be a Struct got %v", path, typ.Kind(),
		)
	}

	for _, field := range c.fields {
		if field.name == "__tid__" {
			continue
		}

		f, ok := marshal.StructField(typ, field.name)
		if !ok {
			return false, fmt.Errorf(
				"expected %v to have a field named %q", path, field.name,
			)
		}

		useReflect, err := field.codec.setType(
			f.Type,
			path.AddField(field.name),
		)

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
		c.DecodeReflect(r, out, Path(out.Type().String()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodeReflect decodes an object into a reflect.Value.
func (c *Object) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != c.typ {
		panic(fmt.Sprintf(
			"expected %v to be %v, got %v", path, c.typ, out.Type(),
		))
	}

	switch out.Kind() {
	case reflect.Struct:
		c.decodeReflectStruct(r, out, path)
	case reflect.Map:
		c.decodeReflectMap(r, out, path)
	default:
		panic(fmt.Sprintf(
			"expected %v to be Struct or Map, got %v", path, out.Kind(),
		))
	}
}

func (c *Object) decodeReflectStruct(
	r *buff.Reader,
	out reflect.Value,
	path Path,
) {
	elmCount := int(r.PopUint32())
	if elmCount != len(c.fields) {
		panic(fmt.Sprintf(
			"wrong number of object fields: expected %v, got %v",
			len(c.fields),
			elmCount,
		))
	}

	for _, field := range c.fields {
		r.Discard(4) // reserved

		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			// element length -1 means missing field
			// https://www.edgedb.com/docs/internals/protocol/dataformats
			continue
		}

		if field.name == "__tid__" {
			r.Discard(16)
			continue
		}

		field.codec.DecodeReflect(
			r.PopSlice(elmLen),
			out.FieldByName(field.name),
			path.AddField(field.name),
		)
	}
}

func (c *Object) decodeReflectMap(
	r *buff.Reader,
	out reflect.Value,
	path Path,
) {
	elmCount := int(r.PopUint32())
	if elmCount != len(c.fields) {
		panic(fmt.Sprintf(
			"wrong number of object fields: expected %v, got %v",
			len(c.fields),
			elmCount,
		))
	}

	out.Set(reflect.MakeMapWithSize(c.typ, elmCount))

	for _, field := range c.fields {
		r.Discard(4) // reserved

		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			// element length -1 means missing field
			// https://www.edgedb.com/docs/internals/protocol/dataformats
			continue
		}

		if field.name == "__tid__" {
			r.Discard(16)
			continue
		}

		val := reflect.New(field.codec.Type()).Elem()
		field.codec.DecodeReflect(
			r.PopSlice(elmLen),
			val,
			path.AddField(field.name),
		)
		out.SetMapIndex(reflect.ValueOf(field.name), val)
	}
}

// DecodePtr decodes an object into an unsafe.Pointer.
func (c *Object) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	elmCount := int(r.PopUint32())
	if elmCount != len(c.fields) {
		panic(fmt.Sprintf(
			"wrong number of object fields: expected %v, got %v",
			len(c.fields),
			elmCount,
		))
	}

	for _, field := range c.fields {
		r.Discard(4) // reserved

		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			// element length -1 means missing field
			// https://www.edgedb.com/docs/internals/protocol/dataformats
			continue
		}

		if field.name == "__tid__" {
			r.Discard(16)
			continue
		}

		field.codec.DecodePtr(r.PopSlice(elmLen), pAdd(out, field.offset))
	}
}

// Encode an object
func (c *Object) Encode(buf *buff.Writer, val interface{}, path Path) error {
	panic("objects can't be query parameters")
}
