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
	strID = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}
	strType         = reflect.TypeOf("")
	optionalStrType = reflect.TypeOf(types.OptionalStr{})
)

type strCodec struct {
	id types.UUID
}

func (c *strCodec) Type() reflect.Type { return strType }

func (c *strCodec) DescriptorID() types.UUID { return c.id }

func (c *strCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*string)(out) = string(r.Buf)
	r.Discard(len(r.Buf))
	return nil
}

type optionalStrMarshaler interface {
	marshal.StrMarshaler
	marshal.OptionalMarshaler
}

func (c *strCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case string:
		return c.encodeData(w, in)
	case types.OptionalStr:
		str, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, str) },
			func() error {
				return missingValueError("edgedb.OptionalStr", path)
			})
	case optionalStrMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.StrMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be string, edgedb.OptionalStr "+
			"or StrMarshaler got %T", path, val)
	}
}

func (c *strCodec) encodeData(w *buff.Writer, data string) error {
	w.PushString(data)
	return nil
}

func (c *strCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.StrMarshaler,
	path Path,
) error {
	data, err := val.MarshalEdgeDBStr()
	if err != nil {
		return err
	}
	w.PushUint32(uint32(len(data)))
	w.PushBytes(data)
	return nil
}

type optionalStr struct {
	val string
	set bool
}

type optionalStrDecoder struct {
	id types.UUID
}

func (c *optionalStrDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalStrDecoder) Decode(r *buff.Reader, out unsafe.Pointer) error {
	opstr := (*optionalStr)(out)
	opstr.val = string(r.Buf)
	opstr.set = true
	r.Discard(len(r.Buf))
	return nil
}

func (c *optionalStrDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalStr)(out).Unset()
}

func (c *optionalStrDecoder) DecodePresent(out unsafe.Pointer) {}
