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

func popObjectCodec(
	bts *[]byte,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []objectField{}

	elmCount := int(protocol.PopUint16(bts))
	for i := 0; i < elmCount; i++ {
		flags := protocol.PopUint8(bts)
		name := protocol.PopString(bts)
		index := protocol.PopUint16(bts)

		field := objectField{
			isImplicit:     flags&0b1 != 0,
			isLinkProperty: flags&0b10 != 0,
			isLink:         flags&0b100 != 0,
			name:           name,
			codec:          codecs[index],
		}

		fields = append(fields, field)
	}

	return &Object{idField{id}, fields}
}

// Object is an EdgeDB object type codec.
type Object struct {
	idField
	fields []objectField
}

type objectField struct {
	isImplicit     bool
	isLinkProperty bool
	isLink         bool
	name           string
	codec          Codec
}

// Decode an object
func (c *Object) Decode(bts *[]byte) interface{} {
	buf := protocol.PopBytes(bts)

	elmCount := int(int32(protocol.PopUint32(&buf)))
	out := make(types.Object)

	for i := 0; i < elmCount; i++ {
		protocol.PopUint32(&buf) // reserved
		field := c.fields[i]

		switch int32(protocol.PeekUint32(&buf)) {
		case -1:
			// element length -1 means missing field
			// https://www.edgedb.com/docs/internals/protocol/dataformats
			protocol.PopUint32(&buf)
			out[field.name] = types.Set{}
		default:
			out[field.name] = field.codec.Decode(&buf)
		}
	}

	return out
}

// Encode an object
func (c *Object) Encode(bts *[]byte, val interface{}) {
	panic("objects can't be query parameters")
}
