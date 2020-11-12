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

func popSetCodec(
	msg *buff.Message,
	id types.UUID,
	codecs []Codec,
) Codec {
	n := msg.PopUint16()
	// todo type value
	return &Set{id: id, child: codecs[n]}
}

// Set is an EdgeDB set type codec.
type Set struct {
	id    types.UUID
	child Codec
	t     reflect.Type
}

// ID returns the descriptor id.
func (c *Set) ID() types.UUID {
	return c.id
}

func (c *Set) setType(t reflect.Type) error {
	if t.Kind() != reflect.Slice {
		return fmt.Errorf("expected Slice got %v", t.Kind())
	}

	c.t = t
	return c.child.setType(t.Elem())
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Set) Type() reflect.Type {
	return c.t
}

// Decode a set
func (c *Set) Decode(msg *buff.Message, out reflect.Value) {
	msg.PopUint32() // data length

	dimCount := msg.PopUint32() // number of dimensions, either 0 or 1
	if dimCount == 0 {
		msg.Discard(8) // skip 2 reserved fields
		return
	}

	msg.PopUint32() // reserved
	msg.PopUint32() // reserved

	upper := int32(msg.PopUint32())
	lower := int32(msg.PopUint32())
	n := int(upper - lower + 1)
	tmp := reflect.MakeSlice(c.t, n, n)

	for i := 0; i < n; i++ {
		c.child.Decode(msg, tmp.Index(i))
	}

	out.Set(tmp)
}

// Encode a set
func (c *Set) Encode(buf *buff.Writer, val interface{}) {
	panic("not implemented")
}
