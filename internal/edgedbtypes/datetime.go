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

package edgedbtypes

import (
	"time"
)

// timeShift is the number of microseconds
// between 0001-01-01T00:00 and 2000-01-01T00:00
const timeShift = 62_135_596_800_000_000

// NewLocalDateTime returns a new LocalDateTime
func NewLocalDateTime(
	year int, month time.Month, day, hour, minute, second, microsecond int,
) LocalDateTime {
	t := time.Date(
		year, month, day, hour, minute, second, microsecond*1_000, time.UTC,
	)
	sec := t.Unix()
	nsec := int64(t.Sub(time.Unix(sec, 0)))
	return LocalDateTime{sec*1_000_000 + nsec/1_000 + timeShift}
}

// LocalDateTime is a date and time without timezone.
// https://www.edgedb.com/docs/datamodel/scalars/datetime/
type LocalDateTime struct {
	usec int64
}

func (dt LocalDateTime) String() string {
	usec := dt.usec - timeShift
	sec := usec / 1_000_000
	nsec := (usec % 1_000_000) * 1_000
	return time.Unix(sec, nsec).UTC().Format("2006-01-02T15:04:05.999999")
}
