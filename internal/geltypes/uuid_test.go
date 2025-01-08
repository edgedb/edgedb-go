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
	"encoding/binary"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUUIDString(t *testing.T) {
	uuid := UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	assert.Equal(t, "00010203-0405-0607-0809-0a0b0c0d0e0f", uuid.String())
}

func TestUUIDParse(t *testing.T) {
	parsed, err := ParseUUID("00010203-0405-0607-0809-0a0b0c0d0e0f")
	require.NoError(t, err)
	expected := UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	require.Equal(t, expected, parsed)

	samples := make([]UUID, 1000)
	for i := 0; i < 1000; i++ {
		var id UUID
		binary.BigEndian.PutUint64(id[:8], rand.Uint64())
		binary.BigEndian.PutUint64(id[8:], rand.Uint64())
		samples[i] = id
	}

	for _, id := range samples {
		t.Run(id.String(), func(t *testing.T) {
			parsed, err := ParseUUID(id.String())
			require.NoError(t, err)
			assert.Equal(t, id, parsed)
		})
	}
}

func TestUUIDMarshalJSON(t *testing.T) {
	uuid := UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	bts, err := json.Marshal(uuid)
	require.NoError(t, err)

	expected := `"00010203-0405-0607-0809-0a0b0c0d0e0f"`
	assert.Equal(t, expected, string(bts))
}

func TestUUIDUnmarshalJSON(t *testing.T) {
	str := `"00010203-0405-0607-0809-0a0b0c0d0e0f"`
	var uuid UUID
	err := json.Unmarshal([]byte(str), &uuid)
	require.NoError(t, err)

	expected := UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	assert.Equal(t, expected, uuid)
}

func TestUUIDUnmarshalJSONInvalid(t *testing.T) {
	samples := []string{
		`""`,
		`"000102030405060708090a0b0c0d0e"`,
		`"00010203-0405-060700809-0a0b0c0d0e0f"`,
		`"00010203-0405-06070-08090a0b0c-0d0e0f"`,
		`"zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz"`,
	}

	for _, s := range samples {
		t.Run(s, func(t *testing.T) {
			var uuid UUID
			err := json.Unmarshal([]byte(s), &uuid)
			assert.EqualError(t, err, "malformed gel.UUID")
		})
	}
}

func TestMarshalOptionalUUID(t *testing.T) {
	cases := []struct {
		input    OptionalUUID
		expected string
	}{
		{OptionalUUID{}, "null"},
		{
			OptionalUUID{
				UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
				true,
			},
			`"00010203-0405-0607-0809-0a0b0c0d0e0f"`,
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

func TestUnmarshalOptionalUUID(t *testing.T) {
	cases := []struct {
		expected OptionalUUID
		input    string
	}{
		{OptionalUUID{}, "null"},
		{
			OptionalUUID{
				UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
				true,
			},
			`"00010203-0405-0607-0809-0a0b0c0d0e0f"`,
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var empty OptionalUUID
			err := json.Unmarshal([]byte(c.input), &empty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, empty)

			notEmpty := OptionalUUID{
				UUID{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				true,
			}
			err = json.Unmarshal([]byte(c.input), &notEmpty)
			require.NoError(t, err)
			assert.Equal(t, c.expected, notEmpty)
		})
	}
}
