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
