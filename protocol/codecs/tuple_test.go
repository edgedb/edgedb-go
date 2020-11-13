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

	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/types"
	"github.com/stretchr/testify/assert"
)

func TestDecodeTuple(t *testing.T) {
	buf := buff.NewMessage([]byte{
		0, 0, 0, 32, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 2,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4, // data length
		0, 0, 0, 3,
	})

	codec := &Tuple{fields: []Codec{
		&Int64{t: reflect.TypeOf(int64(0))},
		&Int32{t: reflect.TypeOf(int32(0))},
	}}

	var result []interface{}
	val := reflect.ValueOf(&result).Elem()
	codec.Decode(buf, val)

	expected := []interface{}{int64(2), int32(3)}
	assert.Equal(t, expected, result)
}

func TestEncodeNullTuple(t *testing.T) {
	buf := buff.NewWriter(nil)
	buf.BeginMessage(0xff)
	(&Tuple{}).Encode(buf, []interface{}{})
	buf.EndMessage()

	expected := []byte{
		0xff,         // message type
		0, 0, 0, 0xc, // message length
		0, 0, 0, 4, // data length
		0, 0, 0, 0, // number of elements
	}

	assert.Equal(t, expected, *buf.Unwrap())
}

func TestEncodeTuple(t *testing.T) {
	buf := buff.NewWriter(nil)
	buf.BeginMessage(0xff)

	codec := &Tuple{fields: []Codec{&Int64{}, &Int64{}}}
	codec.Encode(buf, []interface{}{int64(2), int64(3)})
	buf.EndMessage()

	expected := []byte{
		0xff,          // message type
		0, 0, 0, 0x2c, // message length
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

	assert.Equal(t, expected, *buf.Unwrap())
}

func BenchmarkEncodeTuple(b *testing.B) {
	codec := Tuple{fields: []Codec{&UUID{}}}
	id := types.UUID{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	ids := []interface{}{id}
	buf := buff.NewWriter(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.BeginMessage(0)
		codec.Encode(buf, ids)
	}
}
