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

func popSetCodec(
	r *buff.Reader,
	id types.UUID,
	codecs []Codec,
) Codec {
	n := r.PopUint16()
	return &Set{id: id, child: codecs[n]}
}

// Set is an EdgeDB set type codec.
type Set struct {
	id    types.UUID
	child Codec
	typ   reflect.Type

	// step is the element width in bytes for a go array of type `Array.typ`.
	step int

	// useReflect indicates weather reflection or a known memory layout
	// should be used to deserialize data.
	useReflect bool
}

// ID returns the descriptor id.
func (c *Set) ID() types.UUID {
	return c.id
}

func (c *Set) setDefaultType() {
	c.child.setDefaultType()
	c.typ = reflect.SliceOf(c.child.Type())
	c.step = calcStep(c.typ.Elem())
	c.useReflect = true
}
func (c *Set) setType(typ reflect.Type, path Path) (bool, error) {
	if typ.Kind() != reflect.Slice {
		return false, fmt.Errorf(
			"expected %v to be a Slice got %v", path, typ.Kind(),
		)
	}

	c.typ = typ
	c.step = calcStep(typ.Elem())

	var err error
	c.useReflect, err = c.child.setType(typ.Elem(), path)
	return c.useReflect, err
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Set) Type() reflect.Type {
	return c.typ
}

// Decode a set
func (c *Set) Decode(r *buff.Reader, out reflect.Value) {
	if c.useReflect {
		c.DecodeReflect(r, out, Path(out.Type().String()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodeReflect decodes a set into a reflect.Value.
func (c *Set) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	// number of dimensions, either 0 or 1
	if r.PopUint32() == 0 {
		r.Discard(8) // skip 2 reserved fields
		return
	}

	r.Discard(8) // reserved

	upper := int32(r.PopUint32())
	lower := int32(r.PopUint32())
	n := int(upper - lower + 1)

	if out.Cap() < n {
		out.Set(reflect.MakeSlice(c.typ, n, n))
	} else {
		out.SetLen(n)
	}

	for i := 0; i < n; i++ {
		elmLen := r.PopUint32()

		c.child.DecodeReflect(
			r.PopSlice(elmLen),
			out.Index(i),
			path.AddIndex(i),
		)
	}
}

// DecodePtr decodes a set into an unsafe.Pointer.
func (c *Set) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	// number of dimensions, either 0 or 1
	if r.PopUint32() == 0 {
		r.Discard(8) // skip 2 reserved fields
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
		p := unsafe.Pointer(val.Pointer())
		*slice = *(*sliceHeader)(p)
	} else {
		slice.Len = n
	}

	for i := 0; i < n; i++ {
		elmLen := r.PopUint32()
		c.child.DecodePtr(
			r.PopSlice(elmLen),
			pAdd(slice.Data, uintptr(i*c.step)),
		)
	}
}

// Encode a set
func (c *Set) Encode(buf *buff.Writer, val interface{}, path Path) error {
	panic("sets can not be query parameters")
}
