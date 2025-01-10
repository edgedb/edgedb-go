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
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalOptionalBigInt(t *testing.T) {
	cases := []struct {
		input    OptionalBigInt
		expected string
	}{
		{OptionalBigInt{}, "null"},
		{OptionalBigInt{big.NewInt(7), true}, `7`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalBigInt(t *testing.T) {
	cases := []struct {
		expected OptionalBigInt
		input    string
	}{
		{OptionalBigInt{}, "null"},
		{OptionalBigInt{big.NewInt(7), true}, `7`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalBigInt
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalBigInt{big.NewInt(1), true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}
