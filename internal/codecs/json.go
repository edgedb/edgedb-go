// This source file is part of the EdgeDB open source project.
//
// Copyright EdgeDB Inc. and the EdgeDB authors.
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
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/geldata/gel-go/internal/buff"
	types "github.com/geldata/gel-go/internal/geltypes"
	"github.com/geldata/gel-go/internal/marshal"
)

// JSONCodec encodes/decodes json.
type JSONCodec struct {
	baseJSONDecoder
	typ reflect.Type
}

// Type returns the type the codec encodes/decodes
func (c *JSONCodec) Type() reflect.Type { return bytesType }

// Decode decodes a value
func (c *JSONCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	if e := popJSONFormat(r); e != nil {
		return e
	}

	if c.typ != bytesType {
		ptr := reflect.NewAt(c.typ, out).Interface()
		return json.Unmarshal(r.Buf, ptr)
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
	return nil
}

type optionalJSONMarshaler interface {
	marshal.JSONMarshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *JSONCodec) Encode(
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
				return missingValueError("gel.OptionalBytes", path)
			})
	case optionalJSONMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.JSONMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be []byte, gel.OptionalBytes or "+
			"JSONMarshaler got %T", path, val)
	}
}

func (c *JSONCodec) encodeData(w *buff.Writer, data []byte) error {
	// data length
	w.PushUint32(uint32(1 + len(data)))

	// json format is always 1
	// https://www.edgedb.com/docs/internals/protocol/dataformats
	w.PushUint8(1)

	w.PushBytes(data)
	return nil
}

func (c *JSONCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.JSONMarshaler,
	_ Path,
) error {
	data, err := val.MarshalEdgeDBJSON()
	if err != nil {
		return err
	}
	w.PushUint32(uint32(len(data)))
	w.PushBytes(data)
	return nil
}

type baseJSONDecoder struct{}

func popJSONFormat(r *buff.Reader) error {
	format := r.PopUint8()
	if format != 1 {
		return fmt.Errorf(
			"unexpected json format: expected 1, got %v", format)
	}

	return nil
}

func (c *baseJSONDecoder) DescriptorID() types.UUID { return JSONID }

type optionalNilableJSONDecoder struct {
	baseJSONDecoder
	typ reflect.Type
}

func (c *optionalNilableJSONDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	if e := popJSONFormat(r); e != nil {
		return e
	}

	ptr := reflect.NewAt(c.typ, out).Interface()
	return json.Unmarshal(r.Buf, ptr)
}

func (c *optionalNilableJSONDecoder) DecodeMissing(out unsafe.Pointer) {
	val := reflect.NewAt(c.typ, out).Elem()
	if !val.IsZero() {
		val.Set(reflect.Zero(c.typ))
	}
}

type optionalUnmarshalerJSONDecoder struct {
	baseJSONDecoder
	typ reflect.Type
}

func (c *optionalUnmarshalerJSONDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	if e := popJSONFormat(r); e != nil {
		return e
	}

	ptr := reflect.NewAt(c.typ, out).Interface()
	ptr.(marshal.OptionalUnmarshaler).SetMissing(false)
	return json.Unmarshal(r.Buf, ptr)
}

func (c *optionalUnmarshalerJSONDecoder) DecodeMissing(out unsafe.Pointer) {
	ptr := reflect.NewAt(c.typ, out).Interface()
	ptr.(marshal.OptionalUnmarshaler).SetMissing(true)
}

type optionalScalarUnmarshalerJSONDecoder struct {
	baseJSONDecoder
	typ reflect.Type
}

func (c *optionalScalarUnmarshalerJSONDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	if e := popJSONFormat(r); e != nil {
		return e
	}

	ptr := reflect.NewAt(c.typ, out).Interface()
	return json.Unmarshal(r.Buf, ptr)
}

func (c *optionalScalarUnmarshalerJSONDecoder) DecodeMissing(
	out unsafe.Pointer,
) {
	ptr := reflect.NewAt(c.typ, out).Interface()
	ptr.(marshal.OptionalScalarUnmarshaler).Unset()
}

type optionalJSONDecoder struct {
	typ reflect.Type
}

func (c *optionalJSONDecoder) DescriptorID() types.UUID { return JSONID }

func (c *optionalJSONDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	if e := popJSONFormat(r); e != nil {
		return e
	}

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
	return nil
}

func (c *optionalJSONDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalBytes)(out).Unset()
}

func (c *optionalJSONDecoder) DecodePresent(_ unsafe.Pointer) {}
