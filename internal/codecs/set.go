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
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
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

type setDecoder struct {
	id    types.UUID
	child Decoder
	typ   reflect.Type

	// step is the element width in bytes for a go array of type `Array.typ`.
	step int
}

func (c *setDecoder) DescriptorID() types.UUID { return c.id }

func (c *setDecoder) Decode(r *buff.Reader, out unsafe.Pointer) {
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

	_, isSetOfArrays := c.child.(*arrayDecoder)

	for i := 0; i < n; i++ {
		if isSetOfArrays {
			r.Discard(12)
		}

		elmLen := r.PopUint32()
		c.child.Decode(
			r.PopSlice(elmLen),
			pAdd(slice.Data, uintptr(i*c.step)),
		)
	}
}
