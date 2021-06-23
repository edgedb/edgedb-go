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
	"math"
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
	"github.com/edgedb/edgedb-go/internal/marshal"
)

var (
	int16ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3}
	int32ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 4}
	int64ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 5}
	float32ID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 6}
	float64ID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 7}

	int16Type           = reflect.TypeOf(int16(0))
	int32Type           = reflect.TypeOf(int32(0))
	int64Type           = reflect.TypeOf(int64(0))
	float32Type         = reflect.TypeOf(float32(0))
	float64Type         = reflect.TypeOf(float64(0))
	optionalInt16Type   = reflect.TypeOf(types.OptionalInt16{})
	optionalInt32Type   = reflect.TypeOf(types.OptionalInt32{})
	optionalInt64Type   = reflect.TypeOf(types.OptionalInt64{})
	optionalFloat32Type = reflect.TypeOf(types.OptionalFloat32{})
	optionalFloat64Type = reflect.TypeOf(types.OptionalFloat64{})
)

type int16Codec struct{}

func (c *int16Codec) Type() reflect.Type { return int16Type }

func (c *int16Codec) DescriptorID() types.UUID { return int16ID }

func (c *int16Codec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint16)(out) = r.PopUint16()
}

func (c *int16Codec) DecodeMissing(out unsafe.Pointer) { panic("unreachable") }

func (c *int16Codec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case int16:
		w.PushUint32(2) // data length
		w.PushUint16(uint16(in))
	case types.OptionalInt16:
		i, ok := in.Get()
		if !ok {
			return fmt.Errorf("cannot encode edgedb.OptionalInt16 at %v "+
				"because its value is missing", path)
		}

		w.PushUint32(2) // data length
		w.PushUint16(uint16(i))
	case marshal.Int16Marshaler:
		data, err := in.MarshalEdgeDBInt16()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be int16, edgedb.OptionalInt16 or "+
			"Int16Marshaler got %T", path, val)
	}

	return nil
}

type optionalInt16 struct {
	val uint16
	set bool
}

type optionalInt16Decoder struct{}

func (c *optionalInt16Decoder) DescriptorID() types.UUID { return int16ID }

func (c *optionalInt16Decoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	opint16 := (*optionalInt16)(out)
	opint16.val = r.PopUint16()
	opint16.set = true
}

func (c *optionalInt16Decoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalInt16)(out).Unset()
}

func (c *optionalInt16Decoder) DecodePresent(out unsafe.Pointer) {}

type int32Codec struct{}

func (c *int32Codec) Type() reflect.Type { return int32Type }

func (c *int32Codec) DescriptorID() types.UUID { return int32ID }

func (c *int32Codec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint32)(out) = r.PopUint32()
}

func (c *int32Codec) DecodeMissing(out unsafe.Pointer) { panic("unreachable") }

func (c *int32Codec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case int32:
		w.PushUint32(4) // data length
		w.PushUint32(uint32(in))
	case types.OptionalInt32:
		i, ok := in.Get()
		if !ok {
			return fmt.Errorf("cannot encode edgedb.OptionalInt32 at %v "+
				"because its value is missing", path)
		}

		w.PushUint32(4) // data length
		w.PushUint32(uint32(i))
	case marshal.Int32Marshaler:
		data, err := in.MarshalEdgeDBInt32()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be int32, edgedb.OptionalInt32 "+
			"or Int32Marshaler got %T", path, val)
	}

	return nil
}

type optionalInt32 struct {
	val uint32
	set bool
}

type optionalInt32Decoder struct{}

func (c *optionalInt32Decoder) DescriptorID() types.UUID { return int32ID }

func (c *optionalInt32Decoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	opint32 := (*optionalInt32)(out)
	opint32.val = r.PopUint32()
	opint32.set = true
}

func (c *optionalInt32Decoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalInt32)(out).Unset()
}

func (c *optionalInt32Decoder) DecodePresent(out unsafe.Pointer) {}

type int64Codec struct{}

func (c *int64Codec) Type() reflect.Type { return int64Type }

func (c *int64Codec) DescriptorID() types.UUID { return int64ID }

func (c *int64Codec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint64)(out) = r.PopUint64()
}

func (c *int64Codec) DecodeMissing(out unsafe.Pointer) { panic("unreachable") }

func (c *int64Codec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case int64:
		w.PushUint32(8) // data length
		w.PushUint64(uint64(in))
	case types.OptionalInt64:
		i, ok := in.Get()
		if !ok {
			return fmt.Errorf("cannot encode edgedb.OptionalInt64 at %v "+
				"because its value is missing", path)
		}

		w.PushUint32(8) // data length
		w.PushUint64(uint64(i))
	case marshal.Int64Marshaler:
		data, err := in.MarshalEdgeDBInt64()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be int64, edgedb.OptionalInt64 or "+
			"Int64Marshaler got %T", path, val)
	}

	return nil
}

type optionalInt64 struct {
	val uint64
	set bool
}

type optionalInt64Decoder struct{}

func (c *optionalInt64Decoder) DescriptorID() types.UUID { return int64ID }

func (c *optionalInt64Decoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	opint64 := (*optionalInt64)(out)
	opint64.val = r.PopUint64()
	opint64.set = true
}

func (c *optionalInt64Decoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalInt64)(out).Unset()
}

func (c *optionalInt64Decoder) DecodePresent(out unsafe.Pointer) {}

type float32Codec struct{}

func (c *float32Codec) Type() reflect.Type { return float32Type }

func (c *float32Codec) DescriptorID() types.UUID { return float32ID }

func (c *float32Codec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint32)(out) = r.PopUint32()
}

func (c *float32Codec) DecodeMissing(out unsafe.Pointer) {
	panic("unreachable")
}

func (c *float32Codec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	switch in := val.(type) {
	case float32:
		w.PushUint32(4)
		w.PushUint32(math.Float32bits(in))
	case types.OptionalFloat32:
		f, ok := in.Get()
		if !ok {
			return fmt.Errorf("cannot encode edgedb.OptionalFloat32 at %v "+
				"because its value is missing", path)
		}

		w.PushUint32(4)
		w.PushUint32(math.Float32bits(f))
	case marshal.Float32Marshaler:
		data, err := in.MarshalEdgeDBFloat32()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be float32, edgedb.OptionalFloat32 "+
			"or Float32Marshaler got %T", path, val)
	}

	return nil
}

type optionalFloat32 struct {
	val uint32
	set bool
}

type optionalFloat32Decoder struct{}

func (c *optionalFloat32Decoder) DescriptorID() types.UUID { return float32ID }

func (c *optionalFloat32Decoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	opint32 := (*optionalFloat32)(out)
	opint32.val = r.PopUint32()
	opint32.set = true
}

func (c *optionalFloat32Decoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalFloat32)(out).Unset()
}

func (c *optionalFloat32Decoder) DecodePresent(out unsafe.Pointer) {}

type float64Codec struct{}

func (c *float64Codec) Type() reflect.Type { return float64Type }

func (c *float64Codec) DescriptorID() types.UUID { return float64ID }

func (c *float64Codec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint64)(out) = r.PopUint64()
}

func (c *float64Codec) DecodeMissing(out unsafe.Pointer) {
	panic("unreachable")
}

func (c *float64Codec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
) error {
	switch in := val.(type) {
	case float64:
		w.PushUint32(8)
		w.PushUint64(math.Float64bits(in))
	case types.OptionalFloat64:
		f, ok := in.Get()
		if !ok {
			return fmt.Errorf("cannot encode edgedb.OptionalFloat64 at %v "+
				"because its value is missing", path)
		}
		w.PushUint32(8)
		w.PushUint64(math.Float64bits(f))
	case marshal.Float64Marshaler:
		data, err := in.MarshalEdgeDBFloat64()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be float64, edgedb.OptionalFloat64 "+
			"or Float64Marshaler got %T", path, val)
	}

	return nil
}

type optionalFloat64 struct {
	val uint64
	set bool
}

type optionalFloat64Decoder struct{}

func (c *optionalFloat64Decoder) DescriptorID() types.UUID { return float64ID }

func (c *optionalFloat64Decoder) Decode(r *buff.Reader, out unsafe.Pointer) {
	opint64 := (*optionalFloat64)(out)
	opint64.val = r.PopUint64()
	opint64.set = true
}

func (c *optionalFloat64Decoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalFloat64)(out).Unset()
}

func (c *optionalFloat64Decoder) DecodePresent(out unsafe.Pointer) {}
