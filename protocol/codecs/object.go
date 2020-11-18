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

func popObjectCodec(
	buf *buff.Buff,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []*objectField{}

	elmCount := int(buf.PopUint16())
	for i := 0; i < elmCount; i++ {
		flags := buf.PopUint8()
		name := buf.PopString()
		index := buf.PopUint16()

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
	offset         uintptr
	codec          Codec
	isImplicit     bool
	isLinkProperty bool
	isLink         bool
}

// Object is an EdgeDB object type codec.
type Object struct {
	id     types.UUID
	fields []*objectField
	typ    reflect.Type
}

// ID returns the descriptor id.
func (c *Object) ID() types.UUID {
	return c.id
}

func (c *Object) setType(typ reflect.Type) error {
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("expected Struct got %v", typ.Kind())
	}

	for _, field := range c.fields {
		if field.name == "__tid__" {
			continue
		}

		if f, ok := marshal.StructField(typ, field.name); ok {
			field.offset = f.Offset
			if err := field.codec.setType(f.Type); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("%v struct is missing field %q", typ, field.name)
		}
	}

	c.typ = typ
	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Object) Type() reflect.Type {
	return c.typ
}

// Decode an object
func (c *Object) Decode(buf *buff.Buff, out unsafe.Pointer) {
	buf.Discard(8) // data length & element count

	for _, field := range c.fields {
		buf.Discard(4) // reserved

		switch int32(buf.PeekUint32()) {
		case -1:
			// element length -1 means missing field
			// https://www.edgedb.com/docs/internals/protocol/dataformats
			buf.Discard(4)
		default:
			if field.name == "__tid__" {
				buf.Discard(20)
				break
			}

			p := pAdd(out, field.offset)
			field.codec.Decode(buf, p)
		}
	}
}

// Encode an object
func (c *Object) Encode(buf *buff.Buff, val interface{}) {
	panic("objects can't be query parameters")
}
