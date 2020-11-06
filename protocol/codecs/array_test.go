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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeArray(t *testing.T) {
	bts := []byte{
		0, 0, 0, 56, // data length
		0, 0, 0, 1, // dimension count
		0, 0, 0, 0, // reserved
		0, 0, 0, 0x14, // reserved
		0, 0, 0, 3, // dimension.upper
		0, 0, 0, 1, // dimension.lower
		// element 0
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 3, // ing64
		// element 1
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 5, // int64
		// element 2
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 8, // int64
	}

	codec := &Array{
		child: &Int64{},
		t:     reflect.TypeOf([]int64{}),
	}

	var result []int64
	val := reflect.ValueOf(&result).Elem()
	codec.Decode(&bts, val)

	expected := []int64{3, 5, 8}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeArray(t *testing.T) {
	bts := []byte{}
	codec := &Array{
		child: &Int64{},
		t:     nil,
	}
	codec.Encode(&bts, []interface{}{int64(3), int64(5), int64(8)})

	expected := []byte{
		0, 0, 0, 0x38, // data length
		0, 0, 0, 1, // dimension count
		0, 0, 0, 0, // reserved
		0, 0, 0, 0, // reserved
		0, 0, 0, 3, // dimension.upper
		0, 0, 0, 1, // dimension.lower
		// element 0
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 3, // ing64
		// element 1
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 5, // int64
		// element 2
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 8, // int64
	}

	assert.Equal(t, expected, bts)
}
