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
	"fmt"
	"strings"
	"time"
)

const (
	// timeShift is the number of seconds
	// between 0001-01-01T00:00 and 1970-01-01T00:00
	timeShift = 62135596800
)

// NewLocalDateTime returns a new LocalDateTime
func NewLocalDateTime(
	year int, month time.Month, day, hour, minute, second, microsecond int,
) LocalDateTime {
	t := time.Date(
		year, month, day, hour, minute, second, microsecond*1_000, time.UTC,
	)
	sec := t.Unix() + timeShift
	nsec := int64(t.Sub(time.Unix(t.Unix(), 0)))
	return LocalDateTime{sec*1_000_000 + nsec/1_000}
}

// LocalDateTime is a date and time without timezone.
// https://www.edgedb.com/docs/datamodel/scalars/datetime/
type LocalDateTime struct {
	usec int64
}

func (dt LocalDateTime) String() string {
	sec := dt.usec/1_000_000 - timeShift
	nsec := (dt.usec % 1_000_000) * 1_000
	return time.Unix(sec, nsec).UTC().Format("2006-01-02T15:04:05.999999")
}

// NewLocalDate returns a new LocalDate
func NewLocalDate(year int, month time.Month, day int) LocalDate {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return LocalDate{int32((t.Unix() + timeShift) / 86400)}
}

// LocalDate is a date without a time zone.
// https://www.edgedb.com/docs/datamodel/scalars/datetime/
type LocalDate struct {
	days int32
}

func (d LocalDate) String() string {
	return time.Unix(
		int64(d.days)*86400-timeShift,
		0,
	).UTC().Format("2006-01-02")
}

// NewLocalTime returns a new LocalTime
func NewLocalTime(hour, minute, second, microsecond int) LocalTime {
	if hour < 0 || hour > 23 {
		panic("hour out of range 0-23")
	}

	if minute < 0 || minute > 59 {
		panic("minute out of range 0-59")
	}

	if second < 0 || second > 59 {
		panic("second out of range 0-59")
	}

	if microsecond < 0 || microsecond > 999_999 {
		panic("microsecond out of range 0-999_999")
	}

	t := time.Date(
		1970, 1, 1, hour, minute, second, microsecond*1_000, time.UTC,
	)
	return LocalTime{t.UnixNano() / 1_000}
}

// LocalTime is a time without a time zone.
// https://www.edgedb.com/docs/datamodel/scalars/datetime/
type LocalTime struct {
	usec int64
}

func (t LocalTime) String() string {
	return time.Unix(
		t.usec/1_000_000,
		(t.usec%1_000_000)*1_000,
	).UTC().Format("15:04:05.999999")
}

// Duration represents the elapsed time between two instants
// as an int64 microsecond count.
type Duration int64

func (d Duration) String() string {
	if d == 0 {
		return "0s"
	}

	x := int64(d)
	neg := ""
	if x < 0 {
		x = -x
		neg = "-"
	}

	if x < 1_000 {
		return fmt.Sprintf("%v%vus", neg, x)
	}

	us := x % 1_000
	ms := (x % 1_000_000) / 1_000
	if x < 1_000_000 {
		return fmt.Sprintf("%v%vms", neg, fmtFloat(ms, us*1_000))
	}

	ms = x % 1_000_000
	sec := (x % 60000000) / 1_000_000
	str := fmt.Sprintf("%vs", fmtFloat(sec, ms))
	str = strings.TrimLeft(str, "0s")

	min := (x % 3600000000) / 60000000
	str = fmt.Sprintf("%vm%v", min, str)
	str = strings.TrimLeft(str, "0m")

	hrs := x / 3600000000
	str = fmt.Sprintf("%vh%v", hrs, str)
	str = strings.TrimLeft(str, "0h")
	return neg + str
}

func fmtFloat(x int64, y int64) string {
	decimal := strings.TrimRight(fmt.Sprintf(".%06d", y), ".0")
	return fmt.Sprintf("%v%v", x, decimal)
}
