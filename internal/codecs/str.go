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
	strID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}
	strType = reflect.TypeOf("")
)

// StrMarshaler is the interface implemented by an object
// that can marshal itself into the str wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-str
//
// MarshalEdgeDBStr encodes the receiver
// into a binary form and returns the result.
type StrMarshaler interface {
	MarshalEdgeDBStr() ([]byte, error)
}

// StrUnmarshaler is the interface implemented by an object
// that can unmarshal the str wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-str
//
// UnmarshalEdgeDBStr must be able to decode the str wire format.
// UnmarshalEdgeDBStr must copy the data if it wishes to retain the data
// after returning.
type StrUnmarshaler interface {
	UnmarshalEdgeDBStr(data []byte) error
}

type strCodec struct {
	id types.UUID
}

func (c *strCodec) Type() reflect.Type { return strType }

func (c *strCodec) DescriptorID() types.UUID { return c.id }

func (c *strCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*string)(out) = string(r.Buf)
	r.Discard(len(r.Buf))
}

func (c *strCodec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case string:
		w.PushString(in)
	case StrMarshaler:
		data, err := in.MarshalEdgeDBStr()
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
