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
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestDurationString(t *testing.T) {
	samples := []struct {
		str string
		d   Duration
	}{
		{"1us", Duration(1)},
		{"999us", Duration(999)},
		{"-1us", Duration(-1)},
		{"-999us", Duration(-999)},

		{"1.234ms", Duration(1_234)},
		{"1.2ms", Duration(1_200)},
		{"1.004ms", Duration(1_004)},
		{"999.234ms", Duration(999_234)},
		{"999.2ms", Duration(999_200)},
		{"999.004ms", Duration(999_004)},
		{"-1.234ms", Duration(-1_234)},
		{"-1.2ms", Duration(-1_200)},
		{"-1.004ms", Duration(-1_004)},
		{"-999.234ms", Duration(-999_234)},
		{"-999.2ms", Duration(-999_200)},
		{"-999.004ms", Duration(-999_004)},

		{"0s", Duration(0)},
		{"1s", Duration(1_000_000)},
		{"59s", Duration(59_000_000)},
		{"1.234567s", Duration(1_234_567)},
		{"1.2s", Duration(1_200_000)},
		{"1.000007s", Duration(1_000_007)},
		{"59.234567s", Duration(59_234_567)},
		{"59.2s", Duration(59_200_000)},
		{"59.000007s", Duration(59_000_007)},
		{"-1s", Duration(-1_000_000)},
		{"-59s", Duration(-59_000_000)},
		{"-1.234567s", Duration(-1_234_567)},
		{"-1.2s", Duration(-1_200_000)},
		{"-1.000007s", Duration(-1_000_007)},
		{"-59.234567s", Duration(-59_234_567)},
		{"-59.2s", Duration(-59_200_000)},
		{"-59.000007s", Duration(-59_000_007)},

		{"1m", Duration(60_000_000)},
		{"59m", Duration(3540000000)},
		{"-1m", Duration(-60_000_000)},
		{"-59m", Duration(-3540000000)},

		{"1h", Duration(3600000000)},
		{"24h", Duration(86400000000)},
		{"-1h", Duration(-3600000000)},
		{"-24h", Duration(-86400000000)},

		{"59m59s", Duration(3599000000)},
		{"1h59m59s", Duration(7199000000)},
		{"1h59m", Duration(7140000000)},
		{"854015929h20m18.258432s", Duration(3074457345618258432)},
		{"-59m59s", Duration(-3599000000)},
		{"-1h59m59s", Duration(-7199000000)},
		{"-1h59m", Duration(-7140000000)},
		{"-854015929h20m18.258432s", Duration(-3074457345618258432)},
	}

	for _, s := range samples {
		t.Run(s.str, func(t *testing.T) {
			assert.Equal(t, s.str, s.d.String())
		})
	}
}
