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

func TestMarshalOptionalInt16(t *testing.T) {
	cases := []struct {
		input    OptionalInt16
		expected string
	}{
		{OptionalInt16{}, "null"},
		{OptionalInt16{7, true}, `7`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalInt16(t *testing.T) {
	cases := []struct {
		expected OptionalInt16
		input    string
	}{
		{OptionalInt16{}, "null"},
		{OptionalInt16{7, true}, `7`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalInt16
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalInt16{1, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestMarshalOptionalInt32(t *testing.T) {
	cases := []struct {
		input    OptionalInt32
		expected string
	}{
		{OptionalInt32{}, "null"},
		{OptionalInt32{7, true}, `7`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalInt32(t *testing.T) {
	cases := []struct {
		expected OptionalInt32
		input    string
	}{
		{OptionalInt32{}, "null"},
		{OptionalInt32{7, true}, `7`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalInt32
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalInt32{1, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestMarshalOptionalInt64(t *testing.T) {
	cases := []struct {
		input    OptionalInt64
		expected string
	}{
		{OptionalInt64{}, "null"},
		{OptionalInt64{7, true}, `7`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalInt64(t *testing.T) {
	cases := []struct {
		expected OptionalInt64
		input    string
	}{
		{OptionalInt64{}, "null"},
		{OptionalInt64{7, true}, `7`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalInt64
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalInt64{1, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestMarshalOptionalFloat32(t *testing.T) {
	cases := []struct {
		input    OptionalFloat32
		expected string
	}{
		{OptionalFloat32{}, "null"},
		{OptionalFloat32{7.2, true}, `7.2`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalFloat32(t *testing.T) {
	cases := []struct {
		expected OptionalFloat32
		input    string
	}{
		{OptionalFloat32{}, "null"},
		{OptionalFloat32{7.2, true}, `7.2`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalFloat32
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalFloat32{1, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}

func TestMarshalOptionalFloat64(t *testing.T) {
	cases := []struct {
		input    OptionalFloat64
		expected string
	}{
		{OptionalFloat64{}, "null"},
		{OptionalFloat64{7.2, true}, `7.2`},
	}

	for _, c := range cases {
		t.Run(c.expected, func(t *testing.T) {
			b, err := json.Marshal(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(b))
		})
	}
}

func TestUnmarshalOptionalFloat64(t *testing.T) {
	cases := []struct {
		expected OptionalFloat64
		input    string
	}{
		{OptionalFloat64{}, "null"},
		{OptionalFloat64{7.2, true}, `7.2`},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalFloat64
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalFloat64{1, true}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}
