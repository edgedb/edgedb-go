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
	jsonID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xf}
)

type jsonCodec struct{}

func (c *jsonCodec) Type() reflect.Type { return bytesType }

func (c *jsonCodec) DescriptorID() types.UUID { return jsonID }

func (c *jsonCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	format := r.PopUint8()
	if format != 1 {
		panic(fmt.Sprintf(
			"unexpected json format: expected 1, got %v", format,
		))
	}

	n := len(r.Buf)
	p := (*[]byte)(out)
	if cap(*p) >= n {
		*p = (*p)[:n]
	} else {
		*p = make([]byte, n)
	}

	copy(*p, r.Buf)
	r.Discard(n)
}

type optionalJSONMarshaler interface {
	marshal.JSONMarshaler
	marshal.OptionalMarshaler
}

func (c *jsonCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case []byte:
		return c.encodeData(w, in)
	case types.OptionalBytes:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalBytes", path)
			})
	case optionalJSONMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.JSONMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be []byte, edgedb.OptionalBytes or "+
			"JSONMarshaler got %T", path, val)
	}
}

func (c *jsonCodec) encodeData(w *buff.Writer, data []byte) error {
	// data length
	w.PushUint32(uint32(1 + len(data)))

	// json format is always 1
	// https://www.edgedb.com/docs/internals/protocol/dataformats
	w.PushUint8(1)

	w.PushBytes(data)
	return nil
}

func (c *jsonCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.JSONMarshaler,
	path Path,
) error {
	data, err := val.MarshalEdgeDBJSON()
	if err != nil {
		return err
	}
	w.PushUint32(uint32(len(data)))
	w.PushBytes(data)
	return nil
}

type optionalJSONDecoder struct {
	id types.UUID
}

func (c *optionalJSONDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalJSONDecoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	opbytes := (*optionalBytesLayout)(out)
	opbytes.set = true

	n := len(r.Buf)
	if cap(opbytes.val) >= n {
		opbytes.val = (opbytes.val)[:n]
	} else {
		opbytes.val = make([]byte, n)
	}

	copy(opbytes.val, r.Buf)
	r.Discard(n)
}

func (c *optionalJSONDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalBytes)(out).Unset()
}

func (c *optionalJSONDecoder) DecodePresent(out unsafe.Pointer) {}
