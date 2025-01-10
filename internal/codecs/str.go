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
	"fmt"
	"reflect"
	"unsafe"

	"github.com/geldata/gel-go/internal/buff"
	types "github.com/geldata/gel-go/internal/geltypes"
	"github.com/geldata/gel-go/internal/marshal"
)

// StrCodec encodes/decodes strings.
type StrCodec struct {
	ID types.UUID
}

// Type returns the type the codec encodes/decodes
func (c *StrCodec) Type() reflect.Type { return strType }

// DescriptorID returns the codecs descriptor id.
func (c *StrCodec) DescriptorID() types.UUID { return c.ID }

// Decode decodes a string.
func (c *StrCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*string)(out) = string(r.Buf)
	r.Discard(len(r.Buf))
	return nil
}

type optionalStrMarshaler interface {
	marshal.StrMarshaler
	marshal.OptionalMarshaler
}

// Encode encodes a string.
func (c *StrCodec) Encode(
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
				return missingValueError("gel.OptionalStr", path)
			})
	case optionalStrMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.StrMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be string, gel.OptionalStr "+
			"or StrMarshaler got %T", path, val)
	}
}

func (c *StrCodec) encodeData(w *buff.Writer, data string) error {
	w.PushString(data)
	return nil
}

func (c *StrCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.StrMarshaler,
	_ Path,
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

func (c *optionalStrDecoder) DecodePresent(_ unsafe.Pointer) {}
