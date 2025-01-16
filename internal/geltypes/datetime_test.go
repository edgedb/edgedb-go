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

package geltypes

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalOptionalDateTime(t *testing.T) {
	cases := []struct {
		input    OptionalDateTime
		expected string
	}{
		{OptionalDateTime{}, "null"},
		{
			OptionalDateTime{time.Unix(30, 1_000).UTC(), true},
			`"1970-01-01T00:00:30.000001Z"`,
		},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalDateTime(t *testing.T) {
	cases := []struct {
		expected OptionalDateTime
		input    string
	}{
		{OptionalDateTime{}, "null"},
		{
			OptionalDateTime{time.Unix(30, 1_000).UTC(), true},
			`"1970-01-01T00:00:30.000001Z"`,
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalDateTime
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalDateTime{time.Unix(999999999, 999999), true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

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

func TestMarshalLocalDateTime(t *testing.T) {
	cases := []struct {
		input    LocalDateTime
		expected string
	}{
		{LocalDateTime{30000000}, `"0001-01-01T00:00:30"`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalLocalDateTime(t *testing.T) {
	cases := []struct {
		expected LocalDateTime
		input    string
	}{
		{LocalDateTime{30000000}, `"0001-01-01T00:00:30"`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty LocalDateTime
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := LocalDateTime{99999}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestMarshalOptionalLocalDateTime(t *testing.T) {
	cases := []struct {
		input    OptionalLocalDateTime
		expected string
	}{
		{OptionalLocalDateTime{}, `null`},
		{
			OptionalLocalDateTime{LocalDateTime{30000000}, true},
			`"0001-01-01T00:00:30"`,
		},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalLocalDateTime(t *testing.T) {
	cases := []struct {
		expected OptionalLocalDateTime
		input    string
	}{
		{OptionalLocalDateTime{}, `null`},
		{
			OptionalLocalDateTime{LocalDateTime{30000000}, true},
			`"0001-01-01T00:00:30"`,
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalLocalDateTime
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalLocalDateTime{LocalDateTime{999999999}, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
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

func TestMarshalLocalDate(t *testing.T) {
	cases := []struct {
		input    LocalDate
		expected string
	}{
		{LocalDate{7}, `"0001-01-08"`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalLocalDate(t *testing.T) {
	cases := []struct {
		expected LocalDate
		input    string
	}{
		{LocalDate{7}, `"0001-01-08"`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty LocalDate
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := LocalDate{7}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestMarshalOptionalLocalDate(t *testing.T) {
	cases := []struct {
		input    OptionalLocalDate
		expected string
	}{
		{OptionalLocalDate{}, `null`},
		{OptionalLocalDate{LocalDate{7}, true}, `"0001-01-08"`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalLocalDate(t *testing.T) {
	cases := []struct {
		expected OptionalLocalDate
		input    string
	}{
		{OptionalLocalDate{}, `null`},
		{OptionalLocalDate{LocalDate{7}, true}, `"0001-01-08"`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalLocalDate
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalLocalDate{LocalDate{999999}, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
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

func TestMarshalLocalTime(t *testing.T) {
	cases := []struct {
		input    LocalTime
		expected string
	}{
		{LocalTime{30_000_000}, `"00:00:30"`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalLocalTime(t *testing.T) {
	cases := []struct {
		expected LocalTime
		input    string
	}{
		{LocalTime{30_000_000}, `"00:00:30"`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty LocalTime
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := LocalTime{99999}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestMarshalOptionalLocalTime(t *testing.T) {
	cases := []struct {
		input    OptionalLocalTime
		expected string
	}{
		{OptionalLocalTime{}, "null"},
		{OptionalLocalTime{LocalTime{30_000_000}, true}, `"00:00:30"`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalLocalTime(t *testing.T) {
	cases := []struct {
		expected OptionalLocalTime
		input    string
	}{
		{OptionalLocalTime{}, "null"},
		{OptionalLocalTime{LocalTime{30_000_000}, true}, `"00:00:30"`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalLocalTime
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalLocalTime{LocalTime{99999}, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestParseDuration(t *testing.T) {
	samples := []struct {
		str string
		d   Duration
	}{
		// seconds
		{"PT0S", Duration(0)},
		{"PT-0S", Duration(0)},
		{"PT0.000001S", Duration(1)},
		{"PT-0.000001S", Duration(-1)},
		{"PT1S", Duration(time.Second / 1_000)},
		{"PT-1S", Duration(time.Second / -1_000)},
		{"PT1.0S", Duration(time.Second / 1_000)},
		{"PT-1.0S", Duration(time.Second / -1_000)},
		{"PT1.S", Duration(time.Second / 1_000)},
		{"PT-1.S", Duration(time.Second / -1_000)},
		{"PT.1S", Duration(time.Second / 10 / 1_000)},
		{"PT-.1S", Duration(time.Second / 10 / -1_000)},

		// minutes
		{"PT1M", Duration(time.Minute / 1_000)},
		{"PT-1M", Duration(time.Minute / -1_000)},
		{"PT1.0M", Duration(time.Minute / 1_000)},
		{"PT-1.0M", Duration(time.Minute / -1_000)},
		{"PT1.M", Duration(time.Minute / 1_000)},
		{"PT-1.M", Duration(time.Minute / -1_000)},
		{"PT.1M", Duration(time.Minute / 10 / 1_000)},
		{"PT-.1M", Duration(time.Minute / 10 / -1_000)},

		// hours
		{"PT1H", Duration(time.Hour / 1_000)},
		{"PT-1H", Duration(time.Hour / -1_000)},
		{"PT1.0H", Duration(time.Hour / 1_000)},
		{"PT-1.0H", Duration(time.Hour / -1_000)},
		{"PT1.H", Duration(time.Hour / 1_000)},
		{"PT-1.H", Duration(time.Hour / -1_000)},
		{"PT.1H", Duration(time.Hour / 10 / 1_000)},
		{"PT-.1H", Duration(time.Hour / 10 / -1_000)},

		// no unit is hours
		{"PT", Duration(0)},
		{"PT1", Duration(3600_000_000)},
		{"PT-1", Duration(-3600_000_000)},
		{"PT1.", Duration(3600_000_000)},
		{"PT-1.", Duration(-3600_000_000)},
		{"PT.1", Duration(360_000_000)},
		{"PT-.1", Duration(-360_000_000)},
		{"PT1.0", Duration(3600_000_000)},
		{"PT-1.0", Duration(-3600_000_000)},

		{"PT2H46M39S", Duration(9999 * time.Second / 1_000)},
		{"PT2.0H46.0M39.0S", Duration(9999 * time.Second / 1_000)},

		{"-0s", Duration(0)},
		{"1s", Duration(time.Second / 1_000)},
		{"-1s", Duration(time.Second / -1_000)},
		{"1.0s", Duration(time.Second / 1_000)},
		{"-1.0s", Duration(time.Second / -1_000)},
		{".1s", Duration(time.Second / 10 / 1_000)},

		{"1ms", Duration(time.Millisecond / 1_000)},
		{"-1ms", Duration(time.Millisecond / -1_000)},
		{"1.0ms", Duration(time.Millisecond / 1_000)},
		{"-1.0ms", Duration(time.Millisecond / -1_000)},
		{".1ms", Duration(time.Millisecond / 10 / 1_000)},

		{"1us", Duration(time.Microsecond / 1_000)},
		{"-1us", Duration(time.Microsecond / -1_000)},
		{"1.0us", Duration(time.Microsecond / 1_000)},
		{"-1.0us", Duration(time.Microsecond / -1_000)},
		{".1us", Duration(time.Microsecond / 10 / 1_000)},

		{"1m", Duration(time.Minute / 1_000)},
		{"-1m", Duration(time.Minute / -1_000)},
		{"1.0m", Duration(time.Minute / 1_000)},
		{"-1.0m", Duration(time.Minute / -1_000)},
		{".1m", Duration(time.Minute / 10 / 1_000)},

		{"1h", Duration(time.Hour / 1_000)},
		{"-1h", Duration(time.Hour / -1_000)},
		{"1.0h", Duration(time.Hour / 1_000)},
		{"-1.0h", Duration(time.Hour / -1_000)},
		{".1h", Duration(time.Hour / 10 / 1_000)},

		{"2h46m39s", Duration(9999 * time.Second / 1_000)},
		{"2h 46m 39s", Duration(9999 * time.Second / 1_000)},
		{"2  h  46  m  39  s", Duration(9999 * time.Second / 1_000)},
		{"2.0h46.0m39.0s", Duration(9999 * time.Second / 1_000)},
		{"2.0h 46.0m 39.0s", Duration(9999 * time.Second / 1_000)},
		{"2.0  h  46.0  m  39.0  s", Duration(9999 * time.Second / 1_000)},
		{"1h -120m 3600s", Duration(0)},
		{"1h -120m3600s", Duration(0)},
		{"-2h 60m 3600s", Duration(0)},
		{"1h 60m -7200s", Duration(0)},

		{"1second", Duration(time.Second / 1_000)},
		{"-1second", Duration(-time.Second / 1_000)},
		{"1.0second", Duration(time.Second / 1_000)},
		{"-1.0second", Duration(-time.Second / 1_000)},
		{".1second", Duration(time.Second / 10 / 1_000)},
		{"1minute", Duration(time.Minute / 1_000)},
		{"-1minute", Duration(-time.Minute / 1_000)},
		{"1.0minute", Duration(time.Minute / 1_000)},
		{"-1.0minute", Duration(-time.Minute / 1_000)},
		{".1minute", Duration(time.Minute / 10 / 1_000)},
		{"1hour", Duration(time.Hour / 1_000)},
		{"-1hour", Duration(-time.Hour / 1_000)},
		{"1.0hour", Duration(time.Hour / 1_000)},
		{"-1.0hour", Duration(-time.Hour / 1_000)},
		{".1hour", Duration(time.Hour / 10 / 1_000)},
		{"-\t2\thour\t60\tminute\t3600\tsecond", Duration(0)},
		{"1   hour 60  minute -   7200   second", Duration(0)},
		{"2hour 46minute 39second", Duration(9999 * time.Second / 1_000)},
		{
			"2  hour  46  minute  39  second",
			Duration(9999 * time.Second / 1_000),
		},
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
		{"-1seconds", Duration(-time.Second / 1_000)},
		{"1.0seconds", Duration(time.Second / 1_000)},
		{"-1.0seconds", Duration(-time.Second / 1_000)},
		{".1seconds", Duration(time.Second / 10 / 1_000)},
		{"1minutes", Duration(time.Minute / 1_000)},
		{"-1minutes", Duration(-time.Minute / 1_000)},
		{"1.0minutes", Duration(time.Minute / 1_000)},
		{"-1.0minutes", Duration(-time.Minute / 1_000)},
		{".1minutes", Duration(time.Minute / 10 / 1_000)},
		{"1hours", Duration(time.Hour / 1_000)},
		{"-1hours", Duration(-time.Hour / 1_000)},
		{"1.0hours", Duration(time.Hour / 1_000)},
		{"-1.0hours", Duration(-time.Hour / 1_000)},
		{".1hours", Duration(time.Hour / 10 / 1_000)},
		{"\t-\t2\thours\t60\tminutes\t3600\tseconds\t", Duration(0)},
		{"1   hours 60  minutes -   7200   seconds", Duration(0)},
		{"1hours -120minutes 3600seconds", Duration(0)},
		{"2hours 46minutes 39seconds", Duration(9999 * time.Second / 1_000)},
		{
			"2  hours  46  minutes  39  seconds",
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
		" ",
		" PT1S",
		"",
		"-.1s",
		"-.5second",
		"-.5seconds",
		"-.1 s",
		"-.5 second",
		"-.5 seconds",
		"-.s",
		"-1.s",
		".s",
		".seconds",
		"1.s",
		"1h-120m3600s",
		"1hour-120minute3600second",
		"1hours-120minutes3600seconds",
		"1hours120minutes3600seconds",
		"2.0hour46.0minutes39.0seconds",
		"2.0hours46.0minutes39.0seconds",
		"20 hours with other stuff should not be valid",
		"20 minutes with other stuff should not be valid",
		"20 ms with other stuff should not be valid",
		"20 seconds with other stuff should not be valid",
		"20 us with other stuff should not be valid",
		"2hour46minute39second",
		"2hours46minutes39seconds",
		"3 hours is longer than 10 seconds",
		"P-.D",
		"P-D",
		"PD",
		"PT.S",
		"PT1S ",
		"\t",
		"not a duration",
		"s",
	}

	for _, s := range cases {
		t.Run(s, func(t *testing.T) {
			d, err := ParseDuration(s)
			require.NotNil(t, err, "expected an error but got nil")
			expected := fmt.Sprintf(
				"could not parse gel.Duration from %q",
				s)
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

func TestMarshalDuration(t *testing.T) {
	cases := []struct {
		input    Duration
		expected string
	}{
		{Duration(30_000_000), `30000000`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalDuration(t *testing.T) {
	cases := []struct {
		expected Duration
		input    string
	}{
		{Duration(30_000_000), `30000000`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty Duration
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := Duration(99999999)
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestAsNanosecondsDuration(t *testing.T) {
	var durationTruncMicroseconds = func(i int64) time.Duration {
		return time.Duration(time.Duration(i).Microseconds() * 1000)
	}

	cases := []struct {
		input    Duration
		mustFail bool
		expected time.Duration
	}{
		{Duration(math.MaxInt64), true, time.Duration(0)},
		{Duration(math.MaxInt64 / 100), true, time.Duration(0)},
		{Duration(math.MaxInt64/1000 + 1), true, time.Duration(0)},
		// Maximum possible value:
		{Duration(math.MaxInt64 / 1000), false,
			durationTruncMicroseconds(math.MaxInt64)},
		// Some arbitrary value within range
		{Duration(math.MaxInt64 / 1452), false,
			durationTruncMicroseconds(math.MaxInt64 / 1452 * 1000)},
		{Duration(0), false, time.Duration(0)},
		{Duration(math.MinInt64), true, time.Duration(0)},
		{Duration(math.MinInt64 / 100), true, time.Duration(0)},
		{Duration(math.MinInt64/1000 - 1), true, time.Duration(0)},
		// Minimum possible value
		{Duration(math.MinInt64 / 1000), false,
			durationTruncMicroseconds(math.MinInt64)},
		// Some arbitrary value within range
		{Duration(math.MinInt64 / 6946), false,
			durationTruncMicroseconds(math.MinInt64 / 6946 * 1000)},
	}

	for _, c := range cases {
		t.Run(c.input.String(), func(t *testing.T) {
			d, err := c.input.AsNanoseconds()
			if c.mustFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, c.expected, d)
		})
	}
}

func TestMarshalOptionalDuration(t *testing.T) {
	cases := []struct {
		input    OptionalDuration
		expected string
	}{
		{OptionalDuration{}, "null"},
		{OptionalDuration{Duration(30_000_000), true}, `30000000`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalDuration(t *testing.T) {
	cases := []struct {
		expected OptionalDuration
		input    string
	}{
		{OptionalDuration{}, "null"},
		{OptionalDuration{Duration(30_000_000), true}, `30000000`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalDuration
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalDuration{99999999, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestRelativeDurationUnmarshalText(t *testing.T) {
	samples := []struct {
		str string
		d   RelativeDuration
	}{
		// seconds
		{"PT0S", NewRelativeDuration(0, 0, 0)},
		{"PT-0S", NewRelativeDuration(0, 0, 0)},
		{"PT0.000001S", NewRelativeDuration(0, 0, 1)},
		{"PT-0.000001S", NewRelativeDuration(0, 0, -1)},
		{"PT1S", NewRelativeDuration(0, 0, 1_000_000)},
		{"PT-1S", NewRelativeDuration(0, 0, -1_000_000)},
		{"PT1.0S", NewRelativeDuration(0, 0, 1_000_000)},
		{"PT-1.0S", NewRelativeDuration(0, 0, -1_000_000)},
		{"PT1.S", NewRelativeDuration(0, 0, 1_000_000)},
		{"PT-1.S", NewRelativeDuration(0, 0, -1_000_000)},
		{"PT.1S", NewRelativeDuration(0, 0, 100_000)},
		{"PT-.1S", NewRelativeDuration(0, 0, -100_000)},

		// minutes
		{"PT1M", NewRelativeDuration(0, 0, 60_000_000)},
		{"PT-1M", NewRelativeDuration(0, 0, -60_000_000)},
		{"PT1.0M", NewRelativeDuration(0, 0, 60_000_000)},
		{"PT-1.0M", NewRelativeDuration(0, 0, -60_000_000)},
		{"PT1.M", NewRelativeDuration(0, 0, 60_000_000)},
		{"PT-1.M", NewRelativeDuration(0, 0, -60_000_000)},
		{"PT.1M", NewRelativeDuration(0, 0, 6_000_000)},
		{"PT-.1M", NewRelativeDuration(0, 0, -6_000_000)},

		// hours
		{"PT1H", NewRelativeDuration(0, 0, 3600_000_000)},
		{"PT-1H", NewRelativeDuration(0, 0, -3600_000_000)},
		{"PT1.0H", NewRelativeDuration(0, 0, 3600_000_000)},
		{"PT-1.0H", NewRelativeDuration(0, 0, -3600_000_000)},
		{"PT1.H", NewRelativeDuration(0, 0, 3600_000_000)},
		{"PT-1.H", NewRelativeDuration(0, 0, -3600_000_000)},
		{"PT.1H", NewRelativeDuration(0, 0, 360_000_000)},
		{"PT-.1H", NewRelativeDuration(0, 0, -360_000_000)},
		{"PT.01H", NewRelativeDuration(0, 0, 36_000_000)},
		{"PT-.01H", NewRelativeDuration(0, 0, -36_000_000)},

		// no unit is hours
		{"PT", NewRelativeDuration(0, 0, 0)},
		{"PT1", NewRelativeDuration(0, 0, 3600_000_000)},
		{"PT-1", NewRelativeDuration(0, 0, -3600_000_000)},
		{"PT1.", NewRelativeDuration(0, 0, 3600_000_000)},
		{"PT-1.", NewRelativeDuration(0, 0, -3600_000_000)},
		{"PT.1", NewRelativeDuration(0, 0, 360_000_000)},
		{"PT-.1", NewRelativeDuration(0, 0, -360_000_000)},
		{"PT1.0", NewRelativeDuration(0, 0, 3600_000_000)},
		{"PT-1.0", NewRelativeDuration(0, 0, -3600_000_000)},

		// days
		{"P1D", NewRelativeDuration(0, 1, 0)},
		{"P-1D", NewRelativeDuration(0, -1, 0)},
		{"P1DT", NewRelativeDuration(0, 1, 0)},
		{"P-1DT", NewRelativeDuration(0, -1, 0)},
		{"P1DT0", NewRelativeDuration(0, 1, 0)},
		{"P-1DT-0", NewRelativeDuration(0, -1, 0)},
		{"P1.0D", NewRelativeDuration(0, 1, 0)},
		{"P-1.0D", NewRelativeDuration(0, -1, 0)},
		{"P1.D", NewRelativeDuration(0, 1, 0)},
		{"P-1.D", NewRelativeDuration(0, -1, 0)},
		{"P.1D", NewRelativeDuration(0, 0, 2*3600_000_000+24*60_000_000)},
		{
			"P-.1D",
			NewRelativeDuration(0, 0, -(2*3600_000_000 + 24*60_000_000)),
		},
		{"P.01D", NewRelativeDuration(0, 0, 14*60_000_000+24*1_000_000)},
		{"P-.01D", NewRelativeDuration(0, 0, -(14*60_000_000 + 24*1_000_000))},

		// weeks
		{"P1W", NewRelativeDuration(0, 7, 0)},
		{"P-1W", NewRelativeDuration(0, -7, 0)},
		{"P1WT", NewRelativeDuration(0, 7, 0)},
		{"P-1WT", NewRelativeDuration(0, -7, 0)},
		{"P1.0W", NewRelativeDuration(0, 7, 0)},
		{"P-1.0W", NewRelativeDuration(0, -7, 0)},
		{"P1.W", NewRelativeDuration(0, 7, 0)},
		{"P-1.W", NewRelativeDuration(0, -7, 0)},
		{"P.1W", NewRelativeDuration(0, 0, 16*3600_000_000+48*60_000_000)},
		{
			"P-.1W",
			NewRelativeDuration(0, 0, -(16*3600_000_000 + 48*60_000_000)),
		},
		{"P.01W", NewRelativeDuration(
			0,
			0,
			1*3600_000_000+40*60_000_000+48*1_000_000,
		)},
		{"P-.01W", NewRelativeDuration(
			0,
			0,
			-(1*3600_000_000 + 40*60_000_000 + 48*1_000_000),
		)},

		// months
		{"P1M", NewRelativeDuration(1, 0, 0)},
		{"P-1M", NewRelativeDuration(-1, 0, 0)},
		{"P1MT", NewRelativeDuration(1, 0, 0)},
		{"P-1MT", NewRelativeDuration(-1, 0, 0)},
		{"P1.0M", NewRelativeDuration(1, 0, 0)},
		{"P-1.0M", NewRelativeDuration(-1, 0, 0)},
		{"P1.M", NewRelativeDuration(1, 0, 0)},
		{"P-1.M", NewRelativeDuration(-1, 0, 0)},
		{"P.1M", NewRelativeDuration(0, 3, 0)},
		{"P-.1M", NewRelativeDuration(0, -3, 0)},
		{"P.01M", NewRelativeDuration(0, 0, 7*3600_000_000+12*60_000_000)},
		{
			"P-.01M",
			NewRelativeDuration(0, 0, -(7*3600_000_000 + 12*60_000_000)),
		},

		// years
		{"P1Y", NewRelativeDuration(12, 0, 0)},
		{"P-1Y", NewRelativeDuration(-12, 0, 0)},
		{"P1YT", NewRelativeDuration(12, 0, 0)},
		{"P-1YT", NewRelativeDuration(-12, 0, 0)},
		{"P1.0Y", NewRelativeDuration(12, 0, 0)},
		{"P-1.0Y", NewRelativeDuration(-12, 0, 0)},
		{"P1.Y", NewRelativeDuration(12, 0, 0)},
		{"P-1.Y", NewRelativeDuration(-12, 0, 0)},
		{"P.1Y", NewRelativeDuration(1, 0, 0)},
		{"P-.1Y", NewRelativeDuration(-1, 0, 0)},
		{"P.01Y", NewRelativeDuration(0, 0, 0)},
		{"P-.01Y", NewRelativeDuration(0, 0, 0)},

		// order of units doesn't matter
		{"P1Y1M1W1D", NewRelativeDuration(13, 8, 0)},
		{"P1Y1D1M1W", NewRelativeDuration(13, 8, 0)},
		{"P1Y1W1D1M", NewRelativeDuration(13, 8, 0)},
		{"P1D1W1M1Y", NewRelativeDuration(13, 8, 0)},
		{
			"PT1H1M1S",
			NewRelativeDuration(0, 0, 3600_000_000+60_000_000+1_000_000),
		},
		{
			"PT1S1H1M",
			NewRelativeDuration(0, 0, 3600_000_000+60_000_000+1_000_000),
		},
		{
			"PT1M1S1H",
			NewRelativeDuration(0, 0, 3600_000_000+60_000_000+1_000_000),
		},

		// sings are independent
		{"P-1M-1DT-0.000001S", NewRelativeDuration(-1, -1, -1)},
		{"P-1M-1DT0.000001S", NewRelativeDuration(-1, -1, 1)},
		{"P1M-1DT-0.000001S", NewRelativeDuration(1, -1, -1)},
		{"P-1M1DT-0.000001S", NewRelativeDuration(-1, 1, -1)},
		{"P-1M1DT0.000001S", NewRelativeDuration(-1, 1, 1)},
		{"P1M-1DT0.000001S", NewRelativeDuration(1, -1, 1)},
		{"P1M1DT-0.000001S", NewRelativeDuration(1, 1, -1)},
		{"P1M1DT0.000001S", NewRelativeDuration(1, 1, 1)},

		{"P1DT1", NewRelativeDuration(0, 1, 3600_000_000)},
		{"P-1DT-1", NewRelativeDuration(0, -1, -3600_000_000)},
		{"P2Y3M4W5DT6H7M8.9S", NewRelativeDuration(
			27,
			5+4*7,
			6*3600_000_000+7*60_000_000+8.9*1_000_000,
		)},
		{"P2Y3M-4DT23M12.345678S", NewRelativeDuration(
			2*12+3,
			-4,
			23*60_000_000+12345678,
		)},

		{"1us", NewRelativeDuration(0, 0, 1)},
		{"-1us", NewRelativeDuration(0, 0, -1)},
		{"1.0us", NewRelativeDuration(0, 0, 1)},
		{"-1.0us", NewRelativeDuration(0, 0, -1)},
		{".1us", NewRelativeDuration(0, 0, 0)},

		{"1ms", NewRelativeDuration(0, 0, 1_000)},
		{"-1ms", NewRelativeDuration(0, 0, -1_000)},
		{"1.0ms", NewRelativeDuration(0, 0, 1_000)},
		{"-1.0ms", NewRelativeDuration(0, 0, -1_000)},
		{".1ms", NewRelativeDuration(0, 0, 100)},

		{"-0s", NewRelativeDuration(0, 0, 0)},
		{"1s", NewRelativeDuration(0, 0, 1_000_000)},
		{"-1s", NewRelativeDuration(0, 0, -1_000_000)},
		{"1.0s", NewRelativeDuration(0, 0, 1_000_000)},
		{"-1.0s", NewRelativeDuration(0, 0, -1_000_000)},
		{".1s", NewRelativeDuration(0, 0, 100_000)},

		{"1m", NewRelativeDuration(0, 0, 60_000_000)},
		{"-1m", NewRelativeDuration(0, 0, -60_000_000)},
		{"1.0m", NewRelativeDuration(0, 0, 60_000_000)},
		{"-1.0m", NewRelativeDuration(0, 0, -60_000_000)},
		{".1m", NewRelativeDuration(0, 0, 6_000_000)},

		{"1h", NewRelativeDuration(0, 0, 3600_000_000)},
		{"-1h", NewRelativeDuration(0, 0, -3600_000_000)},
		{"1.0h", NewRelativeDuration(0, 0, 3600_000_000)},
		{"-1.0h", NewRelativeDuration(0, 0, -3600_000_000)},
		{".1h", NewRelativeDuration(0, 0, 360_000_000)},

		{"1d", NewRelativeDuration(0, 1, 0)},
		{"-1d", NewRelativeDuration(0, -1, 0)},
		{"1.0d", NewRelativeDuration(0, 1, 0)},
		{"-1.0d", NewRelativeDuration(0, -1, 0)},
		{".1d", NewRelativeDuration(0, 0, 2*3600_000_000+24*60_000_000)},
		{
			"-0.1d",
			NewRelativeDuration(0, 0, -(2*3600_000_000 + 24*60_000_000)),
		},

		{"1w", NewRelativeDuration(0, 7, 0)},
		{"-1w", NewRelativeDuration(0, -7, 0)},
		{"1.0w", NewRelativeDuration(0, 7, 0)},
		{"-1.0w", NewRelativeDuration(0, -7, 0)},
		{".1w", NewRelativeDuration(0, 0, 16*3600_000_000+48*60_000_000)},
		{
			"-0.1w",
			NewRelativeDuration(0, 0, -(16*3600_000_000 + 48*60_000_000)),
		},

		{"1mon", NewRelativeDuration(1, 0, 0)},
		{"-1mon", NewRelativeDuration(-1, 0, 0)},
		{"1.0mon", NewRelativeDuration(1, 0, 0)},
		{"-1.0mon", NewRelativeDuration(-1, 0, 0)},
		{".1mon", NewRelativeDuration(0, 3, 0)},
		{"-0.1mon", NewRelativeDuration(0, -3, 0)},

		{"1y", NewRelativeDuration(12, 0, 0)},
		{"-1y", NewRelativeDuration(-12, 0, 0)},
		{"1.0y", NewRelativeDuration(12, 0, 0)},
		{"-1.0y", NewRelativeDuration(-12, 0, 0)},
		{".1y", NewRelativeDuration(1, 0, 0)},
		{"-0.1y", NewRelativeDuration(-1, 0, 0)},

		{"1dec", NewRelativeDuration(120, 0, 0)},
		{"-1dec", NewRelativeDuration(-120, 0, 0)},
		{"1.0dec", NewRelativeDuration(120, 0, 0)},
		{"-1.0dec", NewRelativeDuration(-120, 0, 0)},
		{".1dec", NewRelativeDuration(12, 0, 0)},
		{"-0.1dec", NewRelativeDuration(-12, 0, 0)},

		{"1c", NewRelativeDuration(1_200, 0, 0)},
		{"-1c", NewRelativeDuration(-1_200, 0, 0)},
		{"1.0c", NewRelativeDuration(1_200, 0, 0)},
		{"-1.0c", NewRelativeDuration(-1_200, 0, 0)},
		{".1c", NewRelativeDuration(120, 0, 0)},
		{"-0.1c", NewRelativeDuration(-120, 0, 0)},

		{"1mil", NewRelativeDuration(12_000, 0, 0)},
		{"-1mil", NewRelativeDuration(-12_000, 0, 0)},
		{"1.0mil", NewRelativeDuration(12_000, 0, 0)},
		{"-1.0mil", NewRelativeDuration(-12_000, 0, 0)},
		{".1mil", NewRelativeDuration(1_200, 0, 0)},
		{"-0.1mil", NewRelativeDuration(-1_200, 0, 0)},

		{"2h46m39s", NewRelativeDuration(0, 0, 9999*1_000_000)},
		{"2h 46m 39s", NewRelativeDuration(0, 0, 9999*1_000_000)},
		{"2  h  46  m  39  s", NewRelativeDuration(0, 0, 9999*1_000_000)},
		{"2.0h46.0m39.0s", NewRelativeDuration(0, 0, 9999*1_000_000)},
		{"2.0h 46.0m 39.0s", NewRelativeDuration(0, 0, 9999*1_000_000)},
		{
			"2.0  h  46.0  m  39.0  s",
			NewRelativeDuration(0, 0, 9999*1_000_000),
		},
		{"1h -120m 3600s", NewRelativeDuration(0, 0, 0)},
		{"1h -120m3600s", NewRelativeDuration(0, 0, 0)},
		{"-2h 60m 3600s", NewRelativeDuration(0, 0, 0)},
		{"1h 60m -7200s", NewRelativeDuration(0, 0, 0)},
		{
			"1mil 2c 3dec 4y 5mon 6w 7d 8h 9m 10s 11ms 12us",
			NewRelativeDuration(
				5+1234*12,
				49,
				8*3600_000_000+9*60_000_000+10_011_012,
			),
		},

		{"1second", NewRelativeDuration(0, 0, 1_000_000)},
		{"-1second", NewRelativeDuration(0, 0, -1_000_000)},
		{"1.0second", NewRelativeDuration(0, 0, 1_000_000)},
		{"-1.0second", NewRelativeDuration(0, 0, -1_000_000)},
		{".1second", NewRelativeDuration(0, 0, 100_000)},

		{"1minute", NewRelativeDuration(0, 0, 60_000_000)},
		{"-1minute", NewRelativeDuration(0, 0, -60_000_000)},
		{"1.0minute", NewRelativeDuration(0, 0, 60_000_000)},
		{"-1.0minute", NewRelativeDuration(0, 0, -60_000_000)},
		{".1minute", NewRelativeDuration(0, 0, 6_000_000)},

		{"1hour", NewRelativeDuration(0, 0, 3600_000_000)},
		{"-1hour", NewRelativeDuration(0, 0, -3600_000_000)},
		{"1.0hour", NewRelativeDuration(0, 0, 3600_000_000)},
		{"-1.0hour", NewRelativeDuration(0, 0, -3600_000_000)},
		{".1hour", NewRelativeDuration(0, 0, 360_000_000)},

		{"-\t2\thour\t60\tminute\t3600\tsecond", NewRelativeDuration(0, 0, 0)},
		{
			"1   hour 60  minute -   7200   second",
			NewRelativeDuration(0, 0, 0),
		},
		{"2hour 46minute 39second", NewRelativeDuration(0, 0, 9999*1_000_000)},
		{
			"2  hour  46  minute  39  second",
			NewRelativeDuration(0, 0, 9999*1_000_000),
		},
		{
			"2.0hour 46.0minute 39.0second",
			NewRelativeDuration(0, 0, 9999*1_000_000),
		},
		{
			"2.0  hour  46.0  minute  39.0  second",
			NewRelativeDuration(0, 0, 9999*1_000_000),
		},
		{
			"39.0\tsecond 2.0  hour  46.0  minute",
			NewRelativeDuration(0, 0, 9999*1_000_000),
		},

		{"1seconds", NewRelativeDuration(0, 0, 1_000_000)},
		{"-1seconds", NewRelativeDuration(0, 0, -1_000_000)},
		{"1.0seconds", NewRelativeDuration(0, 0, 1_000_000)},
		{"-1.0seconds", NewRelativeDuration(0, 0, -1_000_000)},
		{".1seconds", NewRelativeDuration(0, 0, 100_000)},

		{"1minutes", NewRelativeDuration(0, 0, 60_000_000)},
		{"-1minutes", NewRelativeDuration(0, 0, -60_000_000)},
		{"1.0minutes", NewRelativeDuration(0, 0, 60_000_000)},
		{"-1.0minutes", NewRelativeDuration(0, 0, -60_000_000)},
		{".1minutes", NewRelativeDuration(0, 0, 6_000_000)},

		{"1hours", NewRelativeDuration(0, 0, 3600_000_000)},
		{"-1hours", NewRelativeDuration(0, 0, -3600_000_000)},
		{"1.0hours", NewRelativeDuration(0, 0, 3600_000_000)},
		{"-1.0hours", NewRelativeDuration(0, 0, -3600_000_000)},
		{".1hours", NewRelativeDuration(0, 0, 360_000_000)},

		{
			"\t-\t2\thours\t60\tminutes\t3600\tseconds\t",
			NewRelativeDuration(0, 0, 0),
		},
		{
			"1   hours 60  minutes -   7200   seconds",
			NewRelativeDuration(0, 0, 0),
		},
		{"1hours -120minutes 3600seconds", NewRelativeDuration(0, 0, 0)},
		{
			"2hours 46minutes 39seconds",
			NewRelativeDuration(0, 0, 9999*1_000_000),
		},
		{
			"2  hours  46  minutes  39  seconds",
			NewRelativeDuration(0, 0, 9999*1_000_000),
		},
		{
			"2.0hours 46.0minutes 39.0seconds",
			NewRelativeDuration(0, 0, 9999*1_000_000),
		},
		{
			"2.0  hours  46.0  minutes  39.0  seconds",
			NewRelativeDuration(0, 0, 9999*1_000_000),
		},
	}
	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			var d RelativeDuration
			err := d.UnmarshalText([]byte(s.str))
			require.NoError(t, err)
			assert.Equal(t, s.d, d)
		})
	}
}

func TestParseInvalidRelativeDuration(t *testing.T) {
	cases := []string{
		" ",
		" PT1S",
		"",
		"-.1 s",
		"-.1s",
		"-.5 second",
		"-.5 seconds",
		"-.5second",
		"-.5seconds",
		"-.s",
		"-1.s",
		".s",
		".seconds",
		"1.s",
		"1h-120m3600s",
		"1hour-120minute3600second",
		"1hours-120minutes3600seconds",
		"1hours120minutes3600seconds",
		"2.0hour46.0minutes39.0seconds",
		"2.0hours46.0minutes39.0seconds",
		"20 hours with other stuff should not be valid",
		"20 minutes with other stuff should not be valid",
		"20 ms with other stuff should not be valid",
		"20 seconds with other stuff should not be valid",
		"20 us with other stuff should not be valid",
		"2hour46minute39second",
		"2hours46minutes39seconds",
		"3 hours is longer than 10 seconds",
		"P-.D",
		"P-D",
		"PD",
		"PT-.S",
		"PT-S",
		"PT.0Y",
		"PT.S",
		"PT0.Y",
		"PT1S ",
		"\t",
		"not a duration",
		"s",
	}

	for _, s := range cases {
		t.Run(s, func(t *testing.T) {
			var d RelativeDuration
			err := d.UnmarshalText([]byte(s))
			require.NotNil(t, err, "expected an error but got nil")
			expected := fmt.Sprintf(
				"could not parse gel.RelativeDuration from %q", s)
			require.True(t,
				strings.Contains(err.Error(), expected),
				`The error message %q should contain the text %q`,
				err.Error(),
				expected,
			)
			assert.Equal(t, NewRelativeDuration(0, 0, 0), d)
		})
	}
}

func TestMarshalRelativeDuration(t *testing.T) {
	cases := []struct {
		input    RelativeDuration
		expected string
	}{
		{RelativeDuration{7, 7, 7}, `"P7M7DT0.000007S"`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalRelativeDuration(t *testing.T) {
	cases := []struct {
		expected RelativeDuration
		input    string
	}{
		{RelativeDuration{7, 7, 7}, `"P7M7DT0.000007S"`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty RelativeDuration
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := RelativeDuration{9, 9, 9}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestMarshalOptionalRelativeDuration(t *testing.T) {
	cases := []struct {
		input    OptionalRelativeDuration
		expected string
	}{
		{OptionalRelativeDuration{}, "null"},
		{
			OptionalRelativeDuration{RelativeDuration{7, 7, 7}, true},
			`"P7M7DT0.000007S"`,
		},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalRelativeDuration(t *testing.T) {
	cases := []struct {
		expected OptionalRelativeDuration
		input    string
	}{
		{OptionalRelativeDuration{}, "null"},
		{
			OptionalRelativeDuration{RelativeDuration{7, 7, 7}, true},
			`"P7M7DT0.000007S"`,
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalRelativeDuration
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalRelativeDuration{
				RelativeDuration{9, 9, 9},
				true,
			}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestDateDurationUnmarshalText(t *testing.T) {
	samples := []struct {
		str string
		d   DateDuration
	}{
		{"PT", NewDateDuration(0, 0)},

		// days
		{"P1D", NewDateDuration(0, 1)},
		{"P-1D", NewDateDuration(0, -1)},
		{"P1DT", NewDateDuration(0, 1)},
		{"P-1DT", NewDateDuration(0, -1)},
		{"P1DT0", NewDateDuration(0, 1)},
		{"P-1DT-0", NewDateDuration(0, -1)},
		{"P1.0D", NewDateDuration(0, 1)},
		{"P-1.0D", NewDateDuration(0, -1)},
		{"P1.D", NewDateDuration(0, 1)},
		{"P-1.D", NewDateDuration(0, -1)},

		// weeks
		{"P1W", NewDateDuration(0, 7)},
		{"P-1W", NewDateDuration(0, -7)},
		{"P1WT", NewDateDuration(0, 7)},
		{"P-1WT", NewDateDuration(0, -7)},
		{"P1.0W", NewDateDuration(0, 7)},
		{"P-1.0W", NewDateDuration(0, -7)},
		{"P1.W", NewDateDuration(0, 7)},
		{"P-1.W", NewDateDuration(0, -7)},

		// months
		{"P1M", NewDateDuration(1, 0)},
		{"P-1M", NewDateDuration(-1, 0)},
		{"P1MT", NewDateDuration(1, 0)},
		{"P-1MT", NewDateDuration(-1, 0)},
		{"P1.0M", NewDateDuration(1, 0)},
		{"P-1.0M", NewDateDuration(-1, 0)},
		{"P1.M", NewDateDuration(1, 0)},
		{"P-1.M", NewDateDuration(-1, 0)},
		{"P.1M", NewDateDuration(0, 3)},
		{"P-.1M", NewDateDuration(0, -3)},

		// years
		{"P1Y", NewDateDuration(12, 0)},
		{"P-1Y", NewDateDuration(-12, 0)},
		{"P1YT", NewDateDuration(12, 0)},
		{"P-1YT", NewDateDuration(-12, 0)},
		{"P1.0Y", NewDateDuration(12, 0)},
		{"P-1.0Y", NewDateDuration(-12, 0)},
		{"P1.Y", NewDateDuration(12, 0)},
		{"P-1.Y", NewDateDuration(-12, 0)},
		{"P.1Y", NewDateDuration(1, 0)},
		{"P-.1Y", NewDateDuration(-1, 0)},
		{"P.01Y", NewDateDuration(0, 0)},
		{"P-.01Y", NewDateDuration(0, 0)},

		// order of units doesn't matter
		{"P1Y1M1W1D", NewDateDuration(13, 8)},
		{"P1Y1D1M1W", NewDateDuration(13, 8)},
		{"P1Y1W1D1M", NewDateDuration(13, 8)},
		{"P1D1W1M1Y", NewDateDuration(13, 8)},

		// sings are independent
		{"P-1M-1D", NewDateDuration(-1, -1)},
		{"P1M-1D", NewDateDuration(1, -1)},
		{"P-1M1D", NewDateDuration(-1, 1)},
		{"P1M1D", NewDateDuration(1, 1)},

		{"P1W", NewDateDuration(0, 7)},
		{"P1WT", NewDateDuration(0, 7)},
		{"P1.0W", NewDateDuration(0, 7)},
		{"P1.W", NewDateDuration(0, 7)},
		{"P1M", NewDateDuration(1, 0)},
		{"P1MT", NewDateDuration(1, 0)},
		{"P1.0M", NewDateDuration(1, 0)},
		{"P1.M", NewDateDuration(1, 0)},
		{"P.1M", NewDateDuration(0, 3)},
		{"P1Y", NewDateDuration(12, 0)},
		{"P1YT", NewDateDuration(12, 0)},
		{"P1.0Y", NewDateDuration(12, 0)},
		{"P1.Y", NewDateDuration(12, 0)},
		{"P.1Y", NewDateDuration(1, 0)},
		{"P.01Y", NewDateDuration(0, 0)},
		{"P.00001Y", NewDateDuration(0, 0)},
		{"P2Y3M4W5D", NewDateDuration(27, 5+4*7)},

		{"1d", NewDateDuration(0, 1)},
		{"-1d", NewDateDuration(0, -1)},
		{"1.0d", NewDateDuration(0, 1)},
		{"-1.0d", NewDateDuration(0, -1)},

		{"1w", NewDateDuration(0, 7)},
		{"-1w", NewDateDuration(0, -7)},
		{"1.0w", NewDateDuration(0, 7)},
		{"-1.0w", NewDateDuration(0, -7)},

		{"1mon", NewDateDuration(1, 0)},
		{"-1mon", NewDateDuration(-1, 0)},
		{"1.0mon", NewDateDuration(1, 0)},
		{"-1.0mon", NewDateDuration(-1, 0)},
		{".1mon", NewDateDuration(0, 3)},
		{"-0.1mon", NewDateDuration(0, -3)},

		{"1y", NewDateDuration(12, 0)},
		{"-1y", NewDateDuration(-12, 0)},
		{"1.0y", NewDateDuration(12, 0)},
		{"-1.0y", NewDateDuration(-12, 0)},
		{".1y", NewDateDuration(1, 0)},
		{"-0.1y", NewDateDuration(-1, 0)},

		{"1dec", NewDateDuration(120, 0)},
		{"-1dec", NewDateDuration(-120, 0)},
		{"1.0dec", NewDateDuration(120, 0)},
		{"-1.0dec", NewDateDuration(-120, 0)},
		{".1dec", NewDateDuration(12, 0)},
		{"-0.1dec", NewDateDuration(-12, 0)},

		{"1c", NewDateDuration(1_200, 0)},
		{"-1c", NewDateDuration(-1_200, 0)},
		{"1.0c", NewDateDuration(1_200, 0)},
		{"-1.0c", NewDateDuration(-1_200, 0)},
		{".1c", NewDateDuration(120, 0)},
		{"-0.1c", NewDateDuration(-120, 0)},

		{"1mil", NewDateDuration(12_000, 0)},
		{"-1mil", NewDateDuration(-12_000, 0)},
		{"1.0mil", NewDateDuration(12_000, 0)},
		{"-1.0mil", NewDateDuration(-12_000, 0)},
		{".1mil", NewDateDuration(1_200, 0)},
		{"-0.1mil", NewDateDuration(-1_200, 0)},

		{"1y -12mon 1w -7d", NewDateDuration(0, 0)},
		{"-1y 12mon -1w 7d", NewDateDuration(0, 0)},
		{"1mil 2c 3dec 4y 5mon 6w 7d", NewDateDuration(5+1234*12, 49)},
	}
	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			var d DateDuration
			err := d.UnmarshalText([]byte(s.str))
			require.NoError(t, err)
			assert.Equal(t, s.d, d)
		})
	}
}

func TestDateDurationUnmarshalTextInvalid(t *testing.T) {
	cases := []string{
		".1d",
		"-0.1d",
		".1w",
		"-0.1w",
		"2h46m39s",
		"2h 46m 39s",
		"2  h  46  m  39  s",
		"2.0h46.0m39.0s",
		"2.0h 46.0m 39.0s",
		"2.0  h  46.0  m  39.0  s",
		"PT1H1M1S",
		"PT1S1H1M",
		"PT1M1S1H",
		"P.01M",
		"P-.01M",
		"not a duration",
		"P.1W",
		"P.01M",
		"PT.S",
		"PT0.Y",
		"PT.0Y",
		" PT1S",
		"PT1S ",
		".seconds",
		".s",
		"s",
		"PT1",
		"20 hours with other stuff should not be valid",
		"20 seconds with other stuff should not be valid",
		"20 minutes with other stuff should not be valid",
		"20 ms with other stuff should not be valid",
		"20 us with other stuff should not be valid",
		"3 hours is longer than 10 seconds",
		"",
		"\t",
		" ",
		"P.1D",
		"P-.1D",
		"P.01D",
		"P-.01D",
		"P.1W",
		"P-.1W",
		"P-.01W",
	}

	for _, s := range cases {
		t.Run(s, func(t *testing.T) {
			var d DateDuration
			err := d.UnmarshalText([]byte(s))
			require.NotNil(t, err, "expected an error but got nil")
			expected := fmt.Sprintf(
				"could not parse gel.DateDuration from %q", s)
			require.True(t,
				strings.Contains(err.Error(), expected),
				`The error message %q should contain the text %q`,
				err.Error(),
				expected,
			)
			assert.Equal(t, NewDateDuration(0, 0), d)
		})
	}
}

func TestMarshalDateDuration(t *testing.T) {
	cases := []struct {
		input    DateDuration
		expected string
	}{
		{DateDuration{7, 7}, `"P7M7D"`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalDateDuration(t *testing.T) {
	cases := []struct {
		expected DateDuration
		input    string
	}{
		{DateDuration{7, 7}, `"P7M7D"`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty DateDuration
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := DateDuration{999, 999}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestMarshalOptionalDateDuration(t *testing.T) {
	cases := []struct {
		input    OptionalDateDuration
		expected string
	}{
		{OptionalDateDuration{}, "null"},
		{OptionalDateDuration{DateDuration{7, 7}, true}, `"P7M7D"`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalDateDuration(t *testing.T) {
	cases := []struct {
		expected OptionalDateDuration
		input    string
	}{
		{OptionalDateDuration{}, "null"},
		{OptionalDateDuration{DateDuration{7, 7}, true}, `"P7M7D"`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalDateDuration
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalDateDuration{DateDuration{7, 7}, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}
