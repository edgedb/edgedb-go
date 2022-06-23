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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalDateTimeString(t *testing.T) {
	samples := []struct {
		str string
		dt  LocalDateTime
	}{
		{"0001-01-01T00:00:00", LocalDateTime{0}},
		{"1970-01-01T00:00:00", LocalDateTime{62_135_596_800_000_000}},
		{"2000-01-01T00:00:00", LocalDateTime{63_082_281_600_000_000}},
		{"9999-09-09T09:09:09", LocalDateTime{315_528_080_949_000_000}},
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			assert.Equal(t, s.str, s.dt.String())
		})
	}
}

func TestNewLocalDateTime(t *testing.T) {
	samples := []struct {
		str string
		dt  LocalDateTime
	}{
		{
			"2000-01-01T00:00:00",
			NewLocalDateTime(2000, 1, 1, 0, 0, 0, 0),
		},
		{
			"1999-12-31T23:59:59.999999",
			NewLocalDateTime(1999, 12, 31, 23, 59, 59, 999999),
		},
		{
			"0001-01-01T01:01:00",
			NewLocalDateTime(1, 1, 1, 1, 1, 0, 0),
		},
		{
			"9999-09-09T09:09:09",
			NewLocalDateTime(9999, 9, 9, 9, 9, 9, 0),
		},
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			assert.Equal(t, s.str, s.dt.String())
		})
	}
}

func TestLocalDateString(t *testing.T) {
	samples := []struct {
		str string
		d   LocalDate
	}{
		{"0001-01-01", LocalDate{0}},
		{"1969-07-20", LocalDate{718997}},
		{"2000-01-01", LocalDate{730119}},
		{"9999-09-09", LocalDate{3651945}},
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			assert.Equal(t, s.str, s.d.String())
		})
	}
}

func TestNewLocalDate(t *testing.T) {
	samples := []struct {
		str string
		d   LocalDate
	}{
		{"0001-01-01", NewLocalDate(1, 1, 1)},
		{"1969-07-20", NewLocalDate(1969, 7, 20)},
		{"2000-01-01", NewLocalDate(2000, 1, 1)},
		{"9999-09-09", NewLocalDate(9999, 9, 9)},
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			assert.Equal(t, s.str, s.d.String())
		})
	}
}

func TestLocalTimeString(t *testing.T) {
	samples := []struct {
		str string
		d   LocalTime
	}{
		{"00:00:00", LocalTime{0}},
		{"00:00:00.000001", LocalTime{1}},
		{"00:00:00.00001", LocalTime{10}},
		{"00:00:00.0001", LocalTime{100}},
		{"00:00:00.001", LocalTime{1000}},
		{"00:00:00.01", LocalTime{10000}},
		{"00:00:00.1", LocalTime{100000}},
		{"00:00:00.123456", LocalTime{123456}},
		{"05:04:03", LocalTime{18_243_000_000}},
		{"20:39:57", LocalTime{74_397_000_000}},
		{"23:59:59.999999", LocalTime{86_399_999_999}},
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			assert.Equal(t, s.str, s.d.String())
		})
	}
}

func TestNewLocalTime(t *testing.T) {
	samples := []struct {
		str string
		d   LocalTime
	}{
		{"00:00:00", NewLocalTime(0, 0, 0, 0)},
		{"00:00:00.000001", NewLocalTime(0, 0, 0, 1)},
		{"00:00:00.00001", NewLocalTime(0, 0, 0, 10)},
		{"00:00:00.0001", NewLocalTime(0, 0, 0, 100)},
		{"00:00:00.001", NewLocalTime(0, 0, 0, 1000)},
		{"00:00:00.01", NewLocalTime(0, 0, 0, 10000)},
		{"00:00:00.1", NewLocalTime(0, 0, 0, 100000)},
		{"00:00:00.123456", NewLocalTime(0, 0, 0, 123456)},
		{"05:04:03", NewLocalTime(5, 4, 3, 0)},
		{"20:39:57", NewLocalTime(20, 39, 57, 0)},
		{"23:59:59.999999", NewLocalTime(23, 59, 59, 999999)},
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			assert.Equal(t, s.str, s.d.String())
		})
	}
}

func TestNewLocalTimeErrors(t *testing.T) {
	samples := []struct {
		name string
		h    int
		m    int
		s    int
		us   int
	}{
		{"negative hours", -1, 0, 0, 0},
		{"negative minutes", 0, -1, 0, 0},
		{"negative seconds", 0, 0, -1, 0},
		{"negative microseconds", 0, 0, 0, -1},
		{"overflow hours", 24, 0, 0, 0},
		{"overflow minutes", 0, 60, 0, 0},
		{"overflow seconds", 0, 0, 60, 0},
		{"overflow microseconds", 0, 0, 0, 1_000_000},
	}

	for _, s := range samples {
		t.Run(s.name, func(t *testing.T) {
			assert.Panics(t, func() {
				_ = NewLocalTime(s.h, s.m, s.s, s.us)
			})
		})
	}
}

func TestParseDuration(t *testing.T) {
	samples := []struct {
		str string
		d   Duration
	}{
		{"PT1S", Duration(time.Second / 1_000)},
		{"PT1.0S", Duration(time.Second / 1_000)},
		{"PT1.S", Duration(time.Second / 1_000)},
		{"PT.1S", Duration(time.Second / 10 / 1_000)},
		{"PT1M", Duration(time.Minute / 1_000)},
		{"PT1.0M", Duration(time.Minute / 1_000)},
		{"PT1.M", Duration(time.Minute / 1_000)},
		{"PT.1M", Duration(time.Minute / 10 / 1_000)},
		{"PT1H", Duration(time.Hour / 1_000)},
		{"PT1.0H", Duration(time.Hour / 1_000)},
		{"PT1.H", Duration(time.Hour / 1_000)},
		{"PT.1H", Duration(time.Hour / 10 / 1_000)},
		{"PT2H46M39S", Duration(9999 * time.Second / 1_000)},
		{"PT2.0H46.0M39.0S", Duration(9999 * time.Second / 1_000)},

		{"1s", Duration(time.Second / 1_000)},
		{"1.0s", Duration(time.Second / 1_000)},
		{".1s", Duration(time.Second / 10 / 1_000)},
		{"1ms", Duration(time.Millisecond / 1_000)},
		{"1.0ms", Duration(time.Millisecond / 1_000)},
		{".1ms", Duration(time.Millisecond / 10 / 1_000)},
		{"1us", Duration(time.Microsecond / 1_000)},
		{"1.0us", Duration(time.Microsecond / 1_000)},
		{".1us", Duration(time.Microsecond / 10 / 1_000)},
		{"1m", Duration(time.Minute / 1_000)},
		{"1.0m", Duration(time.Minute / 1_000)},
		{".1m", Duration(time.Minute / 10 / 1_000)},
		{"1h", Duration(time.Hour / 1_000)},
		{"1.0h", Duration(time.Hour / 1_000)},
		{".1h", Duration(time.Hour / 10 / 1_000)},
		{"2h46m39s", Duration(9999 * time.Second / 1_000)},
		{"2h 46m 39s", Duration(9999 * time.Second / 1_000)},
		{"2  h  46  m  39  s", Duration(9999 * time.Second / 1_000)},
		{"2.0h46.0m39.0s", Duration(9999 * time.Second / 1_000)},
		{"2.0h 46.0m 39.0s", Duration(9999 * time.Second / 1_000)},
		{"2.0  h  46.0  m  39.0  s", Duration(9999 * time.Second / 1_000)},

		{"1second", Duration(time.Second / 1_000)},
		{"1.0second", Duration(time.Second / 1_000)},
		{".1second", Duration(time.Second / 10 / 1_000)},
		{"1minute", Duration(time.Minute / 1_000)},
		{"1.0minute", Duration(time.Minute / 1_000)},
		{".1minute", Duration(time.Minute / 10 / 1_000)},
		{"1hour", Duration(time.Hour / 1_000)},
		{"1.0hour", Duration(time.Hour / 1_000)},
		{".1hour", Duration(time.Hour / 10 / 1_000)},
		{"2hour46minute39second", Duration(9999 * time.Second / 1_000)},
		{"2hour 46minute 39second", Duration(9999 * time.Second / 1_000)},
		{
			"2  hour  46  minute  39  second",
			Duration(9999 * time.Second / 1_000),
		},
		{"2.0hour46.0minute39.0second", Duration(9999 * time.Second / 1_000)},
		{
			"2.0hour 46.0minute 39.0second",
			Duration(9999 * time.Second / 1_000),
		},
		{
			"2.0  hour  46.0  minute  39.0  second",
			Duration(9999 * time.Second / 1_000),
		},
		{
			"39.0\tsecond 2.0  hour  46.0  minute",
			Duration(9999 * time.Second / 1_000),
		},

		{"1seconds", Duration(time.Second / 1_000)},
		{"1.0seconds", Duration(time.Second / 1_000)},
		{".1seconds", Duration(time.Second / 10 / 1_000)},
		{"1minutes", Duration(time.Minute / 1_000)},
		{"1.0minutes", Duration(time.Minute / 1_000)},
		{".1minutes", Duration(time.Minute / 10 / 1_000)},
		{"1hours", Duration(time.Hour / 1_000)},
		{"1.0hours", Duration(time.Hour / 1_000)},
		{".1hours", Duration(time.Hour / 10 / 1_000)},
		{"2hours46minutes39seconds", Duration(9999 * time.Second / 1_000)},
		{"2hours 46minutes 39seconds", Duration(9999 * time.Second / 1_000)},
		{
			"2  hours  46  minutes  39  seconds",
			Duration(9999 * time.Second / 1_000),
		},
		{
			"2.0hours46.0minutes39.0seconds",
			Duration(9999 * time.Second / 1_000),
		},
		{
			"2.0hours 46.0minutes 39.0seconds",
			Duration(9999 * time.Second / 1_000),
		},
		{
			"2.0  hours  46.0  minutes  39.0  seconds",
			Duration(9999 * time.Second / 1_000),
		},
	}
	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			d, err := ParseDuration(s.str)
			require.NoError(t, err)
			assert.Equal(t, s.d, d)
		})
	}
}

func TestParseInvalidDuration(t *testing.T) {
	cases := []string{
		"not a duration",
		"PT.S",
		" PT1S",
		"PT1S ",
		".seconds",
		".s",
		"s",
		"20 hours with other stuff should not be valid",
		"20 seconds with other stuff should not be valid",
		"20 minutes with other stuff should not be valid",
		"20 ms with other stuff should not be valid",
		"20 us with other stuff should not be valid",
		"3 hours is longer than 10 seconds",
		"",
		"\t",
		" ",
	}

	for _, s := range cases {
		t.Run(s, func(t *testing.T) {
			d, err := ParseDuration(s)
			require.NotNil(t, err, "expected an error but got nil")
			expected := fmt.Sprintf("could not parse duration from %q", s)
			require.True(
				t,
				strings.Contains(err.Error(), expected),
				`The error message %q should contain the text %q`,
				err.Error(),
				expected,
			)
			assert.Equal(t, Duration(0), d)
		})
	}
}
