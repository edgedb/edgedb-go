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
	boolID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 9}
	boolType = reflect.TypeOf(false)
)

// BoolMarshaler is the interface implemented by an object
// that can marshal itself into the bool wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bool
//
// MarshalEdgeDBBool encodes the receiver
// into a binary form and returns the result.
type BoolMarshaler interface {
	MarshalEdgeDBBool() ([]byte, error)
}

// BoolUnmarshaler is the interface implemented by an object
// that can unmarshal the bool wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bool
//
// UnmarshalEdgeDBBool must be able to decode the bool wire format.
// UnmarshalEdgeDBBool must copy the data if it wishes to retain the data
// after returning.
type BoolUnmarshaler interface {
	UnmarshalEdgeDBBool(data []byte) error
}

type boolCodec struct{}

func (c *boolCodec) Type() reflect.Type { return boolType }

func (c *boolCodec) DescriptorID() types.UUID { return boolID }

func (c *boolCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint8)(out) = r.PopUint8()
}

func (c *boolCodec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case bool:
		w.PushUint32(1) // data length

		// convert bool to uint8
		var out uint8 = 0
		if in {
			out = 1
		}

		w.PushUint8(out)
	case BoolMarshaler:
		data, err := in.MarshalEdgeDBBool()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be bool got %T", path, val)
	}

	return nil
}
