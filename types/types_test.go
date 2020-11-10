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

package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUUIDString(t *testing.T) {
	uuid := UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	assert.Equal(t, "00010203-0405-0607-0809-0a0b0c0d0e0f", uuid.String())
}

func TestUUIDFromString(t *testing.T) {
	uuid, err := UUIDFromString("00010203-0405-0607-0809-0a0b0c0d0e0f")
	require.Nil(t, err)

	expected := UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	assert.Equal(t, expected, uuid)
}

func TestUUIDMarshalJSON(t *testing.T) {
	uuid := UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	bts, err := json.Marshal(uuid)
	require.Nil(t, err)

	expected := `"00010203-0405-0607-0809-0a0b0c0d0e0f"`
	assert.Equal(t, expected, string(bts))
}

func TestUUIDUnmarshalJSON(t *testing.T) {
	str := `"00010203-0405-0607-0809-0a0b0c0d0e0f"`
	var uuid UUID
	err := json.Unmarshal([]byte(str), &uuid)
	require.Nil(t, err)

	expected := UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	assert.Equal(t, expected, uuid)
}
