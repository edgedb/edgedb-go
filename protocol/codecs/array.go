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
	"reflect"

	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/types"
)

func popArrayCodec(
	msg *buff.Message,
	id types.UUID,
	codecs []Codec,
) Codec {
	i := msg.PopUint16() // element type descriptor index

	n := int(msg.PopUint16()) // number of array dimensions
	for i := 0; i < n; i++ {
		msg.PopUint32() // array dimension
	}

	return &Array{id: id, child: codecs[i]}
}

// Array is an EdgeDB array type codec.
type Array struct {
	id    types.UUID
	child Codec
	t     reflect.Type
}

func (c *Array) setType(t reflect.Type) error {
	if t.Kind() != reflect.Slice {
		return fmt.Errorf("expected Slice got %v", t.Kind())
	}

	c.t = t
	return c.child.setType(t.Elem())
}

// ID returns the descriptor id.
func (c *Array) ID() types.UUID {
	return c.id
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Array) Type() reflect.Type {
	return c.child.Type()
}

// Decode an array.
func (c *Array) Decode(msg *buff.Message, out reflect.Value) {
	msg.PopUint32() // data length

	// number of dimensions is 1 or 0
	dimCount := msg.PopUint32()
	if dimCount == 0 {
		return
	}

	msg.PopUint32() // reserved
	msg.PopUint32() // reserved

	upper := int32(msg.PopUint32())
	lower := int32(msg.PopUint32())
	n := int(upper - lower + 1)

	// todo could reuse existing slice if it is long enough :thinking:
	tmp := reflect.MakeSlice(c.t, n, n)
	for i := 0; i < n; i++ {
		c.child.Decode(msg, tmp.Index(i))
	}

	out.Set(tmp)
}

// Encode an array.
func (c *Array) Encode(buf *buff.Writer, val interface{}) {
	in := val.([]interface{})
	elmCount := len(in)

	buf.BeginBytes()
	buf.PushUint32(1)                // number of dimensions
	buf.PushUint32(0)                // reserved
	buf.PushUint32(0)                // reserved
	buf.PushUint32(uint32(elmCount)) // dimension.upper
	buf.PushUint32(1)                // dimension.lower

	for i := 0; i < elmCount; i++ {
		c.child.Encode(buf, in[i])
	}

	buf.EndBytes()
}
