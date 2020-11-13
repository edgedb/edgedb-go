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
	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/types"
)

func popObjectCodec(
	msg *buff.Message,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []*objectField{}

	elmCount := int(msg.PopUint16())
	for i := 0; i < elmCount; i++ {
		flags := msg.PopUint8()
		name := msg.PopString()
		index := msg.PopUint16()

		field := &objectField{
			isImplicit:     flags&0b1 != 0,
			isLinkProperty: flags&0b10 != 0,
			isLink:         flags&0b100 != 0,
			name:           name,
			codec:          codecs[index],
		}

		fields = append(fields, field)
	}

	// todo needs type
	return &Object{id: id, fields: fields}
}

type objectField struct {
	name           string
	index          []int
	codec          Codec
	isImplicit     bool
	isLinkProperty bool
	isLink         bool
}

// Object is an EdgeDB object type codec.
type Object struct {
	id     types.UUID
	fields []*objectField
	t      reflect.Type
}

// ID returns the descriptor id.
func (c *Object) ID() types.UUID {
	return c.id
}

func (c *Object) setType(t reflect.Type) error {
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("expected Struct got %v", t.Kind())
	}

	for _, field := range c.fields {
		if field.name == "__tid__" {
			continue
		}

		if f, ok := marshal.StructField(t, field.name); ok {
			field.index = f.Index
			if err := field.codec.setType(f.Type); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("%v struct is missing field %q", t, field.name)
		}
	}

	c.t = t
	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Object) Type() reflect.Type {
	return c.t
}

// Decode an object
func (c *Object) Decode(msg *buff.Message, out reflect.Value) {
	msg.PopUint32() // data length
	msg.PopUint32() // element count

	for _, field := range c.fields {
		msg.PopUint32() // reserved

		switch int32(msg.PeekUint32()) {
		case -1:
			// element length -1 means missing field
			// https://www.edgedb.com/docs/internals/protocol/dataformats
			msg.PopUint32()
		default:
			if field.name == "__tid__" {
				msg.PopBytes()
				break
			}

			field.codec.Decode(msg, out.FieldByIndex(field.index))
		}
	}
}

// Encode an object
func (c *Object) Encode(buf *buff.Writer, val interface{}) {
	panic("objects can't be query parameters")
}
