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
)

func BenchmarkEncodeUUID(b *testing.B) {
	codec := &UUID{}
	id := types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	data := [2000]byte{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := data[:0]
		codec.Encode(&buf, id)
	}
}

func BenchmarkEncodeTuple(b *testing.B) {
	codec := Tuple{fields: []DecodeEncoder{&UUID{}}}
	id := types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	ids := []interface{}{id}
	data := [2000]byte{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := data[:0]
		codec.Encode(&buf, ids)
	}
}
