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

type int16Codec struct{}

func (c *int16Codec) Type() reflect.Type { return int16Type }

func (c *int16Codec) DescriptorID() types.UUID { return int16ID }

func (c *int16Codec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint16)(out) = r.PopUint16()
	return nil
}

type optionalInt16Marshaler interface {
	marshal.Int16Marshaler
	marshal.OptionalMarshaler
}

func (c *int16Codec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case int16:
		return c.encodeData(w, in)
	case types.OptionalInt16:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalInt16", path)
			})
	case optionalInt16Marshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error {
				return encodeMarshaler(w, in, in.MarshalEdgeDBInt16, 2, path)
			},
			func() error { return missingValueError(in, path) })
	case marshal.Int16Marshaler:
		return encodeMarshaler(w, in, in.MarshalEdgeDBInt16, 2, path)
	default:
		return fmt.Errorf("expected %v to be int16, edgedb.OptionalInt16 or "+
			"Int16Marshaler got %T", path, val)
	}
}

func (c *int16Codec) encodeData(w *buff.Writer, data int16) error {
	w.PushUint32(2)
	w.PushUint16(uint16(data))
	return nil
}

type optionalInt16 struct {
	val uint16
	set bool
}

type optionalInt16Decoder struct{}

func (c *optionalInt16Decoder) DescriptorID() types.UUID { return int16ID }

func (c *optionalInt16Decoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	opint16 := (*optionalInt16)(out)
	opint16.val = r.PopUint16()
	opint16.set = true
	return nil
}

func (c *optionalInt16Decoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalInt16)(out).Unset()
}

func (c *optionalInt16Decoder) DecodePresent(out unsafe.Pointer) {}

type int32Codec struct{}

func (c *int32Codec) Type() reflect.Type { return int32Type }

func (c *int32Codec) DescriptorID() types.UUID { return int32ID }

func (c *int32Codec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint32)(out) = r.PopUint32()
	return nil
}

type optionalInt32Marshaler interface {
	marshal.Int32Marshaler
	marshal.OptionalMarshaler
}

func (c *int32Codec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case int32:
		return c.encodeData(w, in)
	case types.OptionalInt32:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalInt32", path)
			})
	case optionalInt32Marshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error {
				return encodeMarshaler(w, in, in.MarshalEdgeDBInt32, 4, path)
			},
			func() error { return missingValueError(val, path) })
	case marshal.Int32Marshaler:
		return encodeMarshaler(w, in, in.MarshalEdgeDBInt32, 4, path)
	default:
		return fmt.Errorf("expected %v to be int32, edgedb.OptionalInt32 "+
			"or Int32Marshaler got %T", path, val)
	}
}

func (c *int32Codec) encodeData(w *buff.Writer, data int32) error {
	w.PushUint32(4) // data length
	w.PushUint32(uint32(data))
	return nil
}

type optionalInt32 struct {
	val uint32
	set bool
}

type optionalInt32Decoder struct{}

func (c *optionalInt32Decoder) DescriptorID() types.UUID { return int32ID }

func (c *optionalInt32Decoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	opint32 := (*optionalInt32)(out)
	opint32.val = r.PopUint32()
	opint32.set = true
	return nil
}

func (c *optionalInt32Decoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalInt32)(out).Unset()
}

func (c *optionalInt32Decoder) DecodePresent(out unsafe.Pointer) {}

type int64Codec struct{}

func (c *int64Codec) Type() reflect.Type { return int64Type }

func (c *int64Codec) DescriptorID() types.UUID { return Int64ID }

func (c *int64Codec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint64)(out) = r.PopUint64()
	return nil
}

type optionalInt64Marshaler interface {
	marshal.Int64Marshaler
	marshal.OptionalMarshaler
}

func (c *int64Codec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case int64:
		return c.encodeData(w, in)
	case types.OptionalInt64:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalInt64", path)
			})
	case optionalInt64Marshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error {
				return encodeMarshaler(w, in, in.MarshalEdgeDBInt64, 8, path)
			},
			func() error { return missingValueError(in, path) })
	case marshal.Int64Marshaler:
		return encodeMarshaler(w, in, in.MarshalEdgeDBInt64, 8, path)
	default:
		return fmt.Errorf("expected %v to be int64, edgedb.OptionalInt64 or "+
			"Int64Marshaler got %T", path, val)
	}
}

func (c *int64Codec) encodeData(w *buff.Writer, data int64) error {
	w.PushUint32(8) // data length
	w.PushUint64(uint64(data))
	return nil
}

type optionalInt64 struct {
	val uint64
	set bool
}

type optionalInt64Decoder struct{}

func (c *optionalInt64Decoder) DescriptorID() types.UUID { return Int64ID }

func (c *optionalInt64Decoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	opint64 := (*optionalInt64)(out)
	opint64.val = r.PopUint64()
	opint64.set = true
	return nil
}

func (c *optionalInt64Decoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalInt64)(out).Unset()
}

func (c *optionalInt64Decoder) DecodePresent(out unsafe.Pointer) {}

type float32Codec struct{}

func (c *float32Codec) Type() reflect.Type { return float32Type }

func (c *float32Codec) DescriptorID() types.UUID { return float32ID }

func (c *float32Codec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint32)(out) = r.PopUint32()
	return nil
}

type optionalFloat32Marshaler interface {
	marshal.Float32Marshaler
	marshal.OptionalMarshaler
}

func (c *float32Codec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case float32:
		return c.encodeData(w, in)
	case types.OptionalFloat32:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalFloat32", path)
			})
	case optionalFloat32Marshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error {
				return encodeMarshaler(w, in, in.MarshalEdgeDBFloat32, 4, path)
			},
			func() error { return missingValueError(val, path) })
	case marshal.Float32Marshaler:
		return encodeMarshaler(w, in, in.MarshalEdgeDBFloat32, 4, path)
	default:
		return fmt.Errorf("expected %v to be float32, edgedb.OptionalFloat32 "+
			"or Float32Marshaler got %T", path, val)
	}
}

func (c *float32Codec) encodeData(w *buff.Writer, data float32) error {
	w.PushUint32(4)
	w.PushUint32(math.Float32bits(data))
	return nil
}

type optionalFloat32 struct {
	val uint32
	set bool
}

type optionalFloat32Decoder struct{}

func (c *optionalFloat32Decoder) DescriptorID() types.UUID { return float32ID }

func (c *optionalFloat32Decoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	opint32 := (*optionalFloat32)(out)
	opint32.val = r.PopUint32()
	opint32.set = true
	return nil
}

func (c *optionalFloat32Decoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalFloat32)(out).Unset()
}

func (c *optionalFloat32Decoder) DecodePresent(out unsafe.Pointer) {}

type float64Codec struct{}

func (c *float64Codec) Type() reflect.Type { return float64Type }

func (c *float64Codec) DescriptorID() types.UUID { return float64ID }

func (c *float64Codec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint64)(out) = r.PopUint64()
	return nil
}

type optionalFloat64Marshaler interface {
	marshal.Float64Marshaler
	marshal.OptionalMarshaler
}

func (c *float64Codec) Encode(
	w *buff.Writer,
	val interface{},
	path Path,
	required bool,
) error {
	switch in := val.(type) {
	case float64:
		return c.encodeData(w, in)
	case types.OptionalFloat64:
		data, ok := in.Get()
		return encodeOptional(w, !ok, required,
			func() error { return c.encodeData(w, data) },
			func() error {
				return missingValueError("edgedb.OptionalFloat64", path)
			})
	case optionalFloat64Marshaler:
		return encodeOptional(w, in.Missing(), required,
			func() error {
				return encodeMarshaler(w, in, in.MarshalEdgeDBFloat64, 8, path)
			},
			func() error { return missingValueError(in, path) })
	case marshal.Float64Marshaler:
		return encodeMarshaler(w, in, in.MarshalEdgeDBFloat64, 8, path)
	default:
		return fmt.Errorf("expected %v to be float64, edgedb.OptionalFloat64 "+
			"or Float64Marshaler got %T", path, val)
	}
}

func (c *float64Codec) encodeData(w *buff.Writer, data float64) error {
	w.PushUint32(8)
	w.PushUint64(math.Float64bits(data))
	return nil
}

type optionalFloat64 struct {
	val uint64
	set bool
}

type optionalFloat64Decoder struct{}

func (c *optionalFloat64Decoder) DescriptorID() types.UUID { return float64ID }

func (c *optionalFloat64Decoder) Decode(
	r *buff.Reader,
	out unsafe.Pointer,
) error {
	opint64 := (*optionalFloat64)(out)
	opint64.val = r.PopUint64()
	opint64.set = true
	return nil
}

func (c *optionalFloat64Decoder) DecodeMissing(out unsafe.Pointer) {
	(*types.OptionalFloat64)(out).Unset()
}

func (c *optionalFloat64Decoder) DecodePresent(out unsafe.Pointer) {}
