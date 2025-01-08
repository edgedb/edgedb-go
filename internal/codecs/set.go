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
)

func buildSetDecoder(
	desc descriptor.Descriptor,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf(
			"expected %v to be a Slice got %v", path, typ.Kind(),
		)
	}

	child, err := BuildDecoder(desc.Fields[0].Desc, typ.Elem(), path)
	if err != nil {
		return nil, err
	}

	return &setDecoder{desc.ID, child, typ, calcStep(typ.Elem())}, nil
}

func buildSetDecoderV2(
	desc *descriptor.V2,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf(
			"expected %v to be a Slice got %v", path, typ.Kind(),
		)
	}

	child, err := BuildDecoderV2(&desc.Fields[0].Desc, typ.Elem(), path)
	if err != nil {
		return nil, err
	}

	return &setDecoder{desc.ID, child, typ, calcStep(typ.Elem())}, nil
}

type setDecoder struct {
	id    types.UUID
	child Decoder
	typ   reflect.Type

	// step is the element width in bytes for a go array of type `Array.typ`.
	step int
}

func (c *setDecoder) DescriptorID() types.UUID { return c.id }

func nilUnsafePointer() unsafe.Pointer {
	var a []byte
	slice := (*sliceHeader)(unsafe.Pointer(&a))
	return slice.Data
}

var nilPointer = nilUnsafePointer()

func setSliceLen(slice *sliceHeader, typ reflect.Type, n int) {
	switch {
	case uintptr(slice.Data) == uintptr(0):
		// slice == nil
		val := reflect.New(typ)
		val.Elem().Set(reflect.MakeSlice(typ, n, n))
		p := unsafe.Pointer(val.Pointer())
		*slice = *(*sliceHeader)(p)
	case slice.Cap < n:
		val := reflect.New(typ)
		val.Elem().Set(reflect.MakeSlice(typ, n, n))
		p := unsafe.Pointer(val.Pointer())
		*slice = *(*sliceHeader)(p)
	default:
		slice.Len = n
	}
}

func (c *setDecoder) Decode(r *buff.Reader, out unsafe.Pointer) error {
	// number of dimensions, either 0 or 1
	if r.PopUint32() == 0 {
		r.Discard(8) // skip 2 reserved fields
		slice := (*sliceHeader)(out)
		setSliceLen(slice, c.typ, 0)
		return nil
	}

	r.Discard(8) // reserved

	upper := int32(r.PopUint32())
	lower := int32(r.PopUint32())
	n := int(upper - lower + 1)

	slice := (*sliceHeader)(out)
	setSliceLen(slice, c.typ, n)

	_, isSetOfArrays := c.child.(*arrayDecoder)

	for i := 0; i < n; i++ {
		if isSetOfArrays {
			r.Discard(12)
		}

		elmLen := r.PopUint32()
		err := c.child.Decode(
			r.PopSlice(elmLen),
			pAdd(slice.Data, uintptr(i*c.step)),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *setDecoder) DecodeMissing(out unsafe.Pointer) {
	slice := (*sliceHeader)(out)
	slice.Data = nilPointer
	slice.Len = 0
	slice.Cap = 0
}
