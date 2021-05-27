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
