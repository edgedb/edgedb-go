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
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/geltypes"
	"github.com/edgedb/edgedb-go/internal/marshal"
)

func getType(val interface{}) reflect.Type {
	return reflect.TypeOf(val).Elem()
}

var unmarshalers = map[types.UUID]struct {
	typ        reflect.Type
	methodName string
}{
	BoolID: {
		typ:        getType((*marshal.BoolUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBBool",
	},
	BytesID: {
		typ:        getType((*marshal.BytesUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBBytes",
	},
	DateTimeID: {
		typ:        getType((*marshal.DateTimeUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBDateTime",
	},
	LocalDTID: {
		typ:        getType((*marshal.LocalDateTimeUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBLocalDateTime",
	},
	LocalDateID: {
		typ:        getType((*marshal.LocalDateUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBLocalDate",
	},
	LocalTimeID: {
		typ:        getType((*marshal.LocalTimeUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBLocalTime",
	},
	DurationID: {
		typ:        getType((*marshal.DurationUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBDuration",
	},
	RelativeDurationID: {
		typ:        getType((*marshal.RelativeDurationUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBRelativeDuration",
	},
	DateDurationID: {
		typ:        getType((*marshal.DateDurationUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBDateDuration",
	},
	JSONID: {
		typ:        getType((*marshal.JSONUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBJSON",
	},
	Int16ID: {
		typ:        getType((*marshal.Int16Unmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBInt16",
	},
	Int32ID: {
		typ:        getType((*marshal.Int32Unmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBInt32",
	},
	Int64ID: {
		typ:        getType((*marshal.Int64Unmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBInt64",
	},
	Float32ID: {
		typ:        getType((*marshal.Float32Unmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBFloat32",
	},
	Float64ID: {
		typ:        getType((*marshal.Float64Unmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBFloat64",
	},
	BigIntID: {
		typ:        getType((*marshal.BigIntUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBBigInt",
	},
	DecimalID: {
		typ:        getType((*marshal.DecimalUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBDecimal",
	},
	StrID: {
		typ:        getType((*marshal.StrUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBStr",
	},
	UUIDID: {
		typ:        getType((*marshal.UUIDUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBUUID",
	},
	MemoryID: {
		typ:        getType((*marshal.MemoryUnmarshaler)(nil)),
		methodName: "UnmarshalEdgeDBMemory",
	},
}

func buildUnmarshaler(
	desc descriptor.Descriptor,
	typ reflect.Type,
) (Decoder, bool, error) {
	var id types.UUID
	switch desc.Type {
	case descriptor.BaseScalar:
		id = desc.ID
	case descriptor.Enum:
		id = StrID
	default:
		return nil, false, fmt.Errorf(
			"unexpected descriptor type 0x%x", desc.Type)
	}

	iface, ok := unmarshalers[id]
	if !ok {
		return nil, false, nil
	}

	ptr := reflect.PointerTo(typ)
	if !ptr.Implements(iface.typ) {
		return nil, false, nil
	}

	var decoder = unmarshalerDecoder{desc.ID, typ, iface.methodName}

	if ptr.Implements(optionalUnmarshalerType) {
		return &optionalUnmarshalerDecoder{decoder}, true, nil
	}

	return &decoder, true, nil
}

func buildUnmarshalerV2(
	desc *descriptor.V2,
	typ reflect.Type,
) (Decoder, bool, error) {
	var id types.UUID
	switch desc.Type {
	case descriptor.Scalar:
		id = desc.ID
	case descriptor.Enum:
		id = StrID
	default:
		return nil, false, fmt.Errorf(
			"unexpected descriptor type 0x%x", desc.Type)
	}

	iface, ok := unmarshalers[id]
	if !ok {
		return nil, false, nil
	}

	ptr := reflect.PointerTo(typ)
	if !ptr.Implements(iface.typ) {
		return nil, false, nil
	}

	var decoder = unmarshalerDecoder{desc.ID, typ, iface.methodName}

	if ptr.Implements(optionalUnmarshalerType) {
		return &optionalUnmarshalerDecoder{decoder}, true, nil
	}

	return &decoder, true, nil
}

type unmarshalerDecoder struct {
	id         types.UUID
	typ        reflect.Type
	methodName string
}

func (c *unmarshalerDecoder) DescriptorID() types.UUID { return c.id }

func (c *unmarshalerDecoder) Decode(r *buff.Reader, out unsafe.Pointer) error {
	val := reflect.NewAt(c.typ, out)
	method := val.MethodByName(c.methodName)
	result := method.Call([]reflect.Value{reflect.ValueOf(r.Buf)})
	err := result[0].Interface()
	if err != nil {
		return err.(error)
	}
	return nil
}

type optionalUnmarshalerDecoder struct {
	unmarshalerDecoder
}

func (c *optionalUnmarshalerDecoder) DecodeMissing(out unsafe.Pointer) {
	val := reflect.NewAt(c.unmarshalerDecoder.typ, out)
	method := val.MethodByName("SetMissing")
	method.Call([]reflect.Value{trueValue})
}

func (c *optionalUnmarshalerDecoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	val := reflect.NewAt(c.unmarshalerDecoder.typ, out)
	method := val.MethodByName("SetMissing")
	method.Call([]reflect.Value{falseValue})
	return c.unmarshalerDecoder.Decode(r, out)
}
