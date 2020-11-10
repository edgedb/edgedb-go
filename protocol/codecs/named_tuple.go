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

	"github.com/edgedb/edgedb-go/marshal"
	"github.com/edgedb/edgedb-go/protocol"
	"github.com/edgedb/edgedb-go/types"
)

func popNamedTupleCodec(
	bts *[]byte,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []*objectField{}

	elmCount := int(protocol.PopUint16(bts))
	for i := 0; i < elmCount; i++ {
		name := protocol.PopString(bts)
		index := protocol.PopUint16(bts)

		if name == "__tid__" {
			continue
		}

		field := &objectField{
			name:  name,
			codec: codecs[index],
		}

		fields = append(fields, field)
	}

	// todo missing type
	return &NamedTuple{id: id, fields: fields}
}

// NamedTuple is an EdgeDB namedtuple type codec.
type NamedTuple struct {
	id     types.UUID
	fields []*objectField
	t      reflect.Type
}

// ID returns the descriptor id.
func (c *NamedTuple) ID() types.UUID {
	return c.id
}

func (c *NamedTuple) setType(t reflect.Type) error {
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("expected Struct got %v", t.Kind())
	}

	for i := 0; i < len(c.fields); i++ {
		field := c.fields[i]

		if f, ok := marshal.StructField(t, field.name); ok {
			field.index = f.Index
			if err := field.codec.setType(f.Type); err != nil {
				return err
			}
			continue
		}

		return fmt.Errorf("%v struct is missing field %q", t, field.name)
	}

	c.t = t
	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *NamedTuple) Type() reflect.Type {
	return c.t
}

// Decode a named tuple.
func (c *NamedTuple) Decode(bts *[]byte, out reflect.Value) {
	buf := protocol.PopBytes(bts)
	elmCount := int(int32(protocol.PopUint32(&buf)))

	for i := 0; i < elmCount; i++ {
		protocol.PopUint32(&buf) // reserved
		field := c.fields[i]
		val := out.FieldByIndex(field.index)
		field.codec.Decode(&buf, val)
	}
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
