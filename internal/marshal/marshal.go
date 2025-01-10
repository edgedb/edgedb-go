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

// Package marshal documents marshaling interfaces.
//
// User defined marshaler/unmarshalers can be defined for any scalar Gel
// type except arrays. They must implement the interface for their type.
// For example a custom int64 unmarshaler should implement Int64Unmarshaler.
//
// # Optional Fields
//
// When shape fields in a query result are optional (not required) the client
// requires the out value's optional fields to implement OptionalUnmarshaler.
// For scalar types, this means that the field value will need to implement a
// custom marshaler interface i.e. Int64Unmarshaler AND OptionalUnmarshaler.
// For shapes, only OptionalUnmarshaler needs to be implemented.
package marshal

// OptionalUnmarshaler is used for optional (not required) shape field values.
type OptionalUnmarshaler interface {
	// SetMissing is call with true when the value is missing and false when
	// the value is present.
	SetMissing(bool)
}

// OptionalScalarUnmarshaler is implemented by optional scalar types.
type OptionalScalarUnmarshaler interface {
	Unset()
}

// OptionalMarshaler is used for optional (not required) shape field values.
type OptionalMarshaler interface {
	// Missing returns true when the value is missing.
	Missing() bool
}

// StrMarshaler is the interface implemented by an object
// that can marshal itself into the str wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-str
//
// MarshalEdgeDBStr encodes the receiver
// into a binary form and returns the result.
type StrMarshaler interface {
	MarshalEdgeDBStr() ([]byte, error)
}

// StrUnmarshaler is the interface implemented by an object
// that can unmarshal the str wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-str
//
// UnmarshalEdgeDBStr must be able to decode the str wire format.
// UnmarshalEdgeDBStr must copy the data if it wishes to retain the data
// after returning.
type StrUnmarshaler interface {
	UnmarshalEdgeDBStr(data []byte) error
}

// BoolMarshaler is the interface implemented by an object
// that can marshal itself into the bool wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bool
//
// MarshalEdgeDBBool encodes the receiver
// into a binary form and returns the result.
type BoolMarshaler interface {
	MarshalEdgeDBBool() ([]byte, error)
}

// BoolUnmarshaler is the interface implemented by an object
// that can unmarshal the bool wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bool
//
// UnmarshalEdgeDBBool must be able to decode the bool wire format.
// UnmarshalEdgeDBBool must copy the data if it wishes to retain the data
// after returning.
type BoolUnmarshaler interface {
	UnmarshalEdgeDBBool(data []byte) error
}

// JSONMarshaler is the interface implemented by an object
// that can marshal itself into the json wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-json
//
// MarshalEdgeDBJSON encodes the receiver
// into a binary form and returns the result.
type JSONMarshaler interface {
	MarshalEdgeDBJSON() ([]byte, error)
}

// JSONUnmarshaler is the interface implemented by an object
// that can unmarshal the json wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-json
//
// UnmarshalEdgeDBJSON must be able to decode the json wire format.
// UnmarshalEdgeDBJSON must copy the data if it wishes to retain the data
// after returning.
type JSONUnmarshaler interface {
	UnmarshalEdgeDBJSON(data []byte) error
}

// UUIDMarshaler is the interface implemented by an object
// that can marshal itself into the uuid wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-uuid
//
// MarshalEdgeDBUUID encodes the receiver
// into a binary form and returns the result.
type UUIDMarshaler interface {
	MarshalEdgeDBUUID() ([]byte, error)
}

// UUIDUnmarshaler is the interface implemented by an object
// that can unmarshal the uuid wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-uuid
//
// UnmarshalEdgeDBUUID must be able to decode the uuid wire format.
// UnmarshalEdgeDBUUID must copy the data if it wishes to retain the data
// after returning.
type UUIDUnmarshaler interface {
	UnmarshalEdgeDBUUID(data []byte) error
}

// BytesMarshaler is the interface implemented by an object
// that can marshal itself into the bytes wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bytes
//
// MarshalEdgeDBBytes encodes the receiver
// into a binary form and returns the result.
type BytesMarshaler interface {
	MarshalEdgeDBBytes() ([]byte, error)
}

// BytesUnmarshaler is the interface implemented by an object
// that can unmarshal the bytes wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bytes
//
// UnmarshalEdgeDBBytes must be able to decode the bytes wire format.
// UnmarshalEdgeDBBytes must copy the data if it wishes to retain the data
// after returning.
type BytesUnmarshaler interface {
	UnmarshalEdgeDBBytes(data []byte) error
}

// BigIntMarshaler is the interface implemented by an object
// that can marshal itself into the bigint wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bigint
//
// MarshalEdgeDBBigInt encodes the receiver
// into a binary form and returns the result.
type BigIntMarshaler interface {
	MarshalEdgeDBBigInt() ([]byte, error)
}

// BigIntUnmarshaler is the interface implemented by an object
// that can unmarshal the bigint wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bigint
//
// UnmarshalEdgeDBBigInt must be able to decode the bigint wire format.
// UnmarshalEdgeDBBigInt must copy the data if it wishes to retain the data
// after returning.
type BigIntUnmarshaler interface {
	UnmarshalEdgeDBBigInt(data []byte) error
}

// DecimalMarshaler is the interface implemented by an object
// that can marshal itself into the decimal wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-decimal
//
// MarshalEdgeDBDecimal encodes the receiver
// into a binary form and returns the result.
type DecimalMarshaler interface {
	MarshalEdgeDBDecimal() ([]byte, error)
}

// DecimalUnmarshaler is the interface implemented by an object
// that can unmarshal the decimal wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-decimal
//
// UnmarshalEdgeDBDecimal must be able to decode the decimal wire format.
// UnmarshalEdgeDBDecimal must copy the data if it wishes to retain the data
// after returning.
type DecimalUnmarshaler interface {
	UnmarshalEdgeDBDecimal(data []byte) error
}

// DateTimeMarshaler is the interface implemented by an object
// that can marshal itself into the datetime wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-datetime
//
// MarshalEdgeDBDateTime encodes the receiver
// into a binary form and returns the result.
type DateTimeMarshaler interface {
	MarshalEdgeDBDateTime() ([]byte, error)
}

// DateTimeUnmarshaler is the interface implemented by an object
// that can unmarshal the datetime wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-datetime
//
// UnmarshalEdgeDBDateTime must be able to decode the datetime wire format.
// UnmarshalEdgeDBDateTime must copy the data if it wishes to retain the data
// after returning.
type DateTimeUnmarshaler interface {
	UnmarshalEdgeDBDateTime(data []byte) error
}

// LocalDateTimeMarshaler is the interface implemented by an object
// that can marshal itself into the local_datetime wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats
//
// MarshalEdgeDBLocalDateTime encodes the receiver
// into a binary form and returns the result.
type LocalDateTimeMarshaler interface {
	MarshalEdgeDBLocalDateTime() ([]byte, error)
}

// LocalDateTimeUnmarshaler is the interface implemented by an object
// that can unmarshal the local_datetime wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats
//
// UnmarshalEdgeDBLocalDateTime must be able to decode the local_datetime wire
// format. UnmarshalEdgeDBLocalDateTime must copy the data if it wishes to
// retain the data after returning.
type LocalDateTimeUnmarshaler interface {
	UnmarshalEdgeDBLocalDateTime(data []byte) error
}

// LocalDateMarshaler is the interface implemented by an object
// that can marshal itself into the local_date wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-local-date
//
// MarshalEdgeDBLocalDate encodes the receiver
// into a binary form and returns the result.
type LocalDateMarshaler interface {
	MarshalEdgeDBLocalDate() ([]byte, error)
}

// LocalDateUnmarshaler is the interface implemented by an object
// that can unmarshal the local_date wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-local-date
//
// UnmarshalEdgeDBLocalDate must be able to decode the local_date wire format.
// UnmarshalEdgeDBLocalDate must copy the data if it wishes to retain the data
// after returning.
type LocalDateUnmarshaler interface {
	UnmarshalEdgeDBLocalDate(data []byte) error
}

// LocalTimeMarshaler is the interface implemented by an object
// that can marshal itself into the local_time wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-local-time
//
// MarshalEdgeDBLocalTime encodes the receiver
// into a binary form and returns the result.
type LocalTimeMarshaler interface {
	MarshalEdgeDBLocalTime() ([]byte, error)
}

// LocalTimeUnmarshaler is the interface implemented by an object
// that can unmarshal the local_time wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-local-time
//
// UnmarshalEdgeDBLocalTime must be able to decode the local_time wire format.
// UnmarshalEdgeDBLocalTime must copy the data if it wishes to retain the data
// after returning.
type LocalTimeUnmarshaler interface {
	UnmarshalEdgeDBLocalTime(data []byte) error
}

// DurationMarshaler is the interface implemented by an object
// that can marshal itself into the duration wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-duration
//
// MarshalEdgeDBDuration encodes the receiver
// into a binary form and returns the result.
type DurationMarshaler interface {
	MarshalEdgeDBDuration() ([]byte, error)
}

// DurationUnmarshaler is the interface implemented by an object
// that can unmarshal the duration wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-duration
//
// UnmarshalEdgeDBDuration must be able to decode the duration wire format.
// UnmarshalEdgeDBDuration must copy the data if it wishes to retain the data
// after returning.
type DurationUnmarshaler interface {
	UnmarshalEdgeDBDuration(data []byte) error
}

// RelativeDurationMarshaler is the interface implemented by an object that can
// marshal itself into the cal::relative_duration wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats
//
// MarshalEdgeDBRelativeDuration encodes the receiver into a binary form and
// returns the result.
type RelativeDurationMarshaler interface {
	MarshalEdgeDBRelativeDuration() ([]byte, error)
}

// RelativeDurationUnmarshaler is the interface implemented by an object that
// can unmarshal the cal::relative_duration wire format representation of
// itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-duration
//
// UnmarshalEdgeDBRelativeDuration must be able to decode the
// cal::relative_duration wire format.  UnmarshalEdgeDBRelativeDuration must
// copy the data if it wishes to retain the data after returning.
type RelativeDurationUnmarshaler interface {
	UnmarshalEdgeDBRelativeDuration(data []byte) error
}

// DateDurationMarshaler is the interface implemented by an object that can
// marshal itself into the cal::relative_duration wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats
//
// MarshalEdgeDBDateDuration encodes the receiver into a binary form and
// returns the result.
type DateDurationMarshaler interface {
	MarshalEdgeDBDateDuration() ([]byte, error)
}

// DateDurationUnmarshaler is the interface implemented by an object that
// can unmarshal the cal::relative_duration wire format representation of
// itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-duration
//
// UnmarshalEdgeDBDateDuration must be able to decode the
// cal::relative_duration wire format.  UnmarshalEdgeDBDateDuration must
// copy the data if it wishes to retain the data after returning.
type DateDurationUnmarshaler interface {
	UnmarshalEdgeDBDateDuration(data []byte) error
}

// Int16Marshaler is the interface implemented by an object
// that can marshal itself into the int16 wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int16
//
// MarshalEdgeDBInt16 encodes the receiver
// into a binary form and returns the result.
type Int16Marshaler interface {
	MarshalEdgeDBInt16() ([]byte, error)
}

// Int16Unmarshaler is the interface implemented by an object
// that can unmarshal the int16 wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int16
//
// UnmarshalEdgeDBInt16 must be able to decode the int16 wire format.
// UnmarshalEdgeDBInt16 must copy the data if it wishes to retain the data
// after returning.
type Int16Unmarshaler interface {
	UnmarshalEdgeDBInt16(data []byte) error
}

// Int32Marshaler is the interface implemented by an object
// that can marshal itself into the int32 wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int32
//
// MarshalEdgeDBInt32 encodes the receiver
// into a binary form and returns the result.
type Int32Marshaler interface {
	MarshalEdgeDBInt32() ([]byte, error)
}

// Int32Unmarshaler is the interface implemented by an object
// that can unmarshal the int32 wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int32
//
// UnmarshalEdgeDBInt32 must be able to decode the int32 wire format.
// UnmarshalEdgeDBInt32 must copy the data if it wishes to retain the data
// after returning.
type Int32Unmarshaler interface {
	UnmarshalEdgeDBInt32(data []byte) error
}

// Int64Marshaler is the interface implemented by an object
// that can marshal itself into the int64 wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int64
//
// MarshalEdgeDBInt64 encodes the receiver
// into a binary form and returns the result.
type Int64Marshaler interface {
	MarshalEdgeDBInt64() ([]byte, error)
}

// Int64Unmarshaler is the interface implemented by an object
// that can unmarshal the int64 wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int64
//
// UnmarshalEdgeDBInt64 must be able to decode the int64 wire format.
// UnmarshalEdgeDBInt64 must copy the data if it wishes to retain the data
// after returning.
type Int64Unmarshaler interface {
	UnmarshalEdgeDBInt64(data []byte) error
}

// Float32Marshaler is the interface implemented by an object
// that can marshal itself into the float32 wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-float32
//
// MarshalEdgeDBFloat32 encodes the receiver
// into a binary form and returns the result.
type Float32Marshaler interface {
	MarshalEdgeDBFloat32() ([]byte, error)
}

// Float32Unmarshaler is the interface implemented by an object
// that can unmarshal the float32 wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-float32
//
// UnmarshalEdgeDBFloat32 must be able to decode the float32 wire format.
// UnmarshalEdgeDBFloat32 must copy the data if it wishes to retain the data
// after returning.
type Float32Unmarshaler interface {
	UnmarshalEdgeDBFloat32(data []byte) error
}

// Float64Marshaler is the interface implemented by an object
// that can marshal itself into the float64 wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-float64
//
// MarshalEdgeDBFloat64 encodes the receiver
// into a binary form and returns the result.
type Float64Marshaler interface {
	MarshalEdgeDBFloat64() ([]byte, error)
}

// Float64Unmarshaler is the interface implemented by an object
// that can unmarshal the float64 wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-float64
//
// UnmarshalEdgeDBFloat64 must be able to decode the float64 wire format.
// UnmarshalEdgeDBFloat64 must copy the data if it wishes to retain the data
// after returning.
type Float64Unmarshaler interface {
	UnmarshalEdgeDBFloat64(data []byte) error
}

// MemoryMarshaler is the interface implemented by an object
// that can marshal itself into the memory wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-memory
//
// MarshalEdgeDBMemory encodes the receiver
// into a binary form and returns the result.
type MemoryMarshaler interface {
	MarshalEdgeDBMemory() ([]byte, error)
}

// MemoryUnmarshaler is the interface implemented by an object
// that can unmarshal the memory wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-memory
//
// UnmarshalEdgeDBMemory must be able to decode the memory wire format.
// UnmarshalEdgeDBMemory must copy the data if it wishes to retain the data
// after returning.
type MemoryUnmarshaler interface {
	UnmarshalEdgeDBMemory(data []byte) error
}
