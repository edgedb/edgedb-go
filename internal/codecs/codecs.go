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
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/geldata/gel-go/internal"
	"github.com/geldata/gel-go/internal/buff"
	"github.com/geldata/gel-go/internal/descriptor"
	types "github.com/geldata/gel-go/internal/geltypes"
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
	case descriptor.BaseScalar, descriptor.Enum, descriptor.Scalar:
		return BuildScalarEncoder(desc)
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
	case descriptor.Range:
		return buildRangeEncoder(desc, version)
	default:
		return nil, fmt.Errorf(
			"building encoder: unknown descriptor type 0x%x",
			desc.Type)
	}
}

// BuildEncoderV2 builds and Encoder from a Descriptor.
func BuildEncoderV2(
	desc *descriptor.V2,
	version internal.ProtocolVersion,
) (Encoder, error) {
	if desc.ID == descriptor.IDZero {
		return noOpEncoder{}, nil
	}

	switch desc.Type {
	case descriptor.Set:
		return nil, fmt.Errorf("sets can not be encoded")
	case descriptor.Object:
		return buildArgEncoderV2(desc, version)
	case descriptor.BaseScalar, descriptor.Enum, descriptor.Scalar:
		return BuildScalarEncoderV2(desc)
	case descriptor.Tuple:
		return nil, errors.New("tuples can not be encoded")
	case descriptor.NamedTuple:
		return nil, errors.New("tuples can not be encoded")
	case descriptor.Array:
		return buildArrayEncoderV2(desc, version)
	case descriptor.Range:
		return buildRangeEncoderV2(desc, version)
	case descriptor.MultiRange:
		return buildMultiRangeEncoderV2(desc, version)
	default:
		return nil, fmt.Errorf(
			"building encoder: unknown descriptor type 0x%x",
			desc.Type)
	}
}

// GetScalarDescriptor finds the BaseScalar descriptor at the root of the
// inheritance chain for a Scalar descriptor.
func GetScalarDescriptor(desc descriptor.Descriptor) descriptor.Descriptor {
	for desc.Fields != nil {
		desc = desc.Fields[0].Desc
	}

	return desc
}

// GetScalarDescriptorV2 finds the BaseScalar descriptor at the root of the
// inheritance chain for a Scalar descriptor.
func GetScalarDescriptorV2(
	desc *descriptor.V2,
) *descriptor.V2 {
	if len(desc.Ancestors) > 0 {
		return &desc.Ancestors[len(desc.Ancestors)-1].Desc
	}

	return desc
}

// BuildScalarEncoder builds a scalar encoder.
func BuildScalarEncoder(desc descriptor.Descriptor) (Encoder, error) {
	if desc.Type == descriptor.Scalar {
		desc = GetScalarDescriptor(desc)
	}

	if desc.ID == DecimalID {
		return &decimalEncoder{}, nil
	}

	if desc.Type == descriptor.Enum {
		return &StrCodec{desc.ID}, nil
	}

	switch desc.ID {
	case UUIDID:
		return &UUIDCodec{}, nil
	case StrID:
		return &StrCodec{StrID}, nil
	case BytesID:
		return &BytesCodec{BytesID}, nil
	case Int16ID:
		return &Int16Codec{}, nil
	case Int32ID:
		return &Int32Codec{}, nil
	case Int64ID:
		return &Int64Codec{}, nil
	case Float32ID:
		return &Float32Codec{}, nil
	case Float64ID:
		return &Float64Codec{}, nil
	case DecimalID:
		return nil, errors.New("decimal codec not implemented. " +
			"Consider implementing your own gel.DecimalMarshaler " +
			"and gel.DecimalUnmarshaler.")
	case BoolID:
		return &BoolCodec{}, nil
	case DateTimeID:
		return &DateTimeCodec{}, nil
	case LocalDTID:
		return &LocalDateTimeCodec{}, nil
	case LocalDateID:
		return &LocalDateCodec{}, nil
	case LocalTimeID:
		return &LocalTimeCodec{}, nil
	case DurationID:
		return &DurationCodec{}, nil
	case JSONID:
		return &JSONCodec{}, nil
	case BigIntID:
		return &BigIntCodec{}, nil
	case RelativeDurationID:
		return &RelativeDurationCodec{}, nil
	case DateDurationID:
		return &DateDurationCodec{}, nil
	case MemoryID:
		return &MemoryCodec{}, nil
	default:
		s := fmt.Sprintf("%#v\n", desc)
		return nil, fmt.Errorf("unknown scalar type id %v %v", desc.ID, s)
	}
}

// BuildScalarEncoderV2 builds a scalar encoder.
func BuildScalarEncoderV2(desc *descriptor.V2) (Encoder, error) {
	if desc.Type == descriptor.Scalar {
		desc = GetScalarDescriptorV2(desc)
	}

	if desc.ID == DecimalID {
		return &decimalEncoder{}, nil
	}

	if desc.Type == descriptor.Enum {
		return &StrCodec{desc.ID}, nil
	}

	switch desc.ID {
	case UUIDID:
		return &UUIDCodec{}, nil
	case StrID:
		return &StrCodec{StrID}, nil
	case BytesID:
		return &BytesCodec{BytesID}, nil
	case Int16ID:
		return &Int16Codec{}, nil
	case Int32ID:
		return &Int32Codec{}, nil
	case Int64ID:
		return &Int64Codec{}, nil
	case Float32ID:
		return &Float32Codec{}, nil
	case Float64ID:
		return &Float64Codec{}, nil
	case DecimalID:
		return nil, errors.New("decimal codec not implemented. " +
			"Consider implementing your own gel.DecimalMarshaler " +
			"and gel.DecimalUnmarshaler.")
	case BoolID:
		return &BoolCodec{}, nil
	case DateTimeID:
		return &DateTimeCodec{}, nil
	case LocalDTID:
		return &LocalDateTimeCodec{}, nil
	case LocalDateID:
		return &LocalDateCodec{}, nil
	case LocalTimeID:
		return &LocalTimeCodec{}, nil
	case DurationID:
		return &DurationCodec{}, nil
	case JSONID:
		return &JSONCodec{}, nil
	case BigIntID:
		return &BigIntCodec{}, nil
	case RelativeDurationID:
		return &RelativeDurationCodec{}, nil
	case DateDurationID:
		return &DateDurationCodec{}, nil
	case MemoryID:
		return &MemoryCodec{}, nil
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
	case descriptor.BaseScalar, descriptor.Enum, descriptor.Scalar:
		return buildScalarDecoder(desc, typ, path)
	case descriptor.Tuple:
		return buildTupleDecoder(desc, typ, path)
	case descriptor.NamedTuple:
		return buildNamedTupleDecoder(desc, typ, path)
	case descriptor.Array:
		return buildArrayDecoder(desc, typ, path)
	case descriptor.Range:
		return buildRangeDecoder(desc, typ, path)
	default:
		return nil, fmt.Errorf(
			"building decoder: unknown descriptor type 0x%x",
			desc.Type)
	}
}

// BuildDecoderV2 builds a Decoder from a Descriptor.
func BuildDecoderV2(
	desc *descriptor.V2,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if desc.ID == descriptor.IDZero {
		return noOpDecoder{}, nil
	}

	switch desc.Type {
	case descriptor.Set:
		return buildSetDecoderV2(desc, typ, path)
	case descriptor.Object, descriptor.SQLRecord:
		return buildObjectDecoderV2(desc, typ, path)
	case descriptor.BaseScalar, descriptor.Enum, descriptor.Scalar:
		return buildScalarDecoderV2(desc, typ, path)
	case descriptor.Tuple:
		return buildTupleDecoderV2(desc, typ, path)
	case descriptor.NamedTuple:
		return buildNamedTupleDecoderV2(desc, typ, path)
	case descriptor.Array:
		return buildArrayDecoderV2(desc, typ, path)
	case descriptor.Range:
		return buildRangeDecoderV2(desc, typ, path)
	case descriptor.MultiRange:
		return buildMultiRangeDecoderV2(desc, typ, path)
	default:
		return nil, fmt.Errorf(
			"building decoder: unknown descriptor type 0x%x",
			desc.Type)
	}
}

func buildScalarDecoder(
	desc descriptor.Descriptor,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if desc.Type == descriptor.Scalar {
		desc = GetScalarDescriptor(desc)
	}

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
			return &StrCodec{desc.ID}, nil
		case optionalStrType:
			return &optionalStrDecoder{StrID}, nil
		default:
			expectedType = "string or gel.OptionalStr"
			goto TypeMissmatch
		}
	}

	switch desc.ID {
	case UUIDID:
		switch typ {
		case uuidType:
			return &UUIDCodec{}, nil
		case optionalUUIDType:
			return &optionalUUIDDecoder{}, nil
		default:
			expectedType = "uuid or gel.OptionalUUID"
		}
	case StrID:
		switch typ {
		case strType:
			return &StrCodec{StrID}, nil
		case optionalStrType:
			return &optionalStrDecoder{StrID}, nil
		default:
			expectedType = "string or gel.OptionalStr"
		}
	case BytesID:
		switch typ {
		case bytesType:
			return &BytesCodec{BytesID}, nil
		case optionalBytesType:
			return &optionalBytesDecoder{BytesID}, nil
		default:
			expectedType = "[]byte or gel.OptionalBytes"
		}
	case Int16ID:
		switch typ {
		case int16Type:
			return &Int16Codec{}, nil
		case optionalInt16Type:
			return &optionalInt16Decoder{}, nil
		default:
			expectedType = "int16 or gel.OptionalInt16"
		}
	case Int32ID:
		switch typ {
		case int32Type:
			return &Int32Codec{}, nil
		case optionalInt32Type:
			return &optionalInt32Decoder{}, nil
		default:
			expectedType = "int32 or gel.OptionalInt32"
		}
	case Int64ID:
		switch typ {
		case int64Type:
			return &Int64Codec{}, nil
		case optionalInt64Type:
			return &optionalInt64Decoder{}, nil
		default:
			expectedType = "int64 or gel.OptionalInt64"
		}
	case Float32ID:
		switch typ {
		case float32Type:
			return &Float32Codec{}, nil
		case optionalFloat32Type:
			return &optionalFloat32Decoder{}, nil
		default:
			expectedType = "float32 or gel.OptionalFloat32"
		}
	case Float64ID:
		switch typ {
		case float64Type:
			return &Float64Codec{}, nil
		case optionalFloat64Type:
			return &optionalFloat64Decoder{}, nil
		default:
			expectedType = "float64 or gel.OptionalFloat64"
		}
	case DecimalID:
		return nil, errors.New("decimal codec not implemented. " +
			"Consider implementing your own gel.DecimalMarshaler " +
			"and gel.DecimalUnmarshaler.")
	case BoolID:
		switch typ {
		case boolType:
			return &BoolCodec{}, nil
		case optionalBoolType:
			return &optionalBoolDecoder{}, nil
		default:
			expectedType = "bool or gel.OptionalBool"
		}
	case DateTimeID:
		switch typ {
		case dateTimeType:
			return &DateTimeCodec{}, nil
		case optionalDateTimeType:
			return &optionalDateTimeDecoder{}, nil
		default:
			expectedType = "gel.DateTime or gel.OptionalDateTime"
		}
	case LocalDTID:
		switch typ {
		case localDateTimeType:
			return &LocalDateTimeCodec{}, nil
		case optionalLocalDateTimeType:
			return &optionalLocalDateTimeDecoder{}, nil
		default:
			expectedType = "gel.LocalDateTime or " +
				"gel.OptionalLocalDateTime"
		}
	case LocalDateID:
		switch typ {
		case localDateType:
			return &LocalDateCodec{}, nil
		case optionalLocalDateType:
			return &optionalLocalDateDecoder{}, nil
		default:
			expectedType = "gel.LocalDate or gel.OptionalLocalDate"
		}
	case LocalTimeID:
		switch typ {
		case localTimeType:
			return &LocalTimeCodec{}, nil
		case optionalLocalTimeType:
			return &optionalLocalTimeDecoder{}, nil
		default:
			expectedType = "gel.LocalTime or gel.OptionalLocalTime"
		}
	case DurationID:
		switch typ {
		case durationType:
			return &DurationCodec{}, nil
		case optionalDurationType:
			return &optionalDurationDecoder{}, nil
		default:
			expectedType = "gel.Duration or gel.OptionalDuration"
		}
	case JSONID:
		ptr := reflect.PointerTo(typ)

		switch {
		case typ == bytesType:
			return &JSONCodec{typ: typ}, nil
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
			return &JSONCodec{typ: typ}, nil
		}
	case BigIntID:
		switch typ {
		case bigIntType:
			return &BigIntCodec{}, nil
		case optionalBigIntType:
			return &optionalBigIntDecoder{}, nil
		default:
			expectedType = "*big.Int or gel.OptionalBigInt"
		}
	case RelativeDurationID:
		switch typ {
		case relativeDurationType:
			return &RelativeDurationCodec{}, nil
		case optionalRelativeDurationType:
			return &optionalRelativeDurationDecoder{}, nil
		default:
			expectedType = "gel.RealtiveDuration or " +
				"gel.OptionalRelativeDuration"
		}
	case DateDurationID:
		switch typ {
		case dateDurationType:
			return &DateDurationCodec{}, nil
		case optionalDateDurationType:
			return &optionalDateDurationDecoder{}, nil
		default:
			expectedType = "gel.DateDuration or " +
				"gel.OptionalDateDuration"
		}
	case MemoryID:
		switch typ {
		case memoryType:
			return &MemoryCodec{}, nil
		case optionalMemoryType:
			return &optionalMemoryDecoder{}, nil
		default:
			expectedType = "gel.Memory or gel.OptionalMemory"
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

func buildScalarDecoderV2(
	desc *descriptor.V2,
	typ reflect.Type,
	path Path,
) (Decoder, error) {
	if desc.Type == descriptor.Scalar {
		desc = GetScalarDescriptorV2(desc)
	}

	decoder, ok, err := buildUnmarshalerV2(desc, typ)
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
			return &StrCodec{desc.ID}, nil
		case optionalStrType:
			return &optionalStrDecoder{StrID}, nil
		default:
			expectedType = "string or gel.OptionalStr"
			goto TypeMissmatch
		}
	}

	switch desc.ID {
	case UUIDID:
		switch typ {
		case uuidType:
			return &UUIDCodec{}, nil
		case optionalUUIDType:
			return &optionalUUIDDecoder{}, nil
		default:
			expectedType = "uuid or gel.OptionalUUID"
		}
	case StrID:
		switch typ {
		case strType:
			return &StrCodec{StrID}, nil
		case optionalStrType:
			return &optionalStrDecoder{StrID}, nil
		default:
			expectedType = "string or gel.OptionalStr"
		}
	case BytesID:
		switch typ {
		case bytesType:
			return &BytesCodec{BytesID}, nil
		case optionalBytesType:
			return &optionalBytesDecoder{BytesID}, nil
		default:
			expectedType = "[]byte or gel.OptionalBytes"
		}
	case Int16ID:
		switch typ {
		case int16Type:
			return &Int16Codec{}, nil
		case optionalInt16Type:
			return &optionalInt16Decoder{}, nil
		default:
			expectedType = "int16 or gel.OptionalInt16"
		}
	case Int32ID:
		switch typ {
		case int32Type:
			return &Int32Codec{}, nil
		case optionalInt32Type:
			return &optionalInt32Decoder{}, nil
		default:
			expectedType = "int32 or gel.OptionalInt32"
		}
	case Int64ID:
		switch typ {
		case int64Type:
			return &Int64Codec{}, nil
		case optionalInt64Type:
			return &optionalInt64Decoder{}, nil
		default:
			expectedType = "int64 or gel.OptionalInt64"
		}
	case Float32ID:
		switch typ {
		case float32Type:
			return &Float32Codec{}, nil
		case optionalFloat32Type:
			return &optionalFloat32Decoder{}, nil
		default:
			expectedType = "float32 or gel.OptionalFloat32"
		}
	case Float64ID:
		switch typ {
		case float64Type:
			return &Float64Codec{}, nil
		case optionalFloat64Type:
			return &optionalFloat64Decoder{}, nil
		default:
			expectedType = "float64 or gel.OptionalFloat64"
		}
	case DecimalID:
		return nil, errors.New("decimal codec not implemented. " +
			"Consider implementing your own gel.DecimalMarshaler " +
			"and gel.DecimalUnmarshaler.")
	case BoolID:
		switch typ {
		case boolType:
			return &BoolCodec{}, nil
		case optionalBoolType:
			return &optionalBoolDecoder{}, nil
		default:
			expectedType = "bool or gel.OptionalBool"
		}
	case DateTimeID:
		switch typ {
		case dateTimeType:
			return &DateTimeCodec{}, nil
		case optionalDateTimeType:
			return &optionalDateTimeDecoder{}, nil
		default:
			expectedType = "gel.DateTime or gel.OptionalDateTime"
		}
	case LocalDTID:
		switch typ {
		case localDateTimeType:
			return &LocalDateTimeCodec{}, nil
		case optionalLocalDateTimeType:
			return &optionalLocalDateTimeDecoder{}, nil
		default:
			expectedType = "gel.LocalDateTime or " +
				"gel.OptionalLocalDateTime"
		}
	case LocalDateID:
		switch typ {
		case localDateType:
			return &LocalDateCodec{}, nil
		case optionalLocalDateType:
			return &optionalLocalDateDecoder{}, nil
		default:
			expectedType = "gel.LocalDate or gel.OptionalLocalDate"
		}
	case LocalTimeID:
		switch typ {
		case localTimeType:
			return &LocalTimeCodec{}, nil
		case optionalLocalTimeType:
			return &optionalLocalTimeDecoder{}, nil
		default:
			expectedType = "gel.LocalTime or gel.OptionalLocalTime"
		}
	case DurationID:
		switch typ {
		case durationType:
			return &DurationCodec{}, nil
		case optionalDurationType:
			return &optionalDurationDecoder{}, nil
		default:
			expectedType = "gel.Duration or gel.OptionalDuration"
		}
	case JSONID:
		ptr := reflect.PointerTo(typ)

		switch {
		case typ == bytesType:
			return &JSONCodec{typ: typ}, nil
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
			return &JSONCodec{typ: typ}, nil
		}
	case BigIntID:
		switch typ {
		case bigIntType:
			return &BigIntCodec{}, nil
		case optionalBigIntType:
			return &optionalBigIntDecoder{}, nil
		default:
			expectedType = "*big.Int or gel.OptionalBigInt"
		}
	case RelativeDurationID:
		switch typ {
		case relativeDurationType:
			return &RelativeDurationCodec{}, nil
		case optionalRelativeDurationType:
			return &optionalRelativeDurationDecoder{}, nil
		default:
			expectedType = "gel.RealtiveDuration or " +
				"gel.OptionalRelativeDuration"
		}
	case DateDurationID:
		switch typ {
		case dateDurationType:
			return &DateDurationCodec{}, nil
		case optionalDateDurationType:
			return &optionalDateDurationDecoder{}, nil
		default:
			expectedType = "gel.DateDuration or " +
				"gel.OptionalDateDuration"
		}
	case MemoryID:
		switch typ {
		case memoryType:
			return &MemoryCodec{}, nil
		case optionalMemoryType:
			return &optionalMemoryDecoder{}, nil
		default:
			expectedType = "gel.Memory or gel.OptionalMemory"
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
