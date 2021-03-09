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

	"github.com/edgedb/edgedb-go/internal/buff"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
	"github.com/edgedb/edgedb-go/internal/marshal"
)

func popNamedTupleCodec(
	r *buff.Reader,
	id types.UUID,
	codecs []Codec,
) Codec {
	fields := []*objectField{}

	elmCount := int(r.PopUint16())
	for i := 0; i < elmCount; i++ {
		name := r.PopString()
		index := r.PopUint16()

		if name == "__tid__" {
			continue
		}

		field := &objectField{
			name:  name,
			codec: codecs[index],
		}

		fields = append(fields, field)
	}

	return &NamedTuple{id: id, fields: fields}
}

// NamedTuple is an EdgeDB namedtuple type codec.
type NamedTuple struct {
	id     types.UUID
	fields []*objectField
	typ    reflect.Type
}

// ID returns the descriptor id.
func (c *NamedTuple) ID() types.UUID { return c.id }

func (c *NamedTuple) setType(typ reflect.Type, path Path) error {
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf(
			"expected %v to be a struct got %v", path, typ.Kind(),
		)
	}

	for _, field := range c.fields {
		f, ok := marshal.StructField(typ, field.name)
		if !ok {
			return fmt.Errorf("%v struct is missing field %q", typ, field.name)
		}

		err := field.codec.setType(
			f.Type,
			path.AddField(field.name),
		)

		if err != nil {
			return err
		}

		field.offset = f.Offset
	}

	c.typ = typ
	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *NamedTuple) Type() reflect.Type { return c.typ }

// Decode a named tuple.
func (c *NamedTuple) Decode(r *buff.Reader, out unsafe.Pointer) {
	elmCount := int(int32(r.PopUint32()))
	if elmCount != len(c.fields) {
		panic(fmt.Sprintf(
			"wrong number of elements expected %v got %v",
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

// Encode a named tuple.
func (c *NamedTuple) Encode(w *buff.Writer, val interface{}, path Path) error {
	args, ok := val.([]interface{})
	if !ok {
		return fmt.Errorf("expected %v to be []interface{} got %T", path, val)
	}

	elmCount := len(c.fields)

	w.BeginBytes()
	w.PushUint32(uint32(elmCount))

	if len(args) != 1 {
		return fmt.Errorf(
			"wrong number of arguments, expected 1 got: %v", len(args),
		)
	}

	in, ok := args[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf(
			"expected %v to be map[string]interface{} got %T", path, args[0],
		)
	}

	var err error
	for _, field := range c.fields {
		w.PushUint32(0) // reserved
		err = field.codec.Encode(w, in[field.name], path.AddField(field.name))
		if err != nil {
			return err
		}
	}

	w.EndBytes()
	return nil
}
