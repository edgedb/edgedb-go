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

// todo improve tuple support  :thinking:

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/types"
)

func popTupleCodec(
	msg *buff.Message,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []Codec{}

	elmCount := int(msg.PopUint16())
	for i := 0; i < elmCount; i++ {
		index := msg.PopUint16()
		fields = append(fields, codecs[index])
	}

	// todo needs type
	return &Tuple{id: id, fields: fields}
}

// Tuple is an EdgeDB tuple type codec.
type Tuple struct {
	id     types.UUID
	fields []Codec
	typ    reflect.Type
	step   int
}

// ID returns the descriptor id.
func (c *Tuple) ID() types.UUID {
	return c.id
}

func (c *Tuple) setType(typ reflect.Type) error {
	if typ.Kind() != reflect.Slice {
		return fmt.Errorf("expected Slice got %v", typ.Kind())
	}

	if typ.Elem().Kind() != reflect.Interface {
		return fmt.Errorf("expected Interface got %v", typ.Elem().Kind())
	}

	for _, field := range c.fields {
		if field.Type() == nil {
			return errors.New("tuples may only contain base scalar types")
		}
	}

	c.typ = typ
	c.step = 16
	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Tuple) Type() reflect.Type {
	return c.typ
}

// Decode a tuple.
func (c *Tuple) Decode(msg *buff.Message, out unsafe.Pointer) {
	msg.Discard(4) // data length

	n := int(int32(msg.PopUint32()))
	slice := reflect.MakeSlice(c.typ, 0, n)

	for i := 0; i < n; i++ {
		msg.Discard(4) // reserved
		field := c.fields[i]
		val := reflect.New(field.Type()).Elem()
		field.Decode(msg, unsafe.Pointer(val.UnsafeAddr()))
		slice = reflect.Append(slice, val)
	}

	val := reflect.New(c.typ)
	val.Elem().Set(slice)
	*(*sliceHeader)(out) = *(*sliceHeader)(unsafe.Pointer(val.Pointer()))
}

// Encode a tuple.
func (c *Tuple) Encode(buf *buff.Writer, val interface{}) {
	buf.BeginBytes()

	elmCount := len(c.fields)
	buf.PushUint32(uint32(elmCount))

	in := val.([]interface{})
	for i := 0; i < elmCount; i++ {
		buf.PushUint32(0) // reserved
		c.fields[i].Encode(buf, in[i])
	}

	buf.EndBytes()
}
