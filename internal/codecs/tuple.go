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
	"strconv"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
	"github.com/edgedb/edgedb-go/internal/marshal"
)

func popTupleCodec(
	r *buff.Reader,
	id types.UUID,
	codecs []Codec,
) Codec {
	elmCount := int(r.PopUint16())
	fields := make([]*objectField, elmCount)

	for i := 0; i < elmCount; i++ {
		index := r.PopUint16()

		fields[i] = &objectField{
			name:  strconv.Itoa(i),
			codec: codecs[index],
		}
	}

	return &Tuple{id: id, fields: fields}
}

// Tuple is an EdgeDB tuple type codec.
type Tuple struct {
	id     types.UUID
	fields []*objectField
	typ    reflect.Type
}

// ID returns the descriptor id.
func (c *Tuple) ID() types.UUID { return c.id }

func (c *Tuple) setType(typ reflect.Type, path Path) error {
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf(
			"expected %v to be a struct got %v", path, typ.Kind(),
		)
	}

	for _, field := range c.fields {
		sf, ok := marshal.StructField(typ, field.name)
		if !ok {
			return fmt.Errorf(
				"expected %v struct to have a %v field "+
					"with the tag `edgedb:\"%v\"`",
				typ,
				field.codec.Type(),
				field.name,
			)
		}

		err := field.codec.setType(sf.Type, path.AddField(field.name))
		if err != nil {
			return err
		}

		field.offset = sf.Offset
	}

	c.typ = typ
	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Tuple) Type() reflect.Type { return c.typ }

// Decode a tuple.
func (c *Tuple) Decode(r *buff.Reader, out unsafe.Pointer) {
	elmCount := int(int32(r.PopUint32()))
	if elmCount != len(c.fields) {
		panic(fmt.Sprintf(
			"wrong number of elements, expected %v got %v",
			len(c.fields), elmCount,
		))
	}

	for _, field := range c.fields {
		r.Discard(4) // reserved

		elmLen := r.PopUint32()
		if elmLen == 0xffffffff {
			continue
		}

		field.codec.Decode(r.PopSlice(elmLen), pAdd(out, field.offset))
	}
}

// Encode a tuple.
func (c *Tuple) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("expected %v to be []interface{} got %T", path, val)
	}

	if len(in) != len(c.fields) {
		return fmt.Errorf(
			"expected %v to be []interface{} with len=%v, got len=%v",
			path, len(c.fields), len(in),
		)
	}

	w.BeginBytes()

	elmCount := len(c.fields)
	w.PushUint32(uint32(elmCount))

	var err error
	for i, field := range c.fields {
		w.PushUint32(0) // reserved
		err = field.codec.Encode(w, in[i], path.AddIndex(i))
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
