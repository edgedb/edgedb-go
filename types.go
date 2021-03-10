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

package edgedb

import (
	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

type (
	// UUID a universally unique identifier
	UUID = edgedbtypes.UUID

	// LocalDateTime is a date and time without a time zone.
	LocalDateTime = edgedbtypes.LocalDateTime

	// LocalDate is a date without a time zone.
	LocalDate = edgedbtypes.LocalDate

	// LocalTime is a time without a time zone.
	LocalTime = edgedbtypes.LocalTime

	// Duration representing a span of time.
	Duration = edgedbtypes.Duration

	// BoolMarshaler enables encoding user defined bool values.
	BoolMarshaler = codecs.BoolMarshaler

	// BoolUnmarshaler enables decoding user defined bool values.
	BoolUnmarshaler = codecs.BoolUnmarshaler

	// BytesMarshaler enables encoding user defined bytes values.
	BytesMarshaler = codecs.BytesMarshaler

	// BytesUnmarshaler enables decoding user defined bytes values.
	BytesUnmarshaler = codecs.BytesUnmarshaler

	// DateTimeMarshaler enables encoding user defined datetime values.
	DateTimeMarshaler = codecs.DateTimeMarshaler

	// DateTimeUnmarshaler enables decoding user defined datetime values.
	DateTimeUnmarshaler = codecs.DateTimeUnmarshaler

	// LocalDateTimeMarshaler enables encoding user defined local_datetime
	// values.
	LocalDateTimeMarshaler = codecs.LocalDateTimeMarshaler

	// LocalDateTimeUnmarshaler enables decoding user defined local_datetime
	// values.
	LocalDateTimeUnmarshaler = codecs.LocalDateTimeUnmarshaler

	// LocalDateMarshaler enables encoding user defined local_date values.
	LocalDateMarshaler = codecs.LocalDateMarshaler

	// LocalDateUnmarshaler enables decoding user defined local_date values.
	LocalDateUnmarshaler = codecs.LocalDateUnmarshaler

	// LocalTimeMarshaler enables encoding user defined local_time values.
	LocalTimeMarshaler = codecs.LocalTimeMarshaler

	// LocalTimeUnmarshaler enables decoding user defined local_time values.
	LocalTimeUnmarshaler = codecs.LocalTimeUnmarshaler

	// DurationMarshaler enables encoding user defined duration values.
	DurationMarshaler = codecs.DurationMarshaler

	// DurationUnmarshaler enables decoding user defined duration values.
	DurationUnmarshaler = codecs.DurationUnmarshaler

	// JSONMarshaler enables encoding user defined json values.
	JSONMarshaler = codecs.JSONMarshaler

	// JSONUnmarshaler enables decoding user defined json values.
	JSONUnmarshaler = codecs.JSONUnmarshaler

	// Int16Marshaler enables encoding user defined int16 values.
	Int16Marshaler = codecs.Int16Marshaler

	// Int16Unmarshaler enables decoding user defined int16 values.
	Int16Unmarshaler = codecs.Int16Unmarshaler

	// Int32Marshaler enables encoding user defined int32 values.
	Int32Marshaler = codecs.Int32Marshaler

	// Int32Unmarshaler enables decoding user defined int32 values.
	Int32Unmarshaler = codecs.Int32Unmarshaler

	// Int64Marshaler enables encoding user defined int64 values.
	Int64Marshaler = codecs.Int64Marshaler

	// Int64Unmarshaler enables decoding user defined int64 values.
	Int64Unmarshaler = codecs.Int64Unmarshaler

	// Float32Marshaler enables encoding user defined float32 values.
	Float32Marshaler = codecs.Float32Marshaler

	// Float32Unmarshaler enables decoding user defined float32 values.
	Float32Unmarshaler = codecs.Float32Unmarshaler

	// Float64Marshaler enables encoding user defined float64 values.
	Float64Marshaler = codecs.Float64Marshaler

	// Float64Unmarshaler enables decoding user defined float64 values.
	Float64Unmarshaler = codecs.Float64Unmarshaler

	// BigIntMarshaler enables encoding user defined bigint values.
	BigIntMarshaler = codecs.BigIntMarshaler

	// BigIntUnmarshaler enables decoding user defined bigint values.
	BigIntUnmarshaler = codecs.BigIntUnmarshaler

	// DecimalMarshaler enables encoding user defined decimal values.
	DecimalMarshaler = codecs.DecimalMarshaler

	// DecimalUnmarshaler enables decoding user defined decimal values.
	DecimalUnmarshaler = codecs.DecimalUnmarshaler

	// StrMarshaler enables encoding user defined str values.
	StrMarshaler = codecs.StrMarshaler

	// StrUnmarshaler enables decoding user defined str values.
	StrUnmarshaler = codecs.StrUnmarshaler

	// UUIDMarshaler enables encoding user defined uuid values.
	UUIDMarshaler = codecs.UUIDMarshaler

	// UUIDUnmarshaler enables decoding user defined uuid values.
	UUIDUnmarshaler = codecs.UUIDUnmarshaler
)

var (
	// NewLocalDateTime returns a new LocalDateTime
	NewLocalDateTime = edgedbtypes.NewLocalDateTime

	// NewLocalDate returns a new LocalDate
	NewLocalDate = edgedbtypes.NewLocalDate

	// NewLocalTime returns a new LocalTime
	NewLocalTime = edgedbtypes.NewLocalTime
)
