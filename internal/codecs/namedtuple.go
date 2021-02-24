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
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
	"github.com/edgedb/edgedb-go/internal/marshal"
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
	id     types.UUID
	fields []*objectField
	typ    reflect.Type

	// useReflect indicates weather reflection or a known memory layout
	// should be used to deserialize data.
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

func (c *NamedTuple) setType(typ reflect.Type, path Path) (bool, error) {
	if typ.Kind() != reflect.Struct {
		return false, fmt.Errorf(
			"expected %v to be a struct got %v", path, typ.Kind(),
		)
	}

	for i := 0; i < len(c.fields); i++ {
		field := c.fields[i]

		f, ok := marshal.StructField(typ, field.name)
		if !ok {
			return false, fmt.Errorf(
				"%v struct is missing field %q", typ, field.name,
			)
		}

		useReflect, err := field.codec.setType(
			f.Type,
			path.AddField(field.name),
		)

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
		c.DecodeReflect(r, out, Path(out.Type().String()))
		return
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodeReflect decodes a named tuple into a reflect.Value.
func (c *NamedTuple) DecodeReflect(
	r *buff.Reader,
	out reflect.Value,
	path Path,
) {
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

func (c *NamedTuple) decodeReflectStruct(
	r *buff.Reader,
	out reflect.Value,
	path Path,
) {
	elmCount := int(int32(r.PopUint32()))

	for i := 0; i < elmCount; i++ {
		r.Discard(4) // reserved

		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

		field := c.fields[i]
		field.codec.DecodeReflect(
			r.PopSlice(elmLen),
			structField(out, field.name),
			path.AddField(field.name),
		)
	}
}

func (c *NamedTuple) decodeReflectMap(
	r *buff.Reader,
	out reflect.Value,
	path Path,
) {
	elmCount := int(int32(r.PopUint32()))
	out.Set(reflect.MakeMapWithSize(c.typ, elmCount))

	for i := 0; i < elmCount; i++ {
		r.Discard(4) // reserved

		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

		field := c.fields[i]
		val := reflect.New(field.codec.Type()).Elem()
		field.codec.DecodeReflect(
			r.PopSlice(elmLen),
			val,
			path.AddField(field.name),
		)
		out.SetMapIndex(reflect.ValueOf(field.name), val)
	}
}

// DecodePtr decodes a named tuple into an unsafe.Pointer.
func (c *NamedTuple) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	elmCount := int(int32(r.PopUint32()))

	for i := 0; i < elmCount; i++ {
		r.Discard(4) // reserved

		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

		field := c.fields[i]
		field.codec.DecodePtr(r.PopSlice(elmLen), pAdd(out, field.offset))
	}
}

// Encode a named tuple.
func (c *NamedTuple) Encode(w *buff.Writer, val interface{}, path Path) error {
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
	for i := 0; i < elmCount; i++ {
		w.PushUint32(0) // reserved
		field := c.fields[i]
		err = field.codec.Encode(w, in[field.name], path.AddField(field.name))
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
