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
	uuidType         = reflect.TypeOf(uuidID)
	optionalUUIDType = reflect.TypeOf(types.OptionalUUID{})
	uuidID           = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}
)

type uuidCodec struct{}

func (c *uuidCodec) Type() reflect.Type { return uuidType }

func (c *uuidCodec) DescriptorID() types.UUID { return uuidID }

func (c *uuidCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	p := (*types.UUID)(out)
	copy((*p)[:], r.Buf[:16])
	r.Discard(16)
}

type optionalUUIDMarshaler interface {
	marshal.UUIDMarshaler
	marshal.OptionalMarshaler
}

func (c *uuidCodec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case types.UUID:
		return c.encodeData(w, in)
	case types.OptionalUUID:
		id, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, id) },
			func() error {
				return missingValueError("edgedb.OptionalUUID", path)
			})
	case optionalUUIDMarshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error {
				return encodeMarshaler(w, in, in.MarshalEdgeDBUUID, 16, path)
			},
			func() error { return missingValueError(in, path) })
	case marshal.UUIDMarshaler:
		return encodeMarshaler(w, in, in.MarshalEdgeDBUUID, 16, path)
	default:
		return fmt.Errorf("expected %v to be edgedb.UUID, "+
			"edgedb.OptionalUUID or UUIDMarshaler got %T", path, val)
	}
}

func (c *uuidCodec) encodeData(w *buff.Writer, data types.UUID) error {
	w.PushUint32(16)
	w.PushBytes(data[:])
	return nil
}

type optionalUUID struct {
	val types.UUID
	set bool
}

type optionalUUIDDecoder struct{}

func (c *optionalUUIDDecoder) DescriptorID() types.UUID { return uuidID }

func (c *optionalUUIDDecoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	opuuid := (*optionalUUID)(out)
	opuuid.set = true
	copy(opuuid.val[:], r.Buf[:16])
	r.Discard(16)
}

func (c *optionalUUIDDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalUUID)(out).Unset()
}

func (c *optionalUUIDDecoder) DecodePresent(out unsafe.Pointer) {}
