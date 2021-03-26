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
)

var (
	int16ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3}
	int32ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 4}
	int64ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 5}
	float32ID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 6}
	float64ID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 7}

	int16Type   = reflect.TypeOf(int16(0))
	int32Type   = reflect.TypeOf(int32(0))
	int64Type   = reflect.TypeOf(int64(0))
	float32Type = reflect.TypeOf(float32(0))
	float64Type = reflect.TypeOf(float64(0))
)

// Int16Marshaler is the interface implemented by an object
// that can marshal itself into the int16 wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int16
//
// MarshalEdgeDBInt16 encodes the receiver
// into a binary form and returns the result.
type Int16Marshaler interface {
	MarshalEdgeDBInt16() ([]byte, error)
}

// Int16Unmarshaler is the interface implemented by an object
// that can unmarshal the int16 wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int16
//
// UnmarshalEdgeDBInt16 must be able to decode the int16 wire format.
// UnmarshalEdgeDBInt16 must copy the data if it wishes to retain the data
// after returning.
type Int16Unmarshaler interface {
	UnmarshalEdgeDBInt16(data []byte) error
}

type int16Codec struct{}

func (c *int16Codec) Type() reflect.Type { return int16Type }

func (c *int16Codec) DescriptorID() types.UUID { return int16ID }

func (c *int16Codec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint16)(out) = r.PopUint16()
}

func (c *int16Codec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case int16:
		w.PushUint32(2) // data length
		w.PushUint16(uint16(in))
	case Int16Marshaler:
		data, err := in.MarshalEdgeDBInt16()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be int16 got %T", path, val)
	}

	return nil
}

// Int32Marshaler is the interface implemented by an object
// that can marshal itself into the int32 wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int32
//
// MarshalEdgeDBInt32 encodes the receiver
// into a binary form and returns the result.
type Int32Marshaler interface {
	MarshalEdgeDBInt32() ([]byte, error)
}

// Int32Unmarshaler is the interface implemented by an object
// that can unmarshal the int32 wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int32
//
// UnmarshalEdgeDBInt32 must be able to decode the int32 wire format.
// UnmarshalEdgeDBInt32 must copy the data if it wishes to retain the data
// after returning.
type Int32Unmarshaler interface {
	UnmarshalEdgeDBInt32(data []byte) error
}

type int32Codec struct{}

func (c *int32Codec) Type() reflect.Type { return int32Type }

func (c *int32Codec) DescriptorID() types.UUID { return int32ID }

func (c *int32Codec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint32)(out) = r.PopUint32()
}

func (c *int32Codec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case int32:
		w.PushUint32(4) // data length
		w.PushUint32(uint32(in))
	case Int32Marshaler:
		data, err := in.MarshalEdgeDBInt32()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be int32 got %T", path, val)
	}

	return nil
}

// Int64Marshaler is the interface implemented by an object
// that can marshal itself into the int64 wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int64
//
// MarshalEdgeDBInt64 encodes the receiver
// into a binary form and returns the result.
type Int64Marshaler interface {
	MarshalEdgeDBInt64() ([]byte, error)
}

// Int64Unmarshaler is the interface implemented by an object
// that can unmarshal the int64 wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-int64
//
// UnmarshalEdgeDBInt64 must be able to decode the int64 wire format.
// UnmarshalEdgeDBInt64 must copy the data if it wishes to retain the data
// after returning.
type Int64Unmarshaler interface {
	UnmarshalEdgeDBInt64(data []byte) error
}

type int64Codec struct{}

func (c *int64Codec) Type() reflect.Type { return int64Type }

func (c *int64Codec) DescriptorID() types.UUID { return int64ID }

func (c *int64Codec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint64)(out) = r.PopUint64()
}

func (c *int64Codec) Encode(w *buff.Writer, val interface{}, path Path) error {
	switch in := val.(type) {
	case int64:
		w.PushUint32(8) // data length
		w.PushUint64(uint64(in))
	case Int64Marshaler:
		data, err := in.MarshalEdgeDBInt64()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be int64 got %T", path, val)
	}

	return nil
}

// Float32Marshaler is the interface implemented by an object
// that can marshal itself into the float32 wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-float32
//
// MarshalEdgeDBFloat32 encodes the receiver
// into a binary form and returns the result.
type Float32Marshaler interface {
	MarshalEdgeDBFloat32() ([]byte, error)
}

// Float32Unmarshaler is the interface implemented by an object
// that can unmarshal the float32 wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-float32
//
// UnmarshalEdgeDBFloat32 must be able to decode the float32 wire format.
// UnmarshalEdgeDBFloat32 must copy the data if it wishes to retain the data
// after returning.
type Float32Unmarshaler interface {
	UnmarshalEdgeDBFloat32(data []byte) error
}

type float32Codec struct{}

func (c *float32Codec) Type() reflect.Type { return float32Type }

func (c *float32Codec) DescriptorID() types.UUID { return float32ID }

func (c *float32Codec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint32)(out) = r.PopUint32()
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
	case Float32Marshaler:
		data, err := in.MarshalEdgeDBFloat32()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be float32 got %T", path, val)
	}

	return nil
}

// Float64Marshaler is the interface implemented by an object
// that can marshal itself into the float64 wire format.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-float64
//
// MarshalEdgeDBFloat64 encodes the receiver
// into a binary form and returns the result.
type Float64Marshaler interface {
	MarshalEdgeDBFloat64() ([]byte, error)
}

// Float64Unmarshaler is the interface implemented by an object
// that can unmarshal the float64 wire format representation of itself.
// https://www.edgedb.com/docs/internals/protocol/dataformats#std-float64
//
// UnmarshalEdgeDBFloat64 must be able to decode the float64 wire format.
// UnmarshalEdgeDBFloat64 must copy the data if it wishes to retain the data
// after returning.
type Float64Unmarshaler interface {
	UnmarshalEdgeDBFloat64(data []byte) error
}

type float64Codec struct{}

func (c *float64Codec) Type() reflect.Type { return float64Type }

func (c *float64Codec) DescriptorID() types.UUID { return float64ID }

func (c *float64Codec) Decode(r *buff.Reader, out unsafe.Pointer) {
	*(*uint64)(out) = r.PopUint64()
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
	case Float64Marshaler:
		data, err := in.MarshalEdgeDBFloat64()
		if err != nil {
			return err
		}

		w.BeginBytes()
		w.PushBytes(data)
		w.EndBytes()
	default:
		return fmt.Errorf("expected %v to be float64 got %T", path, val)
	}

	return nil
}
