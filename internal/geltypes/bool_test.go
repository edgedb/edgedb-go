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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalOptionalBool(t *testing.T) {
	cases := []struct {
		input    OptionalBool
		expected string
	}{
		{OptionalBool{}, "null"},
		{OptionalBool{true, true}, `true`},
		{OptionalBool{false, true}, `false`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalBool(t *testing.T) {
	cases := []struct {
		expected OptionalBool
		input    string
	}{
		{OptionalBool{}, "null"},
		{OptionalBool{true, true}, `true`},
		{OptionalBool{false, true}, `false`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalBool
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalBool{true, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)

			notEmpty = OptionalBool{false, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}
