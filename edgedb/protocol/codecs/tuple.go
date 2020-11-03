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
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"

	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

func popTupleCodec(
	bts *[]byte,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []Codec{}

	elmCount := int(protocol.PopUint16(bts))
	for i := 0; i < elmCount; i++ {
		index := protocol.PopUint16(bts)
		fields = append(fields, codecs[index])
	}

	// todo needs type
	return &Tuple{id: id, fields: fields}
}

var interfaceSliceType reflect.Type = reflect.TypeOf([]interface{}{})

// Tuple is an EdgeDB tuple type codec.
type Tuple struct {
	id     types.UUID
	fields []Codec
	t      reflect.Type
}

func (c *Tuple) ID() types.UUID {
	return c.id
}

func (c *Tuple) setType(t reflect.Type) error {
	if t.Kind() != reflect.Slice {
		return fmt.Errorf(
			"out value does not match query schema: "+
				"expected Slice got %v",
			t.Kind(),
		)
	}

	if t.Elem().Kind() != reflect.Interface {
		return fmt.Errorf(
			"out value does not match query schema: "+
				"expected Interface got %v",
			t.Elem().Kind(),
		)
	}

	for _, field := range c.fields {
		if field.Type() == nil {
			return errors.New(
				"unsupported schema type: " +
					"tuples may only contain base scalar types",
			)
		}
	}

	c.t = t
	return nil
}

func (c *Tuple) Type() reflect.Type {
	return c.t
}

// Decode a tuple.
func (c *Tuple) Decode(bts *[]byte, out reflect.Value) error {
	buf := protocol.PopBytes(bts)
	n := int(int32(protocol.PopUint32(&buf)))
	tmp := reflect.MakeSlice(interfaceSliceType, 0, n)

	for i := 0; i < n; i++ {
		protocol.PopUint32(&buf) // reserved
		field := c.fields[i]
		val := reflect.New(field.Type()).Elem()
		field.Decode(&buf, val)
		tmp = reflect.Append(tmp, val)
	}

	out.Set(tmp)
	return nil
}

// Encode a tuple.
func (c *Tuple) Encode(bts *[]byte, val interface{}) error {
	p := len(*bts)

	// data length slot to be filled in at end
	*bts = append(*bts, 0, 0, 0, 0)

	elmCount := len(c.fields)
	protocol.PushUint32(bts, uint32(elmCount))

	in := val.([]interface{})
	for i := 0; i < elmCount; i++ {
		*bts = append(*bts, 0, 0, 0, 0) // reserved
		c.fields[i].Encode(bts, in[i])
	}

	n := len(*bts)
	binary.BigEndian.PutUint32((*bts)[p:], uint32(n-p-4))
	return nil
}
