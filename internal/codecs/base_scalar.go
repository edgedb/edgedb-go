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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"time"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/types"
)

var (
	uuidType     = reflect.TypeOf(uuidID)
	strType      = reflect.TypeOf("")
	bytesType    = reflect.TypeOf([]byte{})
	int16Type    = reflect.TypeOf(int16(0))
	int32Type    = reflect.TypeOf(int32(0))
	int64Type    = reflect.TypeOf(int64(0))
	float32Type  = reflect.TypeOf(float32(0))
	float64Type  = reflect.TypeOf(float64(0))
	boolType     = reflect.TypeOf(false)
	dateTimeType = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Second)
)

var (
	uuidID      = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}
	strID       = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}
	bytesID     = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2}
	int16ID     = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3}
	int32ID     = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 4}
	int64ID     = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 5}
	float32ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 6}
	float64ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 7}
	decimalID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 8}
	boolID      = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 9}
	dateTimeID  = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xa}
	localDTID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xb}
	localDateID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xc}
	localTimeID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xd}
	durationID  = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xe}
	jsonID      = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xf}
	bigIntID    = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x10}
)

var (
	// JSONBytes is a special case codec for json queries.
	// In go query json should return bytes not str.
	// but the descriptor type ID sent to the server
	// should still be str.
	JSONBytes = &Bytes{strID, bytesType}
)

func baseScalarCodec(id types.UUID) (Codec, error) {
	switch id {
	case uuidID:
		return &UUID{id, uuidType}, nil
	case strID:
		return &Str{id, strType}, nil
	case bytesID:
		return &Bytes{id, bytesType}, nil
	case int16ID:
		return &Int16{id, int16Type}, nil
	case int32ID:
		return &Int32{id, int32Type}, nil
	case int64ID:
		return &Int64{id, int64Type}, nil
	case float32ID:
		return &Float32{id, float32Type}, nil
	case float64ID:
		return &Float64{id, float64Type}, nil
	case decimalID:
		return nil, errors.New("decimal not implemented")
	case boolID:
		return &Bool{id, boolType}, nil
	case dateTimeID:
		return &DateTime{id, dateTimeType}, nil
	case localDTID:
		return nil, errors.New("local_datetime not implemented")
	case localDateID:
		return nil, errors.New("local_date not implemented")
	case localTimeID:
		return nil, errors.New("local_time not implemented")
	case durationID:
		return &Duration{id, durationType}, nil
	case jsonID:
		return nil, errors.New("JSON type not implemented")
	case bigIntID:
		return nil, errors.New("bigint not implemented")
	default:
		return nil, fmt.Errorf("unknown base scalar type id %v", id)
	}
}

// UUID is an EdgeDB UUID type codec.
type UUID struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *UUID) ID() types.UUID {
	return c.id
}

func (c *UUID) setDefaultType() {}

func (c *UUID) setType(typ reflect.Type) (bool, error) {
	return false, c.checkType(typ)
}

func (c *UUID) checkType(typ reflect.Type) error {
	switch {
	case typ.Kind() != c.typ.Kind():
		return fmt.Errorf("expected edgedb.UUID got %v", typ)
	case typ.Elem() != c.typ.Elem():
		return fmt.Errorf("expected edgedb.UUID got %v", typ)
	case typ.Len() != c.typ.Len():
		return fmt.Errorf("expected edgedb.UUID got %v", typ)
	case typ.PkgPath() != "github.com/edgedb/edgedb-go":
		return fmt.Errorf("expected edgedb.UUID got %v", typ)
	case typ.Name() != "UUID":
		return fmt.Errorf("expected edgedb.UUID got %v", typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *UUID) Type() reflect.Type {
	return c.typ
}

// Decode a UUID.
func (c *UUID) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes a UUID.using reflection
func (c *UUID) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if e := c.checkType(out.Type()); e != nil {
		panic(e)
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a UUID into an unsafe.Pointer.
func (c *UUID) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	p := (*types.UUID)(out)
	copy((*p)[:], r.Buf[:16])
	r.Discard(16)
}

// Encode a UUID.
func (c *UUID) Encode(w *buff.Writer, val interface{}) error {
	tmp, ok := val.(types.UUID)
	if !ok {
		return fmt.Errorf("expected types.UUID got %T", val)
	}

	w.PushBytes(tmp[:])
	return nil
}

// Str is an EdgeDB string type codec.
type Str struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *Str) ID() types.UUID {
	return c.id
}
func (c *Str) setDefaultType() {}

func (c *Str) setType(typ reflect.Type) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Str) Type() reflect.Type {
	return c.typ
}

// Decode a string.
func (c *Str) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes a str into a reflect.Value.
func (c *Str) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Errorf("expected %v got %v", c.typ, out.Type()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a str into an unsafe.Pointer.
func (c *Str) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*string)(out) = string(r.Buf)
	r.Discard(len(r.Buf))
}

// Encode a string.
func (c *Str) Encode(w *buff.Writer, val interface{}) error {
	in, ok := val.(string)
	if !ok {
		return fmt.Errorf("expected types.UUID got %T", val)
	}

	w.PushString(in)
	return nil
}

// Bytes is an EdgeDB bytes type codec.
type Bytes struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *Bytes) ID() types.UUID {
	return c.id
}
func (c *Bytes) setDefaultType() {}

func (c *Bytes) setType(typ reflect.Type) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Bytes) Type() reflect.Type {
	return c.typ
}

// Decode []byte.
func (c *Bytes) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes bytes into a reflect.Value.
func (c *Bytes) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Errorf("expected %v got %v", c.typ, out.Type()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes bytes into an unsafe.Pointer.
func (c *Bytes) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
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

// Encode []byte.
func (c *Bytes) Encode(w *buff.Writer, val interface{}) error {
	in, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte got %T", val)
	}

	w.PushBytes(in)
	return nil
}

// Int16 is an EdgeDB int64 type codec.
type Int16 struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *Int16) ID() types.UUID {
	return c.id
}
func (c *Int16) setDefaultType() {}

func (c *Int16) setType(typ reflect.Type) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Int16) Type() reflect.Type {
	return c.typ
}

// Decode an int16.
func (c *Int16) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes an int16 into a reflect.Value.
func (c *Int16) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Errorf("expected %v got %v", c.typ, out.Type()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes an int16 into an unsafe.Pointer.
func (c *Int16) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint16)(out) = r.PopUint16()
}

// Encode an int16.
func (c *Int16) Encode(w *buff.Writer, val interface{}) error {
	in, ok := val.(int16)
	if !ok {
		return fmt.Errorf("expected int16 got %T", val)
	}

	w.PushUint32(2) // data length
	w.PushUint16(uint16(in))
	return nil
}

// Int32 is an EdgeDB int32 type codec.
type Int32 struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *Int32) ID() types.UUID {
	return c.id
}
func (c *Int32) setDefaultType() {}

func (c *Int32) setType(typ reflect.Type) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Int32) Type() reflect.Type {
	return c.typ
}

// Decode an int32.
func (c *Int32) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes an int32 into a reflect.Value.
func (c *Int32) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Errorf("expected %v got %v", c.typ, out.Type()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes an int32 into an unsafe.Pointer.
func (c *Int32) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint32)(out) = r.PopUint32()
}

// Encode an int32.
func (c *Int32) Encode(w *buff.Writer, val interface{}) error {
	in, ok := val.(int32)
	if !ok {
		return fmt.Errorf("expected int32 got %T", val)
	}

	w.PushUint32(4) // data length
	w.PushUint32(uint32(in))
	return nil
}

// Int64 is an EdgeDB int64 typep codec.
type Int64 struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *Int64) ID() types.UUID {
	return c.id
}
func (c *Int64) setDefaultType() {}

func (c *Int64) setType(typ reflect.Type) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Int64) Type() reflect.Type {
	return c.typ
}

// Decode an int64.
func (c *Int64) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes an int64 into a reflect.Value.
func (c *Int64) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Errorf("expected %v got %v", c.typ, out.Type()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes an int64 into an unsafe.Pointer.
func (c *Int64) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint64)(out) = r.PopUint64()
}

// Encode an int64.
func (c *Int64) Encode(w *buff.Writer, val interface{}) error {
	in, ok := val.(int64)
	if !ok {
		return fmt.Errorf("expected int64 got %T", val)
	}

	w.PushUint32(8) // data length
	w.PushUint64(uint64(in))
	return nil
}

// Float32 is an EdgeDB float32 type codec.
type Float32 struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *Float32) ID() types.UUID {
	return c.id
}
func (c *Float32) setDefaultType() {}

func (c *Float32) setType(typ reflect.Type) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Float32) Type() reflect.Type {
	return c.typ
}

// Decode a float32.
func (c *Float32) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes a float32 into a reflect.Value.
func (c *Float32) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Errorf("expected %v got %v", c.typ, out.Type()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a float32 into an unsafe.Pointer.
func (c *Float32) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint32)(out) = r.PopUint32()
}

// Encode a float32.
func (c *Float32) Encode(w *buff.Writer, val interface{}) error {
	in, ok := val.(float32)
	if !ok {
		return fmt.Errorf("expected float32 got %T", val)
	}

	w.PushUint32(4)
	w.PushUint32(math.Float32bits(in))
	return nil
}

// Float64 is an EdgeDB float64 type codec.
type Float64 struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *Float64) ID() types.UUID {
	return c.id
}
func (c *Float64) setDefaultType() {}

func (c *Float64) setType(typ reflect.Type) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Float64) Type() reflect.Type {
	return c.typ
}

// Decode a float64.
func (c *Float64) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes a float64 into a reflect.Value.
func (c *Float64) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Errorf("expected %v got %v", c.typ, out.Type()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a float64 into an unsafe.Pointer.
func (c *Float64) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint64)(out) = r.PopUint64()
}

// Encode a float64.
func (c *Float64) Encode(w *buff.Writer, val interface{}) error {
	in, ok := val.(float64)
	if !ok {
		return fmt.Errorf("expected float64 got %T", val)
	}

	w.PushUint32(8)
	w.PushUint64(math.Float64bits(in))
	return nil
}

// Bool is an EdgeDB bool type codec.
type Bool struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *Bool) ID() types.UUID {
	return c.id
}
func (c *Bool) setDefaultType() {}

func (c *Bool) setType(typ reflect.Type) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Bool) Type() reflect.Type {
	return c.typ
}

// Decode a bool.
func (c *Bool) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes a bool into a reflect.Value.
func (c *Bool) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Errorf("expected %v got %v", c.typ, out.Type()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a bool into an unsafe.Pointer.
func (c *Bool) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint8)(out) = r.PopUint8()
}

// Encode a bool.
func (c *Bool) Encode(w *buff.Writer, val interface{}) error {
	in, ok := val.(bool)
	if !ok {
		return fmt.Errorf("expected bool got %T", val)
	}

	w.PushUint32(1) // data length

	// convert bool to uint8
	var out uint8 = 0
	if in {
		out = 1
	}

	w.PushUint8(out)
	return nil
}

// DateTime is an EdgeDB datetime type codec.
type DateTime struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *DateTime) ID() types.UUID {
	return c.id
}
func (c *DateTime) setDefaultType() {}

func (c *DateTime) setType(typ reflect.Type) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *DateTime) Type() reflect.Type {
	return c.typ
}

// Decode a datetime.
func (c *DateTime) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes a datetime into a reflect.Value.
func (c *DateTime) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Errorf("expected %v got %v", c.typ, out.Type()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a datetime into an unsafe.Pointer.
func (c *DateTime) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	val := int64(r.PopUint64())
	seconds := val / 1_000_000
	microseconds := val % 1_000_000
	*(*time.Time)(out) = time.Unix(
		946_684_800+seconds,
		1_000*microseconds,
	).UTC()
}

// Encode a datetime.
func (c *DateTime) Encode(w *buff.Writer, val interface{}) error {
	date, ok := val.(time.Time)
	if !ok {
		return fmt.Errorf("expected time.Time got %T", val)
	}

	seconds := date.Unix() - 946_684_800
	nanoseconds := int64(date.Sub(time.Unix(date.Unix(), 0)))
	microseconds := seconds*1_000_000 + nanoseconds/1_000
	w.PushUint32(8) // data length
	w.PushUint64(uint64(microseconds))
	return nil
}

// Duration is an EdgeDB duration codec.
type Duration struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *Duration) ID() types.UUID {
	return c.id
}

func (c *Duration) setDefaultType() {}
func (c *Duration) setType(typ reflect.Type) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Duration) Type() reflect.Type {
	return c.typ
}

// Decode a duration.
func (c *Duration) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes a duration into a reflect.Value.
func (c *Duration) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Errorf("expected %v got %v", c.typ, out.Type()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a duration into an unsafe.Pointer.
func (c *Duration) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	microseconds := int64(r.PopUint64())
	r.Discard(8) // reserved
	*(*int64)(out) = microseconds * 1_000
}

// Encode a duration.
func (c *Duration) Encode(w *buff.Writer, val interface{}) error {
	duration, ok := val.(time.Duration)
	if !ok {
		return fmt.Errorf("expected time.Duration got %T", val)
	}

	w.PushUint32(16) // data length
	w.PushUint64(uint64(duration / 1_000))
	w.PushUint32(0) // reserved
	w.PushUint32(0) // reserved
	return nil
}

// JSON is an EdgeDB json type codec.
type JSON struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *JSON) ID() types.UUID {
	return c.id
}

func (c *JSON) setDefaultType() {} // nolint:unused

func (c *JSON) setType(typ reflect.Type) (bool, error) { // nolint:unused
	if typ != c.typ {
		return false, fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *JSON) Type() reflect.Type {
	return c.typ
}

// Decode json.
func (c *JSON) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out)
}

// DecodeReflect decodes JSON into a reflect.Value.
func (c *JSON) DecodeReflect(r *buff.Reader, out reflect.Value) {
	if out.Type() != c.typ {
		panic(fmt.Errorf("expected %v got %v", c.typ, out.Type()))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes JSON into an unsafe.Pointer.
func (c *JSON) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	r.PopBytes()
}

// Encode json.
func (c *JSON) Encode(w *buff.Writer, val interface{}) {
	bts, err := json.Marshal(val)
	if err != nil {
		// todo err: Encode should return error?
		panic(err)
	}

	// prepend json format, always 1
	bts = append(bts, 0)
	copy(bts[1:], bts)
	bts[0] = 1

	w.PushBytes(bts)
}
