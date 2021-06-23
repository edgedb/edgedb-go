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
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
	"github.com/edgedb/edgedb-go/internal/marshal"
)

func getType(val interface{}) reflect.Type {
	return reflect.TypeOf(val).Elem()
}

var unmarshalers = map[reflect.Type]string{
	getType((*marshal.BoolUnmarshaler)(nil)):  "UnmarshalEdgeDBBool",
	getType((*marshal.BytesUnmarshaler)(nil)): "UnmarshalEdgeDBBytes",
	getType(
		(*marshal.DateTimeUnmarshaler)(nil),
	): "UnmarshalEdgeDBDateTime",
	getType(
		(*marshal.LocalDateTimeUnmarshaler)(nil),
	): "UnmarshalEdgeDBLocalDateTime",
	getType((*marshal.LocalDateUnmarshaler)(nil)): "UnmarshalEdgeDBLocalDate",
	getType((*marshal.LocalTimeUnmarshaler)(nil)): "UnmarshalEdgeDBLocalTime",
	getType((*marshal.DurationUnmarshaler)(nil)):  "UnmarshalEdgeDBDuration",
	getType(
		(*marshal.RelativeDurationUnmarshaler)(nil),
	): "UnmarshalEdgeDBRelativeDuration",
	getType((*marshal.JSONUnmarshaler)(nil)):    "UnmarshalEdgeDBJSON",
	getType((*marshal.Int16Unmarshaler)(nil)):   "UnmarshalEdgeDBInt16",
	getType((*marshal.Int32Unmarshaler)(nil)):   "UnmarshalEdgeDBInt32",
	getType((*marshal.Int64Unmarshaler)(nil)):   "UnmarshalEdgeDBInt64",
	getType((*marshal.Float32Unmarshaler)(nil)): "UnmarshalEdgeDBFloat32",
	getType((*marshal.Float64Unmarshaler)(nil)): "UnmarshalEdgeDBFloat64",
	getType((*marshal.BigIntUnmarshaler)(nil)):  "UnmarshalEdgeDBBigInt",
	getType((*marshal.DecimalUnmarshaler)(nil)): "UnmarshalEdgeDBDecimal",
	getType((*marshal.StrUnmarshaler)(nil)):     "UnmarshalEdgeDBStr",
	getType((*marshal.UUIDUnmarshaler)(nil)):    "UnmarshalEdgeDBUUID",
}

var optionalUnmarshalerType = getType((*marshal.OptionalUnmarshaler)(nil))

func buildUnmarshaler(
	desc descriptor.Descriptor,
	typ reflect.Type,
) (Decoder, bool) {
	for unmarshalerType, methodName := range unmarshalers {
		ptr := reflect.PtrTo(typ)
		if !ptr.Implements(unmarshalerType) {
			continue
		}

		decoder := unmarshalerDecoder{desc.ID, typ, methodName}

		if ptr.Implements(optionalUnmarshalerType) {
			return &optionalUnmarshalerDecoder{decoder}, true
		}

		return &decoder, true
	}

	return nil, false
}

type unmarshalerDecoder struct {
	id         types.UUID
	typ        reflect.Type
	methodName string
}

func (c *unmarshalerDecoder) DescriptorID() types.UUID { return c.id }

func (c *unmarshalerDecoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	val := reflect.NewAt(c.typ, out)
	method := val.MethodByName(c.methodName)
	method.Call([]reflect.Value{reflect.ValueOf(r.Buf)})
}

func (c *unmarshalerDecoder) DecodeMissing(out unsafe.Pointer) {
	panic("unreachable")
}

type optionalUnmarshalerDecoder struct {
	unmarshalerDecoder
}

func (c *optionalUnmarshalerDecoder) DecodeMissing(out unsafe.Pointer) {
	val := reflect.NewAt(c.typ, out)
	method := val.MethodByName("SetMissing")
	method.Call([]reflect.Value{trueValue})
}

func (c *optionalUnmarshalerDecoder) DecodePresent(out unsafe.Pointer) {
	val := reflect.NewAt(c.typ, out)
	method := val.MethodByName("SetMissing")
	method.Call([]reflect.Value{falseValue})
}

func (c *optionalUnmarshalerDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) {
	c.DecodePresent(out)
	c.unmarshalerDecoder.Decode(r, out)
}
