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
	"math/big"
	"reflect"
	"time"

	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
	"github.com/edgedb/edgedb-go/internal/marshal"
)

var (
	relativeDurationID = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x11}
	uuidID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}
	// StrID is the str type descriptor ID
	StrID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}
	bytesID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2}
	int16ID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3}
	int32ID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 4}
	// Int64ID is the int64 type descriptor ID
	Int64ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 5}
	float32ID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 6}
	float64ID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 7}
	decimalID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 8}
	// BoolID is the bool type descriptor ID
	BoolID      = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 9}
	dateTimeID  = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x0a}
	localDTID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x0b}
	localDateID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x0c}
	localTimeID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x0d}
	// DurationID is the duration type descriptor ID
	DurationID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x0e}
	jsonID     = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x0f}
	bigIntID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x10}
	// MemoryID is the cfg::memory type descriptor ID
	MemoryID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x30}

	int16Type                 = reflect.TypeOf(int16(0))
	int32Type                 = reflect.TypeOf(int32(0))
	int64Type                 = reflect.TypeOf(int64(0))
	float32Type               = reflect.TypeOf(float32(0))
	float64Type               = reflect.TypeOf(float64(0))
	optionalInt16Type         = reflect.TypeOf(types.OptionalInt16{})
	optionalInt32Type         = reflect.TypeOf(types.OptionalInt32{})
	optionalInt64Type         = reflect.TypeOf(types.OptionalInt64{})
	optionalFloat32Type       = reflect.TypeOf(types.OptionalFloat32{})
	optionalFloat64Type       = reflect.TypeOf(types.OptionalFloat64{})
	strType                   = reflect.TypeOf("")
	optionalStrType           = reflect.TypeOf(types.OptionalStr{})
	boolType                  = reflect.TypeOf(false)
	optionalBoolType          = reflect.TypeOf(types.OptionalBool{})
	uuidType                  = reflect.TypeOf(uuidID)
	optionalUUIDType          = reflect.TypeOf(types.OptionalUUID{})
	bytesType                 = reflect.TypeOf([]byte{})
	optionalBytesType         = reflect.TypeOf(types.OptionalBytes{})
	dateTimeType              = reflect.TypeOf(time.Time{})
	localDateTimeType         = reflect.TypeOf(types.LocalDateTime{})
	localDateType             = reflect.TypeOf(types.LocalDate{})
	localTimeType             = reflect.TypeOf(types.LocalTime{})
	durationType              = reflect.TypeOf(types.Duration(0))
	relativeDurationType      = reflect.TypeOf(types.RelativeDuration{})
	bigIntType                = reflect.TypeOf(&big.Int{})
	memoryType                = reflect.TypeOf(types.Memory(0))
	optionalBigIntType        = reflect.TypeOf(types.OptionalBigInt{})
	optionalDateTimeType      = reflect.TypeOf(types.OptionalDateTime{})
	optionalLocalDateTimeType = reflect.TypeOf(
		types.OptionalLocalDateTime{})
	optionalLocalDateType        = reflect.TypeOf(types.OptionalLocalDate{})
	optionalLocalTimeType        = reflect.TypeOf(types.OptionalLocalTime{})
	optionalDurationType         = reflect.TypeOf(types.OptionalDuration{})
	optionalRelativeDurationType = reflect.TypeOf(
		types.OptionalRelativeDuration{})
	optionalMemoryType      = reflect.TypeOf(types.OptionalMemory{})
	optionalUnmarshalerType = getType(
		(*marshal.OptionalUnmarshaler)(nil))
	optionalScalarUnmarshalerType = getType(
		(*marshal.OptionalScalarUnmarshaler)(nil))

	big10k  = big.NewInt(10_000)
	bigOne  = big.NewInt(1)
	bigZero = big.NewInt(0)

	// JSONBytes is a special case codec for json queries.
	// In go query json should return bytes not str.
	// but the descriptor type ID sent to the server
	// should still be str.
	JSONBytes = &bytesCodec{StrID}

	trueValue  = reflect.ValueOf(true)
	falseValue = reflect.ValueOf(false)
)
