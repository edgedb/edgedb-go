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
	"unsafe"

	"github.com/edgedb/edgedb-go/marshal"
	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/types"
)

func popNamedTupleCodec(
	msg *buff.Message,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []*objectField{}

	elmCount := int(msg.PopUint16())
	for i := 0; i < elmCount; i++ {
		name := msg.PopString()
		index := msg.PopUint16()

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
	typ    reflect.Type
}

// ID returns the descriptor id.
func (c *NamedTuple) ID() types.UUID {
	return c.id
}

func (c *NamedTuple) setType(typ reflect.Type) error {
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("expected Struct got %v", typ.Kind())
	}

	for i := 0; i < len(c.fields); i++ {
		field := c.fields[i]

		if f, ok := marshal.StructField(typ, field.name); ok {
			field.offset = f.Offset
			if err := field.codec.setType(f.Type); err != nil {
				return err
			}
			continue
		}

		return fmt.Errorf("%v struct is missing field %q", typ, field.name)
	}

	c.typ = typ
	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *NamedTuple) Type() reflect.Type {
	return c.typ
}

// Decode a named tuple.
func (c *NamedTuple) Decode(msg *buff.Message, out unsafe.Pointer) {
	msg.PopUint32() // data length
	elmCount := int(int32(msg.PopUint32()))

	for i := 0; i < elmCount; i++ {
		msg.PopUint32() // reserved
		field := c.fields[i]
		field.codec.Decode(msg, pAdd(out, field.offset))
	}
}

// Encode a named tuple.
func (c *NamedTuple) Encode(buf *buff.Writer, val interface{}) {
	elmCount := len(c.fields)

	buf.BeginBytes()
	buf.PushUint32(uint32(elmCount))

	args := val.([]interface{})
	if len(args) != 1 {
		panic(fmt.Sprintf(
			"wrong number of arguments, expected 1 got: %v",
			args,
		))
	}

	in := args[0].(map[string]interface{})

	for i := 0; i < elmCount; i++ {
		buf.PushUint32(0) // reserved
		field := c.fields[i]
		field.codec.Encode(buf, in[field.name])
	}

	buf.EndBytes()
}
