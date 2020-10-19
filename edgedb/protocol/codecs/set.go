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

func popSetCodec(
	bts *[]byte,
	id types.UUID,
	codecs []DecodeEncoder,
) DecodeEncoder {
	n := protocol.PopUint16(bts)
	return &Set{codecs[n]}
}

// Set is an EdgeDB set type codec.
type Set struct {
	child DecodeEncoder
}

// Decode a set
func (c *Set) Decode(bts *[]byte) interface{} {
	buf := protocol.PopBytes(bts)

	dimCount := protocol.PopUint32(&buf) // number of dimensions, either 0 or 1
	if dimCount == 0 {
		return types.Set{}
	}

	protocol.PopUint32(&buf) // reserved
	protocol.PopUint32(&buf) // reserved

	upper := int32(protocol.PopUint32(&buf))
	lower := int32(protocol.PopUint32(&buf))
	elmCount := int(upper - lower + 1)

	out := make(types.Set, elmCount)
	for i := 0; i < elmCount; i++ {
		out[i] = c.child.Decode(&buf)
	}

	return out
}

// Encode a set
func (c *Set) Encode(bts *[]byte, val interface{}) {
	panic("not implemented")
}
