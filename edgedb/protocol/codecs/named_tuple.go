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
	"fmt"

	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

func popNamedTupleCodec(
	bts *[]byte,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []namedTupleField{}

	elmCount := int(protocol.PopUint16(bts))
	for i := 0; i < elmCount; i++ {
		name := protocol.PopString(bts)
		index := protocol.PopUint16(bts)

		field := namedTupleField{
			name:  name,
			codec: codecs[index],
		}

		fields = append(fields, field)
	}

	return &NamedTuple{idField{id}, fields}
}

type namedTupleField struct {
	name  string
	codec Codec
}

// NamedTuple is an EdgeDB namedtuple typep codec.
type NamedTuple struct {
	idField
	fields []namedTupleField
}

// Decode a named tuple.
func (c *NamedTuple) Decode(bts *[]byte) interface{} {
	buf := protocol.PopBytes(bts)

	elmCount := int(int32(protocol.PopUint32(&buf)))
	out := make(types.NamedTuple)

	for i := 0; i < elmCount; i++ {
		protocol.PopUint32(&buf) // reserved
		field := c.fields[i]
		out[field.name] = field.codec.Decode(&buf)
	}

	return out
}

// Encode a named tuple.
func (c *NamedTuple) Encode(bts *[]byte, val interface{}) {
	// don't know the data length yet
	// put everything in a new slice to get the length
	tmp := []byte{}

	elmCount := len(c.fields)
	protocol.PushUint32(&tmp, uint32(elmCount))

	args := val.([]interface{})
	if len(args) != 1 {
		panic(fmt.Sprintf("wrong number of arguments: %v", args))
	}

	in := args[0].(map[string]interface{})

	for i := 0; i < elmCount; i++ {
		protocol.PushUint32(&tmp, 0) // reserved
		field := c.fields[i]
		field.codec.Encode(&tmp, in[field.name])
	}

	protocol.PushUint32(bts, uint32(len(tmp)))
	*bts = append(*bts, tmp...)
}
