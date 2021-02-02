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
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal/buff"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

var (
	uuidType    = reflect.TypeOf(uuidID)
	strType     = reflect.TypeOf("")
	bytesType   = reflect.TypeOf([]byte{})
	int16Type   = reflect.TypeOf(int16(0))
	int32Type   = reflect.TypeOf(int32(0))
	int64Type   = reflect.TypeOf(int64(0))
	float32Type = reflect.TypeOf(float32(0))
	float64Type = reflect.TypeOf(float64(0))
	boolType    = reflect.TypeOf(false)
	bigIntType  = reflect.TypeOf(&big.Int{})
)

var (
	uuidID    = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}
	strID     = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}
	bytesID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2}
	int16ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3}
	int32ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 4}
	int64ID   = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 5}
	float32ID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 6}
	float64ID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 7}
	decimalID = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 8}
	boolID    = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 9}
	jsonID    = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xf}
	bigIntID  = types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x10}
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
		return &UUID{}, nil
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
		return &LocalDateTime{}, nil
	case localDateID:
		return nil, errors.New("local_date not implemented")
	case localTimeID:
		return nil, errors.New("local_time not implemented")
	case durationID:
		return &Duration{id, durationType}, nil
	case jsonID:
		return &JSON{id: jsonID}, nil
	case bigIntID:
		return &BigInt{}, nil
	default:
		return nil, fmt.Errorf("unknown base scalar type id %v", id)
	}
}

// UUID is an EdgeDB UUID type codec.
type UUID struct{}

// ID returns the descriptor id.
func (c *UUID) ID() types.UUID { return uuidID }

func (c *UUID) setDefaultType() {}

func (c *UUID) setType(typ reflect.Type, path Path) (bool, error) {
	return false, c.checkType(typ, path)
}

func (c *UUID) checkType(typ reflect.Type, path Path) error {
	if typ != uuidType {
		return fmt.Errorf(
			"expected %v to be edgedb.UUID got %v", path, typ,
		)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *UUID) Type() reflect.Type { return uuidType }

// Decode a UUID.
func (c *UUID) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes a UUID using reflection
func (c *UUID) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if e := c.checkType(out.Type(), path); e != nil {
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
func (c *UUID) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.(types.UUID)
	if !ok {
		return fmt.Errorf("expected %v to be edgedb.UUID got %T", path, val)
	}

	w.PushUint32(16)
	w.PushBytes(in[:])
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

func (c *Str) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, typ,
		)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Str) Type() reflect.Type {
	return c.typ
}

// Decode a string.
func (c *Str) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes a str into a reflect.Value.
func (c *Str) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != c.typ {
		panic(fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a str into an unsafe.Pointer.
func (c *Str) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*string)(out) = string(r.Buf)
	r.Discard(len(r.Buf))
}

// Encode a string.
func (c *Str) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.(string)
	if !ok {
		return fmt.Errorf("expected %v to be edgedb.UUID got %T", path, val)
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

func (c *Bytes) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, typ,
		)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Bytes) Type() reflect.Type {
	return c.typ
}

// Decode []byte.
func (c *Bytes) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes bytes into a reflect.Value.
func (c *Bytes) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != c.typ {
		panic(fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, out.Type(),
		))
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
func (c *Bytes) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("expected %v to be []byte got %T", path, val)
	}

	w.PushUint32(uint32(len(in)))
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

func (c *Int16) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, typ,
		)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Int16) Type() reflect.Type {
	return c.typ
}

// Decode an int16.
func (c *Int16) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes an int16 into a reflect.Value.
func (c *Int16) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != c.typ {
		panic(fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes an int16 into an unsafe.Pointer.
func (c *Int16) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint16)(out) = r.PopUint16()
}

// Encode an int16.
func (c *Int16) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.(int16)
	if !ok {
		return fmt.Errorf("expected %v to be int16 got %T", path, val)
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

func (c *Int32) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, typ,
		)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Int32) Type() reflect.Type {
	return c.typ
}

// Decode an int32.
func (c *Int32) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes an int32 into a reflect.Value.
func (c *Int32) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != c.typ {
		panic(fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes an int32 into an unsafe.Pointer.
func (c *Int32) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint32)(out) = r.PopUint32()
}

// Encode an int32.
func (c *Int32) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.(int32)
	if !ok {
		return fmt.Errorf("expected %v to be int32 got %T", path, val)
	}

	w.PushUint32(4) // data length
	w.PushUint32(uint32(in))
	return nil
}

// Int64 is an EdgeDB int64 type codec.
type Int64 struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *Int64) ID() types.UUID {
	return c.id
}
func (c *Int64) setDefaultType() {}

func (c *Int64) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, typ,
		)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Int64) Type() reflect.Type {
	return c.typ
}

// Decode an int64.
func (c *Int64) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes an int64 into a reflect.Value.
func (c *Int64) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != c.typ {
		panic(fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes an int64 into an unsafe.Pointer.
func (c *Int64) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint64)(out) = r.PopUint64()
}

// Encode an int64.
func (c *Int64) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.(int64)
	if !ok {
		return fmt.Errorf("expected %v to be int64 got %T", path, val)
	}

	w.PushUint32(8) // data length
	w.PushUint64(uint64(in))
	return nil
}

// BigInt is and EdgeDB bigint type codec.
type BigInt struct{}

// ID returns the descriptor id.
func (c *BigInt) ID() types.UUID { return bigIntID }

// Type returns the reflect.Type that this codec decodes to.
func (c *BigInt) Type() reflect.Type { return bigIntType }

func (c *BigInt) setDefaultType() {}

func (c *BigInt) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != bigIntType {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, bigIntType, typ,
		)
	}

	return false, nil
}

// Decode a bigint.
func (c *BigInt) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes a bigint into a reflect.Value.
func (c *BigInt) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != bigIntType {
		panic(fmt.Errorf(
			"expected %v to be %v got %v", path, bigIntType, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

var (
	big10k  = big.NewInt(10_000)
	bigOne  = big.NewInt(1)
	bigZero = big.NewInt(0)
)

// DecodePtr decodes a bigint into an unsafe.Pointer.
func (c *BigInt) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	n := int(r.PopUint16())
	weight := big.NewInt(int64(r.PopUint16()))
	sign := r.PopUint16()
	r.Discard(2) // reserved

	result := (**big.Int)(out)
	if *result == nil {
		*result = &big.Int{}
	}

	digit := &big.Int{}
	shift := &big.Int{}

	for i := 0; i < n; i++ {
		shift.Exp(big10k, weight, nil)
		digit.SetBytes(r.Buf[:2])
		digit.Mul(digit, shift)
		(*result).Add(*result, digit)
		weight.Sub(weight, bigOne)
		r.Discard(2)
	}

	if sign == 0x4000 {
		(*result).Neg(*result)
	}
}

// Encode a bigint.
func (c *BigInt) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.(*big.Int)
	if !ok {
		return fmt.Errorf("expected %v to be *big.Int got %T", path, val)
	}

	// copy to prevent mutating the user's value
	cpy := &big.Int{}
	cpy.Set(in)

	var sign uint16 = 0
	if in.Sign() == -1 {
		sign = 0x4000
		cpy = cpy.Neg(cpy)
	}

	digits := []byte{}
	rem := &big.Int{}

	for cpy.CmpAbs(bigZero) != 0 {
		rem.Mod(cpy, big10k)

		// pad bytes
		bts := rem.Bytes()
		for len(bts) < 2 {
			bts = append([]byte{0}, bts...)
		}

		digits = append(bts, digits...)
		cpy = cpy.Div(cpy, big10k)
	}

	w.BeginBytes()
	w.PushUint16(uint16(len(digits) / 2))
	w.PushUint16(uint16(len(digits)/2 - 1))
	w.PushUint16(sign)
	w.PushUint16(0) // reserved
	w.PushBytes(digits)
	w.EndBytes()

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

func (c *Float32) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, typ,
		)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Float32) Type() reflect.Type {
	return c.typ
}

// Decode a float32.
func (c *Float32) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes a float32 into a reflect.Value.
func (c *Float32) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != c.typ {
		panic(fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a float32 into an unsafe.Pointer.
func (c *Float32) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint32)(out) = r.PopUint32()
}

// Encode a float32.
func (c *Float32) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.(float32)
	if !ok {
		return fmt.Errorf("expected %v to be float32 got %T", path, val)
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

func (c *Float64) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, typ,
		)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Float64) Type() reflect.Type {
	return c.typ
}

// Decode a float64.
func (c *Float64) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes a float64 into a reflect.Value.
func (c *Float64) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != c.typ {
		panic(fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a float64 into an unsafe.Pointer.
func (c *Float64) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint64)(out) = r.PopUint64()
}

// Encode a float64.
func (c *Float64) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.(float64)
	if !ok {
		return fmt.Errorf("expected %v to be float64 got %T", path, val)
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

func (c *Bool) setType(typ reflect.Type, path Path) (bool, error) {
	if typ != c.typ {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, typ,
		)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Bool) Type() reflect.Type {
	return c.typ
}

// Decode a bool.
func (c *Bool) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes a bool into a reflect.Value.
func (c *Bool) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != c.typ {
		panic(fmt.Errorf(
			"expected %v to be %v got %v", path, c.typ, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes a bool into an unsafe.Pointer.
func (c *Bool) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
	*(*uint8)(out) = r.PopUint8()
}

// Encode a bool.
func (c *Bool) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.(bool)
	if !ok {
		return fmt.Errorf("expected %v to be bool got %T", path, val)
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

// JSON is an EdgeDB json type codec.
type JSON struct {
	id types.UUID
}

// ID returns the descriptor id.
func (c *JSON) ID() types.UUID {
	return c.id
}

func (c *JSON) setDefaultType() {} // nolint:unused

func (c *JSON) setType( // nolint:unused
	typ reflect.Type,
	path Path,
) (bool, error) {
	if typ != bytesType {
		return false, fmt.Errorf(
			"expected %v to be %v got %v", path, bytesType, typ,
		)
	}

	return false, nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *JSON) Type() reflect.Type { return bytesType }

// Decode json.
func (c *JSON) Decode(r *buff.Reader, out reflect.Value) {
	c.DecodeReflect(r, out, Path(out.Type().String()))
}

// DecodeReflect decodes JSON into a reflect.Value.
func (c *JSON) DecodeReflect(r *buff.Reader, out reflect.Value, path Path) {
	if out.Type() != bytesType {
		panic(fmt.Errorf(
			"expected %v to be %v got %v", path, bytesType, out.Type(),
		))
	}

	c.DecodePtr(r, unsafe.Pointer(out.UnsafeAddr()))
}

// DecodePtr decodes JSON into an unsafe.Pointer.
func (c *JSON) DecodePtr(r *buff.Reader, out unsafe.Pointer) {
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

// Encode json.
func (c *JSON) Encode(w *buff.Writer, val interface{}, path Path) error {
	in, ok := val.([]byte)

	if !ok {
		return fmt.Errorf("expected %v to be []byte, got %T", path, val)
	}

	// data length
	w.PushUint32(uint32(1 + len(in)))

	// json format, always 1
	// https://www.edgedb.com/docs/internals/protocol/dataformats#std-json
	w.PushUint8(1)

	w.PushBytes(in)
	return nil
}
