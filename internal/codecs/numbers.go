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
	"math"
	"reflect"
	"unsafe"

	"github.com/geldata/gel-go/internal/buff"
	types "github.com/geldata/gel-go/internal/geltypes"
	"github.com/geldata/gel-go/internal/marshal"
)

// Int16Codec encodes/decodes int16.
type Int16Codec struct{}

// Type returns the type the codec encodes/decodes
func (c *Int16Codec) Type() reflect.Type { return int16Type }

// DescriptorID returns the codecs descriptor id.
func (c *Int16Codec) DescriptorID() types.UUID { return Int16ID }

// Decode decodes a value
func (c *Int16Codec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint16)(out) = r.PopUint16()
	return nil
}

type optionalInt16Marshaler interface {
	marshal.Int16Marshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *Int16Codec) Encode(
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
				return missingValueError("gel.OptionalInt16", path)
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
		return fmt.Errorf("expected %v to be int16, gel.OptionalInt16 or "+
			"Int16Marshaler got %T", path, val)
	}
}

func (c *Int16Codec) encodeData(w *buff.Writer, data int16) error {
	w.PushUint32(2)
	w.PushUint16(uint16(data))
	return nil
}

type optionalInt16 struct {
	val uint16
	set bool
}

type optionalInt16Decoder struct{}

func (c *optionalInt16Decoder) DescriptorID() types.UUID { return Int16ID }

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

func (c *optionalInt16Decoder) DecodePresent(_ unsafe.Pointer) {}

// Int32Codec encodes/decodes int32.
type Int32Codec struct{}

// Type returns the type the codec encodes/decodes
func (c *Int32Codec) Type() reflect.Type { return int32Type }

// DescriptorID returns the codecs descriptor id.
func (c *Int32Codec) DescriptorID() types.UUID { return Int32ID }

// Decode decodes a value
func (c *Int32Codec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint32)(out) = r.PopUint32()
	return nil
}

type optionalInt32Marshaler interface {
	marshal.Int32Marshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *Int32Codec) Encode(
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
				return missingValueError("gel.OptionalInt32", path)
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
		return fmt.Errorf("expected %v to be int32, gel.OptionalInt32 "+
			"or Int32Marshaler got %T", path, val)
	}
}

func (c *Int32Codec) encodeData(w *buff.Writer, data int32) error {
	w.PushUint32(4) // data length
	w.PushUint32(uint32(data))
	return nil
}

type optionalInt32 struct {
	val uint32
	set bool
}

type optionalInt32Decoder struct{}

func (c *optionalInt32Decoder) DescriptorID() types.UUID { return Int32ID }

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

func (c *optionalInt32Decoder) DecodePresent(_ unsafe.Pointer) {}

// Int64Codec encodes/decodes int64.
type Int64Codec struct{}

// Type returns the type the codec encodes/decodes
func (c *Int64Codec) Type() reflect.Type { return int64Type }

// DescriptorID returns the codecs descriptor id.
func (c *Int64Codec) DescriptorID() types.UUID { return Int64ID }

// Decode decodes a value
func (c *Int64Codec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint64)(out) = r.PopUint64()
	return nil
}

type optionalInt64Marshaler interface {
	marshal.Int64Marshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *Int64Codec) Encode(
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
				return missingValueError("gel.OptionalInt64", path)
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
		return fmt.Errorf("expected %v to be int64, gel.OptionalInt64 or "+
			"Int64Marshaler got %T", path, val)
	}
}

func (c *Int64Codec) encodeData(w *buff.Writer, data int64) error {
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

func (c *optionalInt64Decoder) DecodePresent(_ unsafe.Pointer) {}

// Float32Codec encodes/decodes float32.
type Float32Codec struct{}

// Type returns the type the codec encodes/decodes
func (c *Float32Codec) Type() reflect.Type { return float32Type }

// DescriptorID returns the codecs descriptor id.
func (c *Float32Codec) DescriptorID() types.UUID { return Float32ID }

// Decode decodes a value
func (c *Float32Codec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint32)(out) = r.PopUint32()
	return nil
}

type optionalFloat32Marshaler interface {
	marshal.Float32Marshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *Float32Codec) Encode(
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
				return missingValueError("gel.OptionalFloat32", path)
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
		return fmt.Errorf("expected %v to be float32, gel.OptionalFloat32 "+
			"or Float32Marshaler got %T", path, val)
	}
}

func (c *Float32Codec) encodeData(w *buff.Writer, data float32) error {
	w.PushUint32(4)
	w.PushUint32(math.Float32bits(data))
	return nil
}

type optionalFloat32 struct {
	val uint32
	set bool
}

type optionalFloat32Decoder struct{}

func (c *optionalFloat32Decoder) DescriptorID() types.UUID { return Float32ID }

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

func (c *optionalFloat32Decoder) DecodePresent(_ unsafe.Pointer) {}

// Float64Codec encodes/decodes float64.
type Float64Codec struct{}

// Type returns the type the codec encodes/decodes
func (c *Float64Codec) Type() reflect.Type { return float64Type }

// DescriptorID returns the codecs descriptor id.
func (c *Float64Codec) DescriptorID() types.UUID { return Float64ID }

// Decode decodes a value
func (c *Float64Codec) Decode(r *buff.Reader, out unsafe.Pointer) error {
	*(*uint64)(out) = r.PopUint64()
	return nil
}

type optionalFloat64Marshaler interface {
	marshal.Float64Marshaler
	marshal.OptionalMarshaler
}

// Encode encodes a value
func (c *Float64Codec) Encode(
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
				return missingValueError("gel.OptionalFloat64", path)
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
		return fmt.Errorf("expected %v to be float64, gel.OptionalFloat64 "+
			"or Float64Marshaler got %T", path, val)
	}
}

func (c *Float64Codec) encodeData(w *buff.Writer, data float64) error {
	w.PushUint32(8)
	w.PushUint64(math.Float64bits(data))
	return nil
}

type optionalFloat64 struct {
	val uint64
	set bool
}

type optionalFloat64Decoder struct{}

func (c *optionalFloat64Decoder) DescriptorID() types.UUID { return Float64ID }

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

func (c *optionalFloat64Decoder) DecodePresent(_ unsafe.Pointer) {}
