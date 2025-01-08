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

	"github.com/edgedb/edgedb-go/internal/buff"
	types "github.com/edgedb/edgedb-go/internal/geltypes"
	"github.com/edgedb/edgedb-go/internal/marshal"
)

// MemoryCodec encodes/decodes Memory values.
type MemoryCodec struct{}

// Type returns the type the codec encodes/decodes
func (c *MemoryCodec) Type() reflect.Type { return memoryType }

// DescriptorID returns the codecs descriptor id.
func (c *MemoryCodec) DescriptorID() types.UUID { return MemoryID }

// Decode decodes a value
func (c *MemoryCodec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint64)(out) = r.PopUint64()
	return nil
}

type optionalMemoryMarshaler interface {
	marshal.MemoryMarshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *MemoryCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case types.Memory:
		return c.encodeData(w, in)
	case types.OptionalMemory:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("gel.OptionalMemory", path)
			})
	case optionalMemoryMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error { return c.encodeMarshaler(w, in, path) },
			func() error { return missingValueError(in, path) })
	case marshal.MemoryMarshaler:
		return c.encodeMarshaler(w, in, path)
	default:
		return fmt.Errorf("expected %v to be gel.Memory, "+
			"gel.OptionalMemory or MemoryMarshaler got %T", path, val)
	}
}

func (c *MemoryCodec) encodeData(w *buff.Writer, data types.Memory) error {
	w.PushUint32(8) // data length
	w.PushUint64(uint64(data))
	return nil
}

func (c *MemoryCodec) encodeMarshaler(
	w *buff.Writer,
	val marshal.MemoryMarshaler,
	path Path,
) error {
	return encodeMarshaler(w, val, val.MarshalEdgeDBMemory, 8, path)
}

type optionalMemory struct {
	val uint64
	set bool
}

type optionalMemoryDecoder struct{}

func (c *optionalMemoryDecoder) DescriptorID() types.UUID { return MemoryID }

func (c *optionalMemoryDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	opint64 := (*optionalMemory)(out)
	opint64.val = r.PopUint64()
	opint64.set = true
	return nil
}

func (c *optionalMemoryDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalMemory)(out).Unset()
}

func (c *optionalMemoryDecoder) DecodePresent(_ unsafe.Pointer) {}
