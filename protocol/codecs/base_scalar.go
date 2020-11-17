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
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"time"
	"unsafe"

	"github.com/edgedb/edgedb-go/protocol/buff"
	"github.com/edgedb/edgedb-go/types"
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

func (c *UUID) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *UUID) Type() reflect.Type {
	return c.typ
}

// Decode a UUID.
func (c *UUID) Decode(buf *buff.Buff, out unsafe.Pointer) {
	buf.AssertAllocated(20)
	p := (*types.UUID)(out)
	copy((*p)[:], buf.Msg[4:20])
	buf.Msg = buf.Msg[20:]
}

// Encode a UUID.
func (c *UUID) Encode(buf *buff.Buff, val interface{}) {
	tmp := val.(types.UUID)
	buf.PushBytes(tmp[:])
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

func (c *Str) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Str) Type() reflect.Type {
	return c.typ
}

// Decode a string.
func (c *Str) Decode(buf *buff.Buff, out unsafe.Pointer) {
	*(*string)(out) = buf.PopString()
}

// Encode a string.
func (c *Str) Encode(buf *buff.Buff, val interface{}) {
	buf.PushString(val.(string))
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

func (c *Bytes) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Bytes) Type() reflect.Type {
	return c.typ
}

// Decode []byte.
func (c *Bytes) Decode(buf *buff.Buff, out unsafe.Pointer) {
	n := int(buf.PopUint32())
	buf.AssertAllocated(n)

	p := (*[]byte)(out)
	if cap(*p) >= n {
		*p = (*p)[:n]
	} else {
		*p = make([]byte, n)
	}

	copy(*p, buf.Msg[:n])
	buf.Msg = buf.Msg[n:]
}

// Encode []byte.
func (c *Bytes) Encode(buf *buff.Buff, val interface{}) {
	buf.PushBytes(val.([]byte))
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

func (c *Int16) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Int16) Type() reflect.Type {
	return c.typ
}

// Decode an int16.
func (c *Int16) Decode(buf *buff.Buff, out unsafe.Pointer) {
	buf.Discard(4) // data length
	*(*uint16)(out) = buf.PopUint16()
}

// Encode an int16.
func (c *Int16) Encode(buf *buff.Buff, val interface{}) {
	buf.PushUint32(2) // data length
	buf.PushUint16(uint16(val.(int16)))
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

func (c *Int32) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Int32) Type() reflect.Type {
	return c.typ
}

// Decode an int32.
func (c *Int32) Decode(buf *buff.Buff, out unsafe.Pointer) {
	*(*uint32)(out) = binary.BigEndian.Uint32(buf.Msg[4:8])
	buf.Msg = buf.Msg[8:]
}

// Encode an int32.
func (c *Int32) Encode(buf *buff.Buff, val interface{}) {
	buf.PushUint32(4) // data length
	buf.PushUint32(uint32(val.(int32)))
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

func (c *Int64) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Int64) Type() reflect.Type {
	return c.typ
}

// Decode an int64.
func (c *Int64) Decode(buf *buff.Buff, out unsafe.Pointer) {
	*(*uint64)(out) = binary.BigEndian.Uint64(buf.Msg[4:12])
	buf.Msg = buf.Msg[12:]
}

// Encode an int64.
func (c *Int64) Encode(buf *buff.Buff, val interface{}) {
	buf.PushUint32(8) // data length
	buf.PushUint64(uint64(val.(int64)))
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

func (c *Float32) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Float32) Type() reflect.Type {
	return c.typ
}

// Decode a float32.
func (c *Float32) Decode(buf *buff.Buff, out unsafe.Pointer) {
	*(*uint32)(out) = binary.BigEndian.Uint32(buf.Msg[4:8])
	buf.Msg = buf.Msg[8:]
}

// Encode a float32.
func (c *Float32) Encode(buf *buff.Buff, val interface{}) {
	buf.PushUint32(4)
	buf.PushUint32(math.Float32bits(val.(float32)))
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

func (c *Float64) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Float64) Type() reflect.Type {
	return c.typ
}

// Decode a float64.
func (c *Float64) Decode(buf *buff.Buff, out unsafe.Pointer) {
	*(*uint64)(out) = binary.BigEndian.Uint64(buf.Msg[4:12])
	buf.Msg = buf.Msg[12:]
}

// Encode a float64.
func (c *Float64) Encode(buf *buff.Buff, val interface{}) {
	buf.PushUint32(8)
	buf.PushUint64(math.Float64bits(val.(float64)))
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

func (c *Bool) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Bool) Type() reflect.Type {
	return c.typ
}

// Decode a bool.
func (c *Bool) Decode(buf *buff.Buff, out unsafe.Pointer) {
	buf.PopUint32() // data length
	*(*uint8)(out) = buf.PopUint8()
}

// Encode a bool.
func (c *Bool) Encode(buf *buff.Buff, val interface{}) {
	buf.PushUint32(1) // data length

	// convert bool to uint8
	var out uint8 = 0
	if val.(bool) {
		out = 1
	}

	buf.PushUint8(out)
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

func (c *DateTime) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *DateTime) Type() reflect.Type {
	return c.typ
}

// Decode a datetime.
func (c *DateTime) Decode(buf *buff.Buff, out unsafe.Pointer) {
	buf.PopUint32() // data length
	val := int64(buf.PopUint64())
	seconds := val / 1_000_000
	microseconds := val % 1_000_000
	*(*time.Time)(out) = time.Unix(
		946_684_800+seconds,
		1_000*microseconds,
	).UTC()
}

// Encode a datetime.
func (c *DateTime) Encode(buf *buff.Buff, val interface{}) {
	date := val.(time.Time)
	seconds := date.Unix() - 946_684_800
	nanoseconds := int64(date.Sub(time.Unix(date.Unix(), 0)))
	microseconds := seconds*1_000_000 + nanoseconds/1_000
	buf.PushUint32(8) // data length
	buf.PushUint64(uint64(microseconds))
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

func (c *Duration) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *Duration) Type() reflect.Type {
	return c.typ
}

// Decode a duration.
func (c *Duration) Decode(buf *buff.Buff, out unsafe.Pointer) {
	buf.PopUint32() // data length
	microseconds := int64(buf.PopUint64())
	buf.PopUint32() // reserved
	buf.PopUint32() // reserved
	*(*int64)(out) = microseconds * 1_000
}

// Encode a duration.
func (c *Duration) Encode(buf *buff.Buff, val interface{}) {
	duration := val.(time.Duration)
	buf.PushUint32(16) // data length
	buf.PushUint64(uint64(duration / 1_000))
	buf.PushUint32(0) // reserved
	buf.PushUint32(0) // reserved
}

// todo what type should JSON be decoded to :thinking:
// maybe don't encode/decode json? let the user wrangle bytes?

// JSON is an EdgeDB json type codec.
type JSON struct {
	id  types.UUID
	typ reflect.Type
}

// ID returns the descriptor id.
func (c *JSON) ID() types.UUID {
	return c.id
}

func (c *JSON) setType(typ reflect.Type) error {
	if typ != c.typ {
		return fmt.Errorf("expected %v got %v", c.typ, typ)
	}

	return nil
}

// Type returns the reflect.Type that this codec decodes to.
func (c *JSON) Type() reflect.Type {
	return c.typ
}

// Decode json.
func (c *JSON) Decode(buf *buff.Buff, out unsafe.Pointer) {
	buf.PopBytes()
}

// Encode json.
func (c *JSON) Encode(buf *buff.Buff, val interface{}) {
	bts, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}

	// prepend json format, always 1
	bts = append(bts, 0)
	copy(bts[1:], bts)
	bts[0] = 1

	buf.PushBytes(bts)
}
