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
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

func buildMultiRangeEncoder(
	desc descriptor.Descriptor,
	version internal.ProtocolVersion,
) (Encoder, error) {
	child, err := BuildEncoder(desc.Fields[0].Desc, version)

	if err != nil {
		return nil, err
	}

	return &multiRangeEncoder{desc.ID, child}, nil
}

func buildMultiRangeEncoderV2(
	desc *descriptor.V2,
	version internal.ProtocolVersion,
) (Encoder, error) {
	child, err := BuildEncoderV2(&desc.Fields[0].Desc, version)

	if err != nil {
		return nil, err
	}

	return &multiRangeEncoder{desc.ID, child}, nil
}

type multiRangeEncoder struct {
	id    types.UUID
	child Encoder
}

func (c *multiRangeEncoder) DescriptorID() types.UUID { return c.id }

func (c *multiRangeEncoder) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool, // todo do we need this?
) error {
	in := reflect.ValueOf(val)
	if in.Kind() != reflect.Slice {
		return fmt.Errorf(
			"expected %v to be a slice got: %T", path, val,
		)
	}

	if in.IsNil() && required { // todo do we need this?
		return missingValueError(val, path)
	}

	if in.IsNil() {
		w.PushUint32(0xffffffff)
		return nil
	}

	elmCount := in.Len()

	w.BeginBytes()
	w.PushUint32(uint32(elmCount))

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

func buildMultiRangeDecoder(
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

	return &multiRangeDecoder{desc.ID, child, typ, calcStep(typ.Elem())}, nil
}

func buildMultiRangeDecoderV2(
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

	return &multiRangeDecoder{desc.ID, child, typ, calcStep(typ.Elem())}, nil
}

type multiRangeDecoder struct {
	id    types.UUID
	child Decoder
	typ   reflect.Type

	// step is the element width in bytes for a go array of type `MultiRange.typ`.
	step int
}

func (c *multiRangeDecoder) DescriptorID() types.UUID { return c.id }

func (c *multiRangeDecoder) Decode(r *buff.Reader, out unsafe.Pointer) error {
	upper := int32(r.PopUint32())
	n := int(upper)

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

func (c *multiRangeDecoder) DecodeMissing(out unsafe.Pointer) {
	slice := (*sliceHeader)(out)
	slice.Data = nilPointer
	slice.Len = 0
	slice.Cap = 0
}
