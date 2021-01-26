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
	"github.com/edgedb/edgedb-go/internal/types"
)

func popTupleCodec(
	r *buff.Reader,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []Codec{}

	elmCount := int(r.PopUint16())
	for i := 0; i < elmCount; i++ {
		index := r.PopUint16()
		fields = append(fields, codecs[index])
	}

	return &Tuple{id: id, fields: fields}
}

// Tuple is an EdgeDB tuple type codec.
type Tuple struct {
	id     types.UUID
	fields []Codec
	typ    reflect.Type

	// useReflect indicates weather reflection or a known memory layout
	// should be used to deserialize data.
	useReflect bool
}

// ID returns the descriptor id.
func (c *Tuple) ID() types.UUID {
	return c.id
}

func (c *Tuple) setDefaultType() {
	for _, field := range c.fields {
		field.setDefaultType()
	}

	c.typ = reflect.TypeOf([]interface{}{})
	c.useReflect = true
}

func (c *Tuple) setType(typ reflect.Type, path Path) (bool, error) {
	expectedType := reflect.TypeOf([]interface{}{})

	if typ != expectedType {
		return false, fmt.Errorf(
			"expected %v to be []interface{} got %v", path, typ,
		)
	}

	c.typ = expectedType

	for _, field := range c.fields {
		// scalar codecs have a preset type
		if field.Type() == nil {
			c.useReflect = true
		}

		field.setDefaultType()
	}

	return c.useReflect, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Tuple) Type() reflect.Type {
	return c.typ
}

// Decode a tuple.
func (c *Tuple) Decode(r *buff.Reader, out reflect.Value) {
	if c.useReflect {
		c.DecodeReflect(r, out, Path(out.Type().String()))
		return
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodeReflect decodes a tuple into a reflect.Value.
func (c *Tuple) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	n := int(int32(r.PopUint32()))
	slice := reflect.MakeSlice(c.typ, 0, n)

	for i := 0; i < n; i++ {
		r.Discard(4) // reserved

		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

		field := c.fields[i]
		val := reflect.New(field.Type()).Elem()
		field.DecodeReflect(r.PopSlice(elmLen), val, path.AddIndex(i))
		slice = reflect.Append(slice, val)
	}

	out.Set(slice)
}

// DecodePtr decodes a tuple into an unsafe.Pointer.
func (c *Tuple) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	n := int(int32(r.PopUint32()))
	slice := reflect.MakeSlice(c.typ, 0, n)

	for i := 0; i < n; i++ {
		r.Discard(4) // reserved

		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

		field := c.fields[i]
		val := reflect.New(field.Type()).Elem()
		field.DecodePtr(r.PopSlice(elmLen), unsafe.Pointer(val.UnsafeAddr()))
		slice = reflect.Append(slice, val)
	}

	val := reflect.New(c.typ)
	val.Elem().Set(slice)
	*(*sliceHeader)(out) = *(*sliceHeader)(unsafe.Pointer(val.Pointer()))
}

// Encode a tuple.
func (c *Tuple) Encode(w *buff.Writer, val interface{}, path Path) error {
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
	for i := 0; i < elmCount; i++ {
		w.PushUint32(0) // reserved
		err = c.fields[i].Encode(w, in[i], path.AddIndex(i))
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
