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

func popArrayCodec(
	r *buff.Reader,
	id types.UUID,
	codecs []Codec,
) Codec {
	i := r.PopUint16() // element type descriptor index

	n := int(r.PopUint16()) // number of array dimensions
	if n == 0 {
		panic("too few array dimensions: expected at least 1, got 0")
	}

	for i := 0; i < n; i++ {
		r.Discard(4) // array dimension
	}

	return &Array{id: id, child: codecs[i]}
}

// Array is an EdgeDB array type codec.
type Array struct {
	id    types.UUID
	child Codec
	typ   reflect.Type

	// step is the element width in bytes for a go array of type `Array.typ`.
	step int

	// useReflect indicates weather reflection or a known memory layout
	// should be used to deserialize data.
	useReflect bool
}

func (c *Array) setDefaultType() {
	c.child.setDefaultType()
	c.typ = reflect.SliceOf(c.child.Type())
	c.step = calcStep(c.typ.Elem())
	c.useReflect = true
}

func (c *Array) setType(typ reflect.Type, path Path) (bool, error) {
	if typ.Kind() != reflect.Slice {
		return false, fmt.Errorf(
			"expected %v to be a Slice, got %v", path, typ.Kind(),
		)
	}

	c.typ = typ
	c.step = calcStep(typ.Elem())

	var err error
	c.useReflect, err = c.child.setType(typ.Elem(), path)
	return c.useReflect, err
}

// ID returns the descriptor id.
func (c *Array) ID() types.UUID {
	return c.id
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Array) Type() reflect.Type {
	typ := c.child.Type()

	if typ == nil {
		return nil
	}

	return reflect.SliceOf(typ)
}

// Decode an array.
func (c *Array) Decode(r *buff.Reader, out reflect.Value) {
	if c.useReflect {
		c.DecodeReflect(r, out, Path(out.Type().String()))
		return
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodeReflect decodes an array into a reflect.Value.
func (c *Array) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != c.Type() {
		panic(fmt.Sprintf(
			"expected %v to be %v got: %v", path, c.Type(), out.Type(),
		))
	}

	// number of dimensions is 1 or 0
	if r.PopUint32() == 0 {
		r.Discard(8) // reserved
		return
	}

	r.Discard(8) // reserved

	upper := int32(r.PopUint32())
	lower := int32(r.PopUint32())
	n := int(upper - lower + 1)

	if out.Cap() < n {
		out.Set(reflect.MakeSlice(c.Type(), n, n))
	}

	if out.Len() > n {
		out.Set(out.Slice(0, n))
	}

	for i := 0; i < n; i++ {
		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

		c.child.DecodeReflect(
			r.PopSlice(elmLen),
			out.Index(i),
			path.AddIndex(i),
		)
	}
}

// DecodePtr decodes an array into an unsafe.Pointer.
func (c *Array) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	// number of dimensions is 1 or 0
	if r.PopUint32() == 0 {
		r.Discard(8) // reserved
		return
	}

	r.Discard(8) // reserved

	upper := int32(r.PopUint32())
	lower := int32(r.PopUint32())
	n := int(upper - lower + 1)

	slice := (*sliceHeader)(out)
	if slice.Cap < n {
		val := reflect.New(c.typ)
		val.Elem().Set(reflect.MakeSlice(c.typ, n, n))
		*slice = *(*sliceHeader)(unsafe.Pointer(val.Pointer()))
	} else {
		slice.Len = n
	}

	for i := 0; i < n; i++ {
		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

		c.child.DecodePtr(
			r.PopSlice(elmLen),
			pAdd(slice.Data, uintptr(i*c.step)),
		)
	}
}

// Encode an array.
func (c *Array) Encode(w *buff.Writer, val interface{}, path Path) error {
	in := reflect.ValueOf(val)
	if in.Kind() != reflect.Slice {
		return fmt.Errorf(
			"expected %v to be %v got: %T", path, c.Type(), val,
		)
	}

	elmCount := in.Len()

	w.BeginBytes()
	w.PushUint32(1)                // number of dimensions
	w.PushUint32(0)                // reserved
	w.PushUint32(0)                // reserved
	w.PushUint32(uint32(elmCount)) // dimension.upper
	w.PushUint32(1)                // dimension.lower

	var err error
	for i := 0; i < elmCount; i++ {
		err = c.child.Encode(w, in.Index(i).Interface(), path.AddIndex(i))
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
