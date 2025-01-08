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

	"github.com/edgedb/edgedb-go/internal"
	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/geltypes"
)

func buildArrayEncoder(
	desc descriptor.Descriptor,
	version internal.ProtocolVersion,
) (Encoder, error) {
	child, err := BuildEncoder(desc.Fields[0].Desc, version)

	if err != nil {
		return nil, err
	}

	return &arrayEncoder{desc.ID, child}, nil
}

func buildArrayEncoderV2(
	desc *descriptor.V2,
	version internal.ProtocolVersion,
) (Encoder, error) {
	child, err := BuildEncoderV2(&desc.Fields[0].Desc, version)

	if err != nil {
		return nil, err
	}

	return &arrayEncoder{desc.ID, child}, nil
}

type arrayEncoder struct {
	id    types.UUID
	child Encoder
}

func (c *arrayEncoder) DescriptorID() types.UUID { return c.id }

func (c *arrayEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	in := reflect.ValueOf(val)
	if in.Kind() != reflect.Slice {
		return fmt.Errorf(
			"expected %v to be a slice got: %T", path, val,
		)
	}

	if in.IsNil() && required {
		return missingValueError(val, path)
	}

	if in.IsNil() {
		w.PushUint32(0xffffffff)
		return nil
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
		err = c.child.Encode(
			w,
			in.Index(i).Interface(),
			path.AddIndex(i),
			true,
		)
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}

func buildArrayDecoder(
	desc descriptor.Descriptor,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf(
			"expected %v to be a Slice, got %v", path, typ.Kind(),
		)
	}

	child, err := BuildDecoder(desc.Fields[0].Desc, typ.Elem(), path)
	if err != nil {
		return nil, err
	}

	return &arrayDecoder{desc.ID, child, typ, calcStep(typ.Elem())}, nil
}

func buildArrayDecoderV2(
	desc *descriptor.V2,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf(
			"expected %v to be a Slice, got %v", path, typ.Kind(),
		)
	}

	child, err := BuildDecoderV2(&desc.Fields[0].Desc, typ.Elem(), path)
	if err != nil {
		return nil, err
	}

	return &arrayDecoder{desc.ID, child, typ, calcStep(typ.Elem())}, nil
}

type arrayDecoder struct {
	id    types.UUID
	child Decoder
	typ   reflect.Type

	// step is the element width in bytes for a go array of type `Array.typ`.
	step int
}

func (c *arrayDecoder) DescriptorID() types.UUID { return c.id }

func (c *arrayDecoder) Decode(r *buff.Reader, out unsafe.Pointer) error {
	// number of dimensions is 1 or 0
	if r.PopUint32() == 0 {
		r.Discard(8) // reserved
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

	for i := 0; i < n; i++ {
		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

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

func (c *arrayDecoder) DecodeMissing(out unsafe.Pointer) {
	slice := (*sliceHeader)(out)
	slice.Data = nilPointer
	slice.Len = 0
	slice.Cap = 0
}
