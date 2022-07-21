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
	"math/big"
	"time"

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

	// Duration represents a span of time.
	Duration = edgedbtypes.Duration

	// RelativeDuration represents a fuzzy/human span of time.
	RelativeDuration = edgedbtypes.RelativeDuration

	// DateDuration represents a fuzzy/human span of time in days and months.
	DateDuration = edgedbtypes.DateDuration

	// Memory represents memory in bytes
	Memory = edgedbtypes.Memory

	// Optional is embedded in structs to make them optional. For example:
	//   type User struct {
	//       edgedb.Optional
	//       Name string `edgedb:"name"`
	//   }
	Optional = edgedbtypes.Optional

	// OptionalBool is a bool value that is not required.
	OptionalBool = edgedbtypes.OptionalBool

	// OptionalBytes is a []byte value that is not required.
	OptionalBytes = edgedbtypes.OptionalBytes

	// OptionalStr is a string that is not required.
	OptionalStr = edgedbtypes.OptionalStr

	// OptionalInt16 is an int16 that is not required.
	OptionalInt16 = edgedbtypes.OptionalInt16

	// OptionalInt32 is an int32 that is not required.
	OptionalInt32 = edgedbtypes.OptionalInt32

	// OptionalInt64 is an int64 that is not required.
	OptionalInt64 = edgedbtypes.OptionalInt64

	// OptionalFloat32  is a float32 that is not required.
	OptionalFloat32 = edgedbtypes.OptionalFloat32

	// OptionalFloat64 is a float64 that is not required.
	OptionalFloat64 = edgedbtypes.OptionalFloat64

	// OptionalBigInt is a big.Int that is not required.
	OptionalBigInt = edgedbtypes.OptionalBigInt

	// OptionalUUID is a UUID that is not required.
	OptionalUUID = edgedbtypes.OptionalUUID

	// OptionalDateTime is a time.Time that is not required.
	OptionalDateTime = edgedbtypes.OptionalDateTime

	// OptionalLocalDateTime is a LocalDateTime that is not required.
	OptionalLocalDateTime = edgedbtypes.OptionalLocalDateTime

	// OptionalLocalTime is a LocalTime that is not required.
	OptionalLocalTime = edgedbtypes.OptionalLocalTime

	// OptionalLocalDate is a LocalDate that is not required.
	OptionalLocalDate = edgedbtypes.OptionalLocalDate

	// OptionalDuration is a Duration that is not required.
	OptionalDuration = edgedbtypes.OptionalDuration

	// OptionalRelativeDuration is a RelativeDuration that is not required.
	OptionalRelativeDuration = edgedbtypes.OptionalRelativeDuration

	// OptionalDateDuration is a DateDuration that is not required.
	OptionalDateDuration = edgedbtypes.OptionalDateDuration

	// OptionalMemory is a Memory that is not required.
	OptionalMemory = edgedbtypes.OptionalMemory

	// RangeInt32 is an interval of Int32s
	RangeInt32 = edgedbtypes.RangeInt32

	// OptionalRangeInt32 is a RangeInt32 that is not required.
	OptionalRangeInt32 = edgedbtypes.OptionalRangeInt32

	// RangeInt64 is an interval of Int64s
	RangeInt64 = edgedbtypes.RangeInt64

	// OptionalRangeInt64 is a RangeInt64 that is not required.
	OptionalRangeInt64 = edgedbtypes.OptionalRangeInt64

	// RangeFloat32 is an interval of Float32s
	RangeFloat32 = edgedbtypes.RangeFloat32

	// OptionalRangeFloat32 is a RangeFloat32 that is not required.
	OptionalRangeFloat32 = edgedbtypes.OptionalRangeFloat32

	// RangeFloat64 is an interval of Float64s
	RangeFloat64 = edgedbtypes.RangeFloat64

	// OptionalRangeFloat64 is a RangeFloat64 that is not required.
	OptionalRangeFloat64 = edgedbtypes.OptionalRangeFloat64

	// RangeDateTime is an interval of DateTimes
	RangeDateTime = edgedbtypes.RangeDateTime

	// OptionalRangeDateTime is a RangeDateTime that is not required.
	OptionalRangeDateTime = edgedbtypes.OptionalRangeDateTime

	// RangeLocalDateTime is an interval of LocalDateTimes
	RangeLocalDateTime = edgedbtypes.RangeLocalDateTime

	// OptionalRangeLocalDateTime is a RangeLocalDateTime that is not required.
	OptionalRangeLocalDateTime = edgedbtypes.OptionalRangeLocalDateTime

	// RangeLocalDate is an interval of LocalDates
	RangeLocalDate = edgedbtypes.RangeLocalDate

	// OptionalRangeLocalDate is a RangeLocalDate that is not required.
	OptionalRangeLocalDate = edgedbtypes.OptionalRangeLocalDate
)

var (
	// ParseUUID parses s into a UUID or returns an error.
	ParseUUID = edgedbtypes.ParseUUID

	// NewLocalDateTime returns a new LocalDateTime
	NewLocalDateTime = edgedbtypes.NewLocalDateTime

	// NewLocalDate returns a new LocalDate
	NewLocalDate = edgedbtypes.NewLocalDate

	// NewLocalTime returns a new LocalTime
	NewLocalTime = edgedbtypes.NewLocalTime

	// NewRelativeDuration returns a new RelativeDuration
	NewRelativeDuration = edgedbtypes.NewRelativeDuration

	// NewDateDuration returns a new DateDuration
	NewDateDuration = edgedbtypes.NewDateDuration

	// NewRangeInt32 returns a new RangeInt32
	NewRangeInt32 = edgedbtypes.NewRangeInt32

	// NewRangeInt64 returns a new RangeInt64
	NewRangeInt64 = edgedbtypes.NewRangeInt64

	// NewRangeFloat32 returns a new RangeFloat32
	NewRangeFloat32 = edgedbtypes.NewRangeFloat32

	// NewRangeFloat64 returns a new RangeFloat64
	NewRangeFloat64 = edgedbtypes.NewRangeFloat64

	// NewRangeDateTime returns a new RangeDateTime
	NewRangeDateTime = edgedbtypes.NewRangeDateTime

	// NewRangeLocalDateTime returns a new RangeLocalDateTime
	NewRangeLocalDateTime = edgedbtypes.NewRangeLocalDateTime

	// NewRangeLocalDate returns a new RangeLocalDate
	NewRangeLocalDate = edgedbtypes.NewRangeLocalDate
)

// NewOptionalBool is a convenience function for creating an OptionalBool with
// its value set to v.
func NewOptionalBool(v bool) OptionalBool {
	o := OptionalBool{}
	o.Set(v)
	return o
}

// NewOptionalBytes is a convenience function for creating an OptionalBytes
// with its value set to v.
func NewOptionalBytes(v []byte) OptionalBytes {
	o := OptionalBytes{}
	o.Set(v)
	return o
}

// NewOptionalStr is a convenience function for creating an OptionalStr with
// its value set to v.
func NewOptionalStr(v string) OptionalStr {
	o := OptionalStr{}
	o.Set(v)
	return o
}

// NewOptionalInt16 is a convenience function for creating an OptionalInt16
// with its value set to v.
func NewOptionalInt16(v int16) OptionalInt16 {
	o := OptionalInt16{}
	o.Set(v)
	return o
}

// NewOptionalInt32 is a convenience function for creating an OptionalInt32
// with its value set to v.
func NewOptionalInt32(v int32) OptionalInt32 {
	o := OptionalInt32{}
	o.Set(v)
	return o
}

// NewOptionalInt64 is a convenience function for creating an OptionalInt64
// with its value set to v.
func NewOptionalInt64(v int64) OptionalInt64 {
	o := OptionalInt64{}
	o.Set(v)
	return o
}

// NewOptionalFloat32 is a convenience function for creating an OptionalFloat32
// with its value set to v.
func NewOptionalFloat32(v float32) OptionalFloat32 {
	o := OptionalFloat32{}
	o.Set(v)
	return o
}

// NewOptionalFloat64 is a convenience function for creating an OptionalFloat64
// with its value set to v.
func NewOptionalFloat64(v float64) OptionalFloat64 {
	o := OptionalFloat64{}
	o.Set(v)
	return o
}

// NewOptionalBigInt is a convenience function for creating an OptionalBigInt
// with its value set to v.
func NewOptionalBigInt(v *big.Int) OptionalBigInt {
	o := OptionalBigInt{}
	o.Set(v)
	return o
}

// NewOptionalUUID is a convenience function for creating an OptionalUUID with
// its value set to v.
func NewOptionalUUID(v UUID) OptionalUUID {
	o := OptionalUUID{}
	o.Set(v)
	return o
}

// NewOptionalDateTime is a convenience function for creating an
// OptionalDateTime with its value set to v.
func NewOptionalDateTime(v time.Time) OptionalDateTime {
	o := OptionalDateTime{}
	o.Set(v)
	return o
}

// NewOptionalLocalDateTime is a convenience function for creating an
// OptionalLocalDateTime with its value set to v.
func NewOptionalLocalDateTime(v LocalDateTime) OptionalLocalDateTime {
	o := OptionalLocalDateTime{}
	o.Set(v)
	return o
}

// NewOptionalLocalTime is a convenience function for creating an
// OptionalLocalTime with its value set to v.
func NewOptionalLocalTime(v LocalTime) OptionalLocalTime {
	o := OptionalLocalTime{}
	o.Set(v)
	return o
}

// NewOptionalLocalDate is a convenience function for creating an
// OptionalLocalDate with its value set to v.
func NewOptionalLocalDate(v LocalDate) OptionalLocalDate {
	o := OptionalLocalDate{}
	o.Set(v)
	return o
}

// NewOptionalDuration is a convenience function for creating an
// OptionalDuration with its value set to v.
func NewOptionalDuration(v Duration) OptionalDuration {
	o := OptionalDuration{}
	o.Set(v)
	return o
}

// NewOptionalRelativeDuration is a convenience function for creating an
// OptionalRelativeDuration with its value set to v.
func NewOptionalRelativeDuration(v RelativeDuration) OptionalRelativeDuration {
	o := OptionalRelativeDuration{}
	o.Set(v)
	return o
}

// NewOptionalDateDuration is a convenience function for creating an
// OptionalDateDuration with its value set to v.
func NewOptionalDateDuration(v DateDuration) OptionalDateDuration {
	o := OptionalDateDuration{}
	o.Set(v)
	return o
}

// NewOptionalMemory is a convenience function for creating an
// OptionalMemory with its value set to v.
func NewOptionalMemory(v Memory) OptionalMemory {
	o := OptionalMemory{}
	o.Set(v)
	return o
}

// NewOptionalRangeInt32 is a convenience function for creating an
// OptionalRangeInt32 with its value set to v.
func NewOptionalRangeInt32(v RangeInt32) OptionalRangeInt32 {
	o := OptionalRangeInt32{}
	o.Set(v)
	return o
}

// NewOptionalRangeInt64 is a convenience function for creating an
// OptionalRangeInt64 with its value set to v.
func NewOptionalRangeInt64(v RangeInt64) OptionalRangeInt64 {
	o := OptionalRangeInt64{}
	o.Set(v)
	return o
}

// NewOptionalRangeFloat32 is a convenience function for creating an
// OptionalRangeFloat32 with its value set to v.
func NewOptionalRangeFloat32(v RangeFloat32) OptionalRangeFloat32 {
	o := OptionalRangeFloat32{}
	o.Set(v)
	return o
}

// NewOptionalRangeFloat64 is a convenience function for creating an
// OptionalRangeFloat64 with its value set to v.
func NewOptionalRangeFloat64(v RangeFloat64) OptionalRangeFloat64 {
	o := OptionalRangeFloat64{}
	o.Set(v)
	return o
}

// NewOptionalRangeDateTime is a convenience function for creating an
// OptionalRangeDateTime with its value set to v.
func NewOptionalRangeDateTime(v RangeDateTime) OptionalRangeDateTime {
	o := OptionalRangeDateTime{}
	o.Set(v)
	return o
}

// NewOptionalRangeLocalDateTime is a convenience function for creating an
// OptionalRangeLocalDateTime with its value set to v.
func NewOptionalRangeLocalDateTime(
	v RangeLocalDateTime,
) OptionalRangeLocalDateTime {
	o := OptionalRangeLocalDateTime{}
	o.Set(v)
	return o
}

// NewOptionalRangeLocalDate is a convenience function for creating an
// OptionalRangeLocalDate with its value set to v.
func NewOptionalRangeLocalDate(v RangeLocalDate) OptionalRangeLocalDate {
	o := OptionalRangeLocalDate{}
	o.Set(v)
	return o
}
