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
	boolID = types.UUID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 9}
	boolType         = reflect.TypeOf(false)
	optionalBoolType = reflect.TypeOf(types.OptionalBool{})
)

type boolCodec struct{}

func (c *boolCodec) Type() reflect.Type { return boolType }

func (c *boolCodec) DescriptorID() types.UUID { return boolID }

func (c *boolCodec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint8)(out) = r.PopUint8()
}

func (c *boolCodec) DecodeMissing(out unsafe.Pointer) { panic("unreachable") }

func (c *boolCodec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case bool:
		w.PushUint32(1) // data length

		// convert bool to uint8
		var out uint8
		if in {
			out = 1
		}

		w.PushUint8(out)
	case types.OptionalBool:
		b, ok := in.Get()
		if !ok {
			return fmt.Errorf(
				"cannot encode edgedb.OptionalBool at %v "+
					"because its value is missing", path)
		}

		w.PushUint32(1) // data length

		// convert bool to uint8
		var out uint8
		if b {
			out = 1
		}

		w.PushUint8(out)
	case marshal.BoolMarshaler:
		data, err := in.MarshalEdgeDBBool()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be bool, edgedb.OptionalBool or "+
			"BoolMarshaler got %T", path, val)
	}

	return nil
}

type optionalBoolLayout struct {
	val uint8
	set bool
}

type optionalBoolDecoder struct{}

func (c *optionalBoolDecoder) DescriptorID() types.UUID { return boolID }

func (c *optionalBoolDecoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	opbool := (*optionalBoolLayout)(out)
	opbool.val = r.PopUint8()
	opbool.set = true
}

func (c *optionalBoolDecoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalBool)(out).Unset()
}

func (c *optionalBoolDecoder) DecodePresent(out unsafe.Pointer) {}
