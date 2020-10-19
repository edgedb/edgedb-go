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

func TestDecodeTuple(t *testing.T) {
	bts := []byte{
		0, 0, 0, 36, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 2,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 3,
	}

	codec := &Tuple{[]DecodeEncoder{&Int64{}, &Int64{}}}
	result := codec.Decode(&bts)
	expected := types.Tuple{int64(2), int64(3)}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeNullTuple(t *testing.T) {
	bts := []byte{}
	(&Tuple{}).Encode(&bts, []interface{}{})

	expected := []byte{
		0, 0, 0, 4, // data length
		0, 0, 0, 0, // number of elements
	}

	assert.Equal(t, expected, bts)
}

func TestEncodeTuple(t *testing.T) {
	bts := []byte{}

	codec := &Tuple{[]DecodeEncoder{&Int64{}, &Int64{}}}
	codec.Encode(&bts, []interface{}{int64(2), int64(3)})

	expected := []byte{
		0, 0, 0, 36, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 2,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 3,
	}

	assert.Equal(t, expected, bts)
}
