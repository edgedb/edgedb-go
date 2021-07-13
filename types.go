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

	// Optional ...
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
)
