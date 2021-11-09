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

var (
	boolID = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 9}
	boolType         = reflect.TypeOf(false)
	optionalBoolType = reflect.TypeOf(types.OptionalBool{})
)

type boolCodec struct{}

func (c *boolCodec) Type() reflect.Type { return boolType }

func (c *boolCodec) DescriptorID() types.UUID { return boolID }

func (c *boolCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint8)(out) = r.PopUint8()
	return nil
}

type optionalBoolMarshaler interface {
	marshal.BoolMarshaler
	marshal.OptionalMarshaler
}

func (c *boolCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case bool:
		return c.encodeData(w, in)
	case types.OptionalBool:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalBool", path)
			})
	case optionalBoolMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.BoolMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be bool, edgedb.OptionalBool or "+
			"BoolMarshaler got %T", path, val)
	}
}

func (c *boolCodec) encodeData(w *buff.Writer, data bool) error {
	w.PushUint32(1) // data length
	var out uint8
	if data {
		out = 1
	}
	w.PushUint8(out)
	return nil
}

func (c *boolCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.BoolMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBBool, 1, path)
}

type optionalBoolLayout struct {
	val uint8
	set bool
}

type optionalBoolDecoder struct{}

func (c *optionalBoolDecoder) DescriptorID() types.UUID { return boolID }

func (c *optionalBoolDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	opbool := (*optionalBoolLayout)(out)
	opbool.val = r.PopUint8()
	opbool.set = true
	return nil
}

func (c *optionalBoolDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalBool)(out).Unset()
}

func (c *optionalBoolDecoder) DecodePresent(out unsafe.Pointer) {}
