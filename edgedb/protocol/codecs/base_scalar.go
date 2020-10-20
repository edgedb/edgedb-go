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
	"fmt"
	"math"
	"time"

	"github.com/edgedb/edgedb-go/edgedb/protocol"
	"github.com/edgedb/edgedb-go/edgedb/types"
)

func getBaseScalarCodec(id types.UUID) DecodeEncoder {
	switch id {
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}:
		return &UUID{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}:
		return &String{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2}:
		return &Bytes{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3}:
		return &Int16{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 4}:
		return &Int32{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 5}:
		return &Int64{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 6}:
		return &Float32{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 7}:
		return &Float64{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 8}:
		panic("decimal type not implemented") // todo implement
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 9}:
		return &Bool{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xa}:
		return &DateTime{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xb}:
		panic("cal::local_datetime type not implemented") // todo implement
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xc}:
		panic("cal::local_date type not implemented") // todo implement
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xd}:
		panic("cal::local_time type not implemented") // todo implement
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xe}:
		return &Duration{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xf}:
		return &JSON{idField{id}}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x10}:
		panic("bigint type not implemented") // todo implement
	default:
		panic(fmt.Sprintf("unknown base scalar type descriptor id: % x", id))
	}
}

// UUID is an EdgeDB UUID type codec.
type UUID struct {
	idField
}

// Decode a UUID.
func (c *UUID) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	return protocol.PopUUID(bts)
}

// Encode a UUID.
func (c *UUID) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, uint32(16))
	tmp := val.(types.UUID)
	*bts = append(*bts, tmp[:]...)
}

// String is an EdgeDB string type codec.
type String struct {
	idField
}

// Decode a string.
func (c *String) Decode(bts *[]byte) interface{} {
	return protocol.PopString(bts)
}

// Encode a string.
func (c *String) Encode(bts *[]byte, val interface{}) {
	protocol.PushString(bts, val.(string))
}

// Bytes is an EdgeDB bytes type codec.
type Bytes struct {
	idField
}

// Decode []byte.
func (c *Bytes) Decode(bts *[]byte) interface{} {
	return protocol.PopBytes(bts)
}

// Encode []byte.
func (c *Bytes) Encode(bts *[]byte, val interface{}) {
	protocol.PushBytes(bts, val.([]byte))
}

// Int16 is an EdgeDB int64 type codec.
type Int16 struct {
	idField
}

// Decode an int16.
func (c *Int16) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	return int16(protocol.PopUint16(bts))
}

// Encode an int16.
func (c *Int16) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 2) // data length
	protocol.PushUint16(bts, uint16(val.(int16)))
}

// Int32 is an EdgeDB int32 type codec.
type Int32 struct {
	idField
}

// Decode an int32.
func (c *Int32) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	return int32(protocol.PopUint32(bts))
}

// Encode an int32.
func (c *Int32) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 4) // data length
	protocol.PushUint32(bts, uint32(val.(int32)))
}

// Int64 is an EdgeDB int64 typep codec.
type Int64 struct {
	idField
}

// Decode an int64.
func (c *Int64) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	return int64(protocol.PopUint64(bts))
}

// Encode an int64.
func (c *Int64) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 8) // data length
	protocol.PushUint64(bts, uint64(val.(int64)))
}

// Float32 is an EdgeDB float32 type codec.
type Float32 struct {
	idField
}

// Decode a float32.
func (c *Float32) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	bits := protocol.PopUint32(bts)
	return math.Float32frombits(bits)
}

// Encode a float32.
func (c *Float32) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 4)
	protocol.PushUint32(bts, math.Float32bits(val.(float32)))
}

// Float64 is an EdgeDB float64 type codec.
type Float64 struct {
	idField
}

// Decode a float64.
func (c *Float64) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	bits := protocol.PopUint64(bts)
	return math.Float64frombits(bits)
}

// Encode a float64.
func (c *Float64) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 8)
	protocol.PushUint64(bts, math.Float64bits(val.(float64)))
}

// Bool is an EdgeDB bool type codec.
type Bool struct {
	idField
}

// Decode a bool.
func (c *Bool) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	val := protocol.PopUint8(bts)
	if val > 1 {
		panic(fmt.Sprintf("invalid bool byte, must be 0 or 1, got: 0x%x", val))
	}
	return val != 0
}

// Encode a bool.
func (c *Bool) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 1) // data length

	// convert bool to uint8
	var out uint8 = 0
	if val.(bool) {
		out = 1
	}

	protocol.PushUint8(bts, out)
}

// DateTime is an EdgeDB datetime type codec.
type DateTime struct {
	idField
}

// Decode a datetime.
func (c *DateTime) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	val := int64(protocol.PopUint64(bts))
	seconds := val / 1_000_000
	microseconds := val % 1_000_000
	return time.Unix(946_684_800+seconds, 1_000*microseconds).UTC()
}

// Encode a datetime.
func (c *DateTime) Encode(bts *[]byte, val interface{}) {
	date := val.(time.Time)
	seconds := date.Unix() - 946_684_800
	nanoseconds := int64(date.Sub(time.Unix(date.Unix(), 0)))
	microseconds := seconds*1_000_000 + nanoseconds/1_000
	protocol.PushUint32(bts, 8) // data length
	protocol.PushUint64(bts, uint64(microseconds))
}

// Duration is an EdgeDB duration codec.
type Duration struct {
	idField
}

// Decode a duration.
func (c *Duration) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	microseconds := int64(protocol.PopUint64(bts))
	protocol.PopUint32(bts) // reserved
	protocol.PopUint32(bts) // reserved
	return time.Duration(microseconds * 1_000)
}

// Encode a duration.
func (c *Duration) Encode(bts *[]byte, val interface{}) {
	duration := val.(time.Duration)
	protocol.PushUint32(bts, 16) // data length
	protocol.PushUint64(bts, uint64(duration/1_000))
	protocol.PushUint32(bts, 0) // reserved
	protocol.PushUint32(bts, 0) // reserved
}

// JSON is an EdgeDB json type codec.
type JSON struct {
	idField
}

// Decode json.
func (c *JSON) Decode(bts *[]byte) interface{} {
	n := protocol.PopUint32(bts) // data length
	protocol.PopUint8(bts)       // json format, always 1

	var val interface{}
	err := json.Unmarshal((*bts)[:n-1], &val)
	if err != nil {
		panic(err)
	}

	*bts = (*bts)[n-1:]
	return val
}

// Encode json.
func (c *JSON) Encode(bts *[]byte, val interface{}) {
	buf, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}
	protocol.PushUint32(bts, uint32(1+len(buf))) // data length
	protocol.PushUint8(bts, 1)                   // json format, always 1
	*bts = append(*bts, buf...)
}
