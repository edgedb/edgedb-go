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
)

var (
	uuidType = reflect.TypeOf(uuidID)
	uuidID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}
)

// UUIDMarshaler is the interface implemented by an object
// that can marshal itself into the uuid wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-uuid
//
// MarshalEdgeDBUUID encodes the receiver
// into a binary form and returns the result.
type UUIDMarshaler interface {
	MarshalEdgeDBUUID() ([]byte, error)
}

// UUIDUnmarshaler is the interface implemented by an object
// that can unmarshal the uuid wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-uuid
//
// UnmarshalEdgeDBUUID must be able to decode the uuid wire format.
// UnmarshalEdgeDBUUID must copy the data if it wishes to retain the data
// after returning.
type UUIDUnmarshaler interface {
	UnmarshalEdgeDBUUID(data []byte) error
}

type uuidCodec struct{}

func (c *uuidCodec) Type() reflect.Type { return uuidType }

func (c *uuidCodec) DescriptorID() types.UUID { return uuidID }

func (c *uuidCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	p := (*types.UUID)(out)
	copy((*p)[:], r.Buf[:16])
	r.Discard(16)
}

func (c *uuidCodec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case types.UUID:
		w.PushUint32(16)
		w.PushBytes(in[:])
	case UUIDMarshaler:
		data, err := in.MarshalEdgeDBUUID()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be edgedb.UUID got %T", path, val)
	}

	return nil
}
