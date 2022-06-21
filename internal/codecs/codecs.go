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

	"github.com/edgedb/edgedb-go/internal"
	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

// Encoder can encode objects into the data wire format.
type Encoder interface {
	DescriptorID() types.UUID
	Encode(*buff.Writer, interface{}, Path, bool) error
}

// EncoderField is a link to a child encoder
// used by objects, named tuples and tuples.
type EncoderField struct {
	name     string
	encoder  Encoder
	required bool
}

// Decoder can decode the data wire format into objects.
type Decoder interface {
	DescriptorID() types.UUID
	Decode(*buff.Reader, unsafe.Pointer) error
}

// OptionalDecoder is used when decoding optional shape fields.
type OptionalDecoder interface {
	Decoder
	DecodeMissing(unsafe.Pointer)
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
func BuildEncoder(
	desc descriptor.Descriptor,
	version internal.ProtocolVersion,
) (Encoder, error) {
	if desc.ID == descriptor.IDZero {
		return noOpEncoder{}, nil
	}

	switch desc.Type {
	case descriptor.Set:
		return nil, fmt.Errorf("sets can not be encoded")
	case descriptor.Object:
		if version.GTE(internal.ProtocolVersion{Major: 0, Minor: 12}) {
			return buildArgEncoder(desc, version)
		}
		return nil, errors.New("objects can not be encoded")
	case descriptor.BaseScalar, descriptor.Enum:
		return buildScalarEncoder(desc)
	case descriptor.Tuple:
		if version.GTE(internal.ProtocolVersion{Major: 0, Minor: 12}) {
			return nil, errors.New("tuples can not be encoded")
		}
		return buildTupleEncoder(desc, version)
	case descriptor.NamedTuple:
		if version.GTE(internal.ProtocolVersion{Major: 0, Minor: 12}) {
			return nil, errors.New("tuples can not be encoded")
		}
		return buildNamedTupleEncoder(desc, version)
	case descriptor.Array:
		return buildArrayEncoder(desc, version)
	default:
		return nil, fmt.Errorf("unknown descriptor type 0x%x", desc.Type)
	}
}

func buildScalarEncoder(desc descriptor.Descriptor) (Encoder, error) {
	if desc.ID == decimalID {
		return &decimalEncoder{}, nil
	}

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
	case memoryID:
		return &memoryCodec{}, nil
	default:
		s := fmt.Sprintf("%#v\n", desc)
		return nil, fmt.Errorf("unknown scalar type id %v %v", desc.ID, s)
	}
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
	decoder, ok, err := buildUnmarshaler(desc, typ)
	if err != nil {
		return decoder, err
	}
	if ok {
		return decoder, nil
	}

	var expectedType string

	if desc.Type == descriptor.Enum {
		switch typ {
		case strType:
			return &strCodec{desc.ID}, nil
		case optionalStrType:
			return &optionalStrDecoder{strID}, nil
		default:
			expectedType = "string or edgedb.OptionalStr"
			goto TypeMissmatch
		}
	}

	switch desc.ID {
	case uuidID:
		switch typ {
		case uuidType:
			return &uuidCodec{}, nil
		case optionalUUIDType:
			return &optionalUUIDDecoder{}, nil
		default:
			expectedType = "uuid or edgedb.OptionalUUID"
		}
	case strID:
		switch typ {
		case strType:
			return &strCodec{strID}, nil
		case optionalStrType:
			return &optionalStrDecoder{strID}, nil
		default:
			expectedType = "string or edgedb.OptionalStr"
		}
	case bytesID:
		switch typ {
		case bytesType:
			return &bytesCodec{bytesID}, nil
		case optionalBytesType:
			return &optionalBytesDecoder{bytesID}, nil
		default:
			expectedType = "[]byte or edgedb.OptionalBytes"
		}
	case int16ID:
		switch typ {
		case int16Type:
			return &int16Codec{}, nil
		case optionalInt16Type:
			return &optionalInt16Decoder{}, nil
		default:
			expectedType = "int16 or edgedb.OptionalInt16"
		}
	case int32ID:
		switch typ {
		case int32Type:
			return &int32Codec{}, nil
		case optionalInt32Type:
			return &optionalInt32Decoder{}, nil
		default:
			expectedType = "int32 or edgedb.OptionalInt32"
		}
	case int64ID:
		switch typ {
		case int64Type:
			return &int64Codec{}, nil
		case optionalInt64Type:
			return &optionalInt64Decoder{}, nil
		default:
			expectedType = "int64 or edgedb.OptionalInt64"
		}
	case float32ID:
		switch typ {
		case float32Type:
			return &float32Codec{}, nil
		case optionalFloat32Type:
			return &optionalFloat32Decoder{}, nil
		default:
			expectedType = "float32 or edgedb.OptionalFloat32"
		}
	case float64ID:
		switch typ {
		case float64Type:
			return &float64Codec{}, nil
		case optionalFloat64Type:
			return &optionalFloat64Decoder{}, nil
		default:
			expectedType = "float64 or edgedb.OptionalFloat64"
		}
	case decimalID:
		return nil, errors.New("decimal codec not implemented. " +
			"Consider implementing your own edgedb.DecimalMarshaler " +
			"and edgedb.DecimalUnmarshaler.")
	case boolID:
		switch typ {
		case boolType:
			return &boolCodec{}, nil
		case optionalBoolType:
			return &optionalBoolDecoder{}, nil
		default:
			expectedType = "bool or edgedb.OptionalBool"
		}
	case dateTimeID:
		switch typ {
		case dateTimeType:
			return &dateTimeCodec{}, nil
		case optionalDateTimeType:
			return &optionalDateTimeDecoder{}, nil
		default:
			expectedType = "edgedb.DateTime or edgedb.OptionalDateTime"
		}
	case localDTID:
		switch typ {
		case localDateTimeType:
			return &localDateTimeCodec{}, nil
		case optionalLocalDateTimeType:
			return &optionalLocalDateTimeDecoder{}, nil
		default:
			expectedType = "edgedb.LocalDateTime or " +
				"edgedb.OptionalLocalDateTime"
		}
	case localDateID:
		switch typ {
		case localDateType:
			return &localDateCodec{}, nil
		case optionalLocalDateType:
			return &optionalLocalDateDecoder{}, nil
		default:
			expectedType = "edgedb.LocalDate or edgedb.OptionalLocalDate"
		}
	case localTimeID:
		switch typ {
		case localTimeType:
			return &localTimeCodec{}, nil
		case optionalLocalTimeType:
			return &optionalLocalTimeDecoder{}, nil
		default:
			expectedType = "edgedb.LocalTime or edgedb.OptionalLocalTime"
		}
	case durationID:
		switch typ {
		case durationType:
			return &durationCodec{}, nil
		case optionalDurationType:
			return &optionalDurationDecoder{}, nil
		default:
			expectedType = "edgedb.Duration or edgedb.OptionalDuration"
		}
	case jsonID:
		ptr := reflect.PointerTo(typ)

		switch {
		case typ == bytesType:
			return &jsonCodec{typ: typ}, nil
		case typ == optionalBytesType:
			return &optionalJSONDecoder{typ: typ}, nil
		case ptr.Implements(optionalUnmarshalerType):
			return &optionalUnmarshalerJSONDecoder{typ: typ}, nil
		case ptr.Implements(optionalScalarUnmarshalerType):
			return &optionalScalarUnmarshalerJSONDecoder{typ: typ}, nil
		case typ.Kind() == reflect.Slice:
			fallthrough
		case typ.Kind() == reflect.Interface:
			return &optionalNilableJSONDecoder{typ: typ}, nil
		default:
			return &jsonCodec{typ: typ}, nil
		}
	case bigIntID:
		switch typ {
		case bigIntType:
			return &bigIntCodec{}, nil
		case optionalBigIntType:
			return &optionalBigIntDecoder{}, nil
		default:
			expectedType = "*big.Int or edgedb.OptionalBigInt"
		}
	case relativeDurationID:
		switch typ {
		case relativeDurationType:
			return &relativeDurationCodec{}, nil
		case optionalRelativeDurationType:
			return &optionalRelativeDurationDecoder{}, nil
		default:
			expectedType = "edgedb.RealtiveDuration or " +
				"edgedb.OptionalRelativeDuration"
		}
	case memoryID:
		switch typ {
		case memoryType:
			return &memoryCodec{}, nil
		case optionalMemoryType:
			return &optionalMemoryDecoder{}, nil
		default:
			expectedType = "edgedb.Memory or edgedb.OptionalMemory"
		}
	default:
		s := fmt.Sprintf("%#v\n", desc)
		return nil, fmt.Errorf("unknown scalar type id %v %v", desc.ID, s)
	}

TypeMissmatch:
	return nil, fmt.Errorf(
		"expected %v to be %v got %v", path, expectedType, typ,
	)
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
