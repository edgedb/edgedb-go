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
	jsonID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xf}
)

// JSONMarshaler is the interface implemented by an object
// that can marshal itself into the json wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-json
//
// MarshalEdgeDBJSON encodes the receiver
// into a binary form and returns the result.
type JSONMarshaler interface {
	MarshalEdgeDBJSON() ([]byte, error)
}

// JSONUnmarshaler is the interface implemented by an object
// that can unmarshal the json wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-json
//
// UnmarshalEdgeDBJSON must be able to decode the json wire format.
// UnmarshalEdgeDBJSON must copy the data if it wishes to retain the data
// after returning.
type JSONUnmarshaler interface {
	UnmarshalEdgeDBJSON(data []byte) error
}

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

func (c *jsonCodec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case []byte:
		// data length
		w.PushUint32(uint32(1 + len(in)))

		// json format is always 1
		// https://www.edgedb.com/docs/internals/protocol/dataformats#std-json
		w.PushUint8(1)

		w.PushBytes(in)
	case JSONMarshaler:
		data, err := in.MarshalEdgeDBJSON()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be []byte, got %T", path, val)
	}

	return nil
}
