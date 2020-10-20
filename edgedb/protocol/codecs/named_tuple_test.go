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

package codecs

import (
	"testing"

	"github.com/edgedb/edgedb-go/edgedb/types"
	"github.com/stretchr/testify/assert"
)

func TestDecodeNamedTuple(t *testing.T) {
	bts := []byte{
		0, 0, 0, 28, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 5,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 6,
	}

	codec := &NamedTuple{
		idField{},
		[]namedTupleField{
			{"a", &Int32{}},
			{"b", &Int32{}},
		},
	}

	result := codec.Decode(&bts)
	expected := types.NamedTuple{
		"a": int32(5),
		"b": int32(6),
	}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeNamedTuple(t *testing.T) {
	codec := &NamedTuple{
		idField{},
		[]namedTupleField{
			{"a", &Int32{}},
			{"b", &Int32{}},
		},
	}

	bts := []byte{}
	codec.Encode(&bts, []interface{}{map[string]interface{}{
		"a": int32(5),
		"b": int32(6),
	}})

	expected := []byte{
		0, 0, 0, 28, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 5,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 6,
	}

	assert.Equal(t, expected, bts)
}
