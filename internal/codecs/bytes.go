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
	bytesID = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2}
	bytesType         = reflect.TypeOf([]byte{})
	optionalBytesType = reflect.TypeOf(types.OptionalBytes{})

	// JSONBytes is a special case codec for json queries.
	// In go query json should return bytes not str.
	// but the descriptor type ID sent to the server
	// should still be str.
	JSONBytes = &bytesCodec{strID}
)

type bytesCodec struct {
	id types.UUID
}

func (c *bytesCodec) Type() reflect.Type { return bytesType }

func (c *bytesCodec) DescriptorID() types.UUID { return c.id }

func (c *bytesCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	n := len(r.Buf)

	p := (*[]byte)(out)
	if cap(*p) >= n {
		*p = (*p)[:n]
	} else {
		*p = make([]byte, n)
	}

	copy(*p, r.Buf)
	r.Discard(len(r.Buf))
}

type optionalBytesMarshaler interface {
	marshal.BytesMarshaler
	marshal.OptionalMarshaler
}

func (c *bytesCodec) Encode(
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
	case optionalBytesMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.BytesMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be []byte, edgedb.OptionalBytes or "+
			"BytesMarshaler got %T", path, val)
	}
}

func (c *bytesCodec) encodeData(w *buff.Writer, data []byte) error {
	w.PushUint32(uint32(len(data)))
	w.PushBytes(data)
	return nil
}

func (c *bytesCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.BytesMarshaler,
	path Path,
) error {
	data, err := val.MarshalEdgeDBBytes()
	if err != nil {
		return err
	}
	return c.encodeData(w, data)
}

type optionalBytesLayout struct {
	val []byte
	set bool
}

type optionalBytesDecoder struct {
	id types.UUID
}

func (c *optionalBytesDecoder) DescriptorID() types.UUID { return c.id }

func (c *optionalBytesDecoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	opbytes := (*optionalBytesLayout)(out)
	n := len(r.Buf)

	if cap(opbytes.val) >= n {
		opbytes.val = (opbytes.val)[:n]
	} else {
		opbytes.val = make([]byte, n)
	}

	copy(opbytes.val, r.Buf)
	opbytes.set = true
	r.Discard(len(r.Buf))
}

func (c *optionalBytesDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalBytes)(out).Unset()
}

func (c *optionalBytesDecoder) DecodePresent(out unsafe.Pointer) {}
