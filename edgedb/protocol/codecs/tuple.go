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
	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

func popTupleCodec(
	bts *[]byte,
	id types.UUID,
	codecs []DecodeEncoder,
) DecodeEncoder {
	fields := []DecodeEncoder{}

	elmCount := int(protocol.PopUint16(bts))
	for i := 0; i < elmCount; i++ {
		index := protocol.PopUint16(bts)
		fields = append(fields, codecs[index])
	}

	return &Tuple{idField{id}, fields}
}

// Tuple is an EdgeDB tuple type codec.
type Tuple struct {
	idField
	fields []DecodeEncoder
}

// Decode a tuple.
func (c *Tuple) Decode(bts *[]byte) interface{} {
	buf := protocol.PopBytes(bts)

	elmCount := int(int32(protocol.PopUint32(&buf)))
	out := make(types.Tuple, elmCount)

	for i := 0; i < elmCount; i++ {
		protocol.PopUint32(&buf) // reserved
		out[i] = c.fields[i].Decode(&buf)
	}

	return out
}

// Encode a tuple.
func (c *Tuple) Encode(bts *[]byte, val interface{}) {
	elmCount := len(c.fields)

	// special case for null tuple
	if elmCount == 0 {
		protocol.PushUint32(bts, 4) // data length
		protocol.PushUint32(bts, uint32(elmCount))
		return
	}

	tmp := []byte{}
	protocol.PushUint32(&tmp, uint32(elmCount))
	in := val.([]interface{})
	for i := 0; i < elmCount; i++ {
		protocol.PushUint32(&tmp, 0) // reserved
		c.fields[i].Encode(&tmp, in[i])
	}

	protocol.PushUint32(bts, uint32(len(tmp)))
	*bts = append(*bts, tmp...)
}
