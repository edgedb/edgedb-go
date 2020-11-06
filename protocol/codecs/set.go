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

	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

func popSetCodec(
	bts *[]byte,
	id types.UUID,
	codecs []Codec,
) Codec {
	n := protocol.PopUint16(bts)
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
func (c *Set) Decode(bts *[]byte, out reflect.Value) {
	buf := protocol.PopBytes(bts)

	dimCount := protocol.PopUint32(&buf) // number of dimensions, either 0 or 1
	if dimCount == 0 {
		return
	}

	protocol.PopUint32(&buf) // reserved
	protocol.PopUint32(&buf) // reserved

	upper := int32(protocol.PopUint32(&buf))
	lower := int32(protocol.PopUint32(&buf))
	n := int(upper - lower + 1)
	tmp := reflect.MakeSlice(c.t, n, n)

	for i := 0; i < n; i++ {
		c.child.Decode(&buf, tmp.Index(i))
	}

	out.Set(tmp)
}

// Encode a set
func (c *Set) Encode(bts *[]byte, val interface{}) {
	panic("not implemented")
}
