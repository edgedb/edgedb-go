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
)

func getType(val interface{}) reflect.Type {
	return reflect.TypeOf(val).Elem()
}

var unmarshalers = map[reflect.Type]string{
	getType((*BoolUnmarshaler)(nil)):          "UnmarshalEdgeDBBool",
	getType((*BytesUnmarshaler)(nil)):         "UnmarshalEdgeDBBytes",
	getType((*DateTimeUnmarshaler)(nil)):      "UnmarshalEdgeDBDateTime",
	getType((*LocalDateTimeUnmarshaler)(nil)): "UnmarshalEdgeDBLocalDateTime",
	getType((*LocalDateUnmarshaler)(nil)):     "UnmarshalEdgeDBLocalDate",
	getType((*LocalTimeUnmarshaler)(nil)):     "UnmarshalEdgeDBLocalTime",
	getType((*DurationUnmarshaler)(nil)):      "UnmarshalEdgeDBDuration",
	getType(
		(*RelativeDurationUnmarshaler)(nil),
	): "UnmarshalEdgeDBRelativeDuration",
	getType((*JSONUnmarshaler)(nil)):    "UnmarshalEdgeDBJSON",
	getType((*Int16Unmarshaler)(nil)):   "UnmarshalEdgeDBInt16",
	getType((*Int32Unmarshaler)(nil)):   "UnmarshalEdgeDBInt32",
	getType((*Int64Unmarshaler)(nil)):   "UnmarshalEdgeDBInt64",
	getType((*Float32Unmarshaler)(nil)): "UnmarshalEdgeDBFloat32",
	getType((*Float64Unmarshaler)(nil)): "UnmarshalEdgeDBFloat64",
	getType((*BigIntUnmarshaler)(nil)):  "UnmarshalEdgeDBBigInt",
	getType((*DecimalUnmarshaler)(nil)): "UnmarshalEdgeDBDecimal",
	getType((*StrUnmarshaler)(nil)):     "UnmarshalEdgeDBStr",
	getType((*UUIDUnmarshaler)(nil)):    "UnmarshalEdgeDBUUID",
}

func buildUnmarshaler(
	desc descriptor.Descriptor,
	typ reflect.Type,
) (Decoder, bool) {
	for unmarshalerType, methodName := range unmarshalers {
		if reflect.PtrTo(typ).Implements(unmarshalerType) {
			return &unmarshalerDecoder{desc.ID, typ, methodName}, true
		}
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
