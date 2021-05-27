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
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

// Encoder can encode objects into the data wire format.
type Encoder interface {
	DescriptorID() types.UUID
	Encode(*buff.Writer, interface{}, Path) error
}

// EncoderField is a link to a child encoder
// used by objects, named tuples and tuples.
type EncoderField struct {
	name    string
	encoder Encoder
}

// Decoder can decode the data wire format into objects.
type Decoder interface {
	DescriptorID() types.UUID
	Decode(*buff.Reader, unsafe.Pointer)
}

// DecoderField is a link to a child decoder
// used by objects, named tuples and tuples.
type DecoderField struct {
	name    string
	offset  uintptr
	decoder Decoder
}

// Codec can Encode and Decode
type Codec interface {
	Encoder
	Decoder
	Type() reflect.Type
}

// BuildEncoder builds and Encoder from a Descriptor.
func BuildEncoder(desc descriptor.Descriptor) (Encoder, error) {
	switch desc.Type {
	case descriptor.Set:
		return nil, fmt.Errorf("sets can not be encoded")
	case descriptor.Object:
		return nil, fmt.Errorf("objects can not be encoded")
	case descriptor.BaseScalar, descriptor.Enum:
		return buildScalarEncoder(desc)
	case descriptor.Tuple:
		return buildTupleEncoder(desc)
	case descriptor.NamedTuple:
		return buildNamedTupleEncoder(desc)
	case descriptor.Array:
		return buildArrayEncoder(desc)
	default:
		return nil, fmt.Errorf("unknown descriptor type 0x%x", desc.Type)
	}
}

func buildScalarEncoder(desc descriptor.Descriptor) (Encoder, error) {
	if desc.ID == decimalID {
		return &decimalEncoder{}, nil
	}

	return buildScalarCodec(desc)
}

// BuildDecoder builds a Decoder from a Descriptor.
func BuildDecoder(
	desc descriptor.Descriptor,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if desc.ID == descriptor.IDZero {
		return noOpDecoder{}, nil
	}

	switch desc.Type {
	case descriptor.Set:
		return buildSetDecoder(desc, typ, path)
	case descriptor.Object:
		return buildObjectDecoder(desc, typ, path)
	case descriptor.BaseScalar, descriptor.Enum:
		return buildScalarDecoder(desc, typ, path)
	case descriptor.Tuple:
		return buildTupleDecoder(desc, typ, path)
	case descriptor.NamedTuple:
		return buildNamedTupleDecoder(desc, typ, path)
	case descriptor.Array:
		return buildArrayDecoder(desc, typ, path)
	default:
		return nil, fmt.Errorf("unknown descriptor type 0x%x", desc.Type)
	}
}

func buildScalarDecoder(
	desc descriptor.Descriptor,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	decoder, ok := buildUnmarshaler(desc, typ)
	if ok {
		return decoder, nil
	}

	codec, err := buildScalarCodec(desc)
	if err != nil {
		return nil, err
	}

	if codec.Type() != typ {
		return nil, fmt.Errorf(
			"expected %v to be %v got %v", path, codec.Type(), typ,
		)
	}

	return codec, nil
}

func buildScalarCodec(desc descriptor.Descriptor) (Codec, error) {
	if desc.Type == descriptor.Enum {
		return &strCodec{desc.ID}, nil
	}

	switch desc.ID {
	case uuidID:
		return &uuidCodec{}, nil
	case strID:
		return &strCodec{strID}, nil
	case bytesID:
		return &bytesCodec{bytesID}, nil
	case int16ID:
		return &int16Codec{}, nil
	case int32ID:
		return &int32Codec{}, nil
	case int64ID:
		return &int64Codec{}, nil
	case float32ID:
		return &float32Codec{}, nil
	case float64ID:
		return &float64Codec{}, nil
	case decimalID:
		return nil, errors.New("decimal codec not implemented. " +
			"Consider implementing your own edgedb.DecimalMarshaler " +
			"and edgedb.DecimalUnmarshaler.")
	case boolID:
		return &boolCodec{}, nil
	case dateTimeID:
		return &dateTimeCodec{}, nil
	case localDTID:
		return &localDateTimeCodec{}, nil
	case localDateID:
		return &localDateCodec{}, nil
	case localTimeID:
		return &localTimeCodec{}, nil
	case durationID:
		return &durationCodec{}, nil
	case jsonID:
		return &jsonCodec{}, nil
	case bigIntID:
		return &bigIntCodec{}, nil
	case relativeDurationID:
		return &relativeDurationCodec{}, nil
	default:
		s := fmt.Sprintf("%#v\n", desc)
		return nil, fmt.Errorf("unknown scalar type id %v %v", desc.ID, s)
	}
}

func pAdd(p unsafe.Pointer, i uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + i)
}

// calcStep returns the element width in bytes for a go array of `typ`.
func calcStep(typ reflect.Type) int {
	step := int(typ.Size())
	a := typ.Align()

	if step%a > 0 {
		step = step/a + a
	}

	return step
}

// sliceHeader represent the memory layout for a slice.
type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}
