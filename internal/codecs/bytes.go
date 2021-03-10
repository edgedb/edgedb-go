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
	bytesID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2}
	bytesType = reflect.TypeOf([]byte{})

	// JSONBytes is a special case codec for json queries.
	// In go query json should return bytes not str.
	// but the descriptor type ID sent to the server
	// should still be str.
	JSONBytes = &bytesCodec{strID}
)

// BytesMarshaler is the interface implemented by an object
// that can marshal itself into the bytes wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bytes
//
// MarshalEdgeDBBytes encodes the receiver
// into a binary form and returns the result.
type BytesMarshaler interface {
	MarshalEdgeDBBytes() ([]byte, error)
}

// BytesUnmarshaler is the interface implemented by an object
// that can unmarshal the bytes wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-bytes
//
// UnmarshalEdgeDBBytes must be able to decode the bytes wire format.
// UnmarshalEdgeDBBytes must copy the data if it wishes to retain the data
// after returning.
type BytesUnmarshaler interface {
	UnmarshalEdgeDBBytes(data []byte) error
}

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

func (c *bytesCodec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case []byte:
		w.PushUint32(uint32(len(in)))
		w.PushBytes(in)
	case BytesMarshaler:
		data, err := in.MarshalEdgeDBBytes()
		if err != nil {
			return err
		}

		w.PushUint32(uint32(len(data)))
		w.PushBytes(data)
	default:
		return fmt.Errorf("expected %v to be []byte got %T", path, val)
	}

	return nil
}
