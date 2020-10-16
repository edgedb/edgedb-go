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

package marshal

import (
	"reflect"
	"testing"

	"github.com/edgedb/edgedb-go/edgedb/types"
	"github.com/stretchr/testify/assert"
)

func TestMarshalSetOfScalar(t *testing.T) {
	var result interface{} = &[]int64{}
	input := types.Set{int64(3), int64(5), int64(8)}
	Marshal(&result, input)
	assert.Equal(t, []int64{3, 5, 8}, *(result.(*[]int64)))
}

func TestMarshalSetOfObject(t *testing.T) {
	type Database struct {
		Name string     `edgedb:"name"`
		ID   types.UUID `edgedb:"id"`
	}

	input := types.Set{
		types.Object{
			"name": "edgedb",
			"id":   types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		types.Object{
			"name": "tutorial",
			"id": types.UUID{
				1, 2, 3, 4, 5, 6, 7, 8,
				9, 10, 11, 12, 13, 14, 15, 16,
			},
		},
	}

	expected := []Database{
		{
			Name: "edgedb",
			ID:   types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			Name: "tutorial",
			ID: types.UUID{
				1, 2, 3, 4, 5, 6, 7, 8,
				9, 10, 11, 12, 13, 14, 15, 16,
			},
		},
	}

	var result interface{} = &[]Database{}
	Marshal(&result, input)
	assert.Equal(t, expected, *(result.(*[]Database)))
}

func TestSetNilScalar(t *testing.T) {
	var out int64
	in := types.Set{}

	setScalar(reflect.ValueOf(out), reflect.ValueOf(in))

	assert.Equal(t, int64(0), out)
}

func TestSetScalar(t *testing.T) {
	out := int64(0)
	in := int64(27)

	ov := reflect.ValueOf(&out)
	setScalar(ov.Elem(), reflect.ValueOf(in))

	assert.Equal(t, int64(27), out)
}
