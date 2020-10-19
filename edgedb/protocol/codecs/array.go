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

func popArrayCodec(
	bts *[]byte,
	id types.UUID,
	codecs []DecodeEncoder,
) DecodeEncoder {
	index := protocol.PopUint16(bts) // element type descriptor index

	n := int(protocol.PopUint16(bts)) // number of array dimensions
	for i := 0; i < n; i++ {
		protocol.PopUint32(bts) // array dimension
	}

	return &Array{codecs[index]}
}

// Array is an EdgeDB array type codec.
type Array struct {
	child DecodeEncoder
}

// Decode an array.
func (c *Array) Decode(bts *[]byte) interface{} {
	buf := protocol.PopBytes(bts)

	// number of dimensions is 1 or 0
	dimCount := protocol.PopUint32(&buf)
	if dimCount == 0 {
		return types.Array{}
	}

	protocol.PopUint32(&buf) // reserved
	protocol.PopUint32(&buf) // reserved

	upper := int32(protocol.PopUint32(&buf))
	lower := int32(protocol.PopUint32(&buf))
	elmCount := int(upper - lower + 1)

	out := make(types.Array, elmCount)
	for i := 0; i < elmCount; i++ {
		out[i] = c.child.Decode(&buf)
	}

	return out
}

// Encode an array.
func (c *Array) Encode(bts *[]byte, val interface{}) {
	// the data length is not know until all values have been encoded
	// put the data in temporary slice to get the length
	tmp := []byte{}

	in := val.([]interface{})
	elmCount := len(in)

	protocol.PushUint32(&tmp, 1)                // number of dimensions
	protocol.PushUint32(&tmp, 0)                // reserved
	protocol.PushUint32(&tmp, 0)                // reserved
	protocol.PushUint32(&tmp, uint32(elmCount)) // dimension.upper
	protocol.PushUint32(&tmp, 1)                // dimension.lower

	for i := 0; i < elmCount; i++ {
		c.child.Encode(&tmp, in[i])
	}

	protocol.PushUint32(bts, uint32(len(tmp)))
	*bts = append(*bts, tmp...)
}
