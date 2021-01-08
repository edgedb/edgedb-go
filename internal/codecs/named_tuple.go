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

func popNamedTupleCodec(
	r *buff.Reader,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []*objectField{}

	elmCount := int(r.PopUint16())
	for i := 0; i < elmCount; i++ {
		name := r.PopString()
		index := r.PopUint16()

		if name == "__tid__" {
			continue
		}

		field := &objectField{
			name:  name,
			codec: codecs[index],
		}

		fields = append(fields, field)
	}

	return &NamedTuple{id: id, fields: fields}
}

// NamedTuple is an EdgeDB namedtuple type codec.
type NamedTuple struct {
	id         types.UUID
	fields     []*objectField
	typ        reflect.Type
	useReflect bool
}

// ID returns the descriptor id.
func (c *NamedTuple) ID() types.UUID {
	return c.id
}

func (c *NamedTuple) setDefaultType() {
	for _, field := range c.fields {
		field.codec.setDefaultType()
	}

	c.typ = reflect.TypeOf(map[string]interface{}{})
	c.useReflect = true
}

func (c *NamedTuple) setType(typ reflect.Type) (bool, error) {
	if typ.Kind() != reflect.Struct {
		return false, fmt.Errorf("expected Struct got %v", typ.Kind())
	}

	for i := 0; i < len(c.fields); i++ {
		field := c.fields[i]

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

		c.useReflect = c.useReflect || useReflect
		field.offset = f.Offset
	}

	c.typ = typ
	return c.useReflect, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *NamedTuple) Type() reflect.Type {
	return c.typ
}

// Decode a named tuple.
func (c *NamedTuple) Decode(r *buff.Reader, out reflect.Value) {
	if c.useReflect {
		c.DecodeReflect(r, out)
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodeReflect decodes a named tuple into a reflect.Value.
func (c *NamedTuple) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Sprintf(
			"named tuple codec unexpected type: expected %v, but got %v",
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
		panic(fmt.Sprintf(
			"named tuple codec can not decode into %v",
			out.Kind(),
		))
	}
}

func (c *NamedTuple) decodeReflectStruct(r *buff.Reader, out reflect.Value) {
	r.Discard(4) // data length
	elmCount := int(int32(r.PopUint32()))

	for i := 0; i < elmCount; i++ {
		r.Discard(4) // reserved
		field := c.fields[i]
		field.codec.DecodeReflect(r, out.FieldByName(field.name))
	}
}

func (c *NamedTuple) decodeReflectMap(r *buff.Reader, out reflect.Value) {
	r.Discard(4) // data length
	elmCount := int(int32(r.PopUint32()))
	out.Set(reflect.MakeMapWithSize(c.typ, elmCount))

	for i := 0; i < elmCount; i++ {
		r.Discard(4) // reserved
		field := c.fields[i]
		val := reflect.New(field.codec.Type()).Elem()
		field.codec.DecodeReflect(r, val)
		out.SetMapIndex(reflect.ValueOf(field.name), val)
	}
}

// DecodePtr decodes a named tuple into an unsafe.Pointer.
func (c *NamedTuple) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	r.Discard(4) // data length
	elmCount := int(int32(r.PopUint32()))

	for i := 0; i < elmCount; i++ {
		r.Discard(4) // reserved
		field := c.fields[i]
		field.codec.DecodePtr(r, pAdd(out, field.offset))
	}
}

// Encode a named tuple.
func (c *NamedTuple) Encode(w *buff.Writer, val interface{}) error {
	args, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("expected []interface{} got %T", val)
	}

	elmCount := len(c.fields)

	w.BeginBytes()
	w.PushUint32(uint32(elmCount))

	if len(args) != 1 {
		panic(fmt.Sprintf(
			"wrong number of arguments, expected 1 got: %v",
			len(args),
		))
	}

	in := args[0].(map[string]interface{})

	var err error
	for i := 0; i < elmCount; i++ {
		w.PushUint32(0) // reserved
		field := c.fields[i]
		err = field.codec.Encode(w, in[field.name])
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
