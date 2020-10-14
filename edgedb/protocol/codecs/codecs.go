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

const (
	setType = iota
	objectType
	baseScalarType
	scalarType
	tupleType
	namedTupleType
	arrayType
	enumType
)

// CodecLookup ...
type CodecLookup map[types.UUID]DecodeEncoder

// DecodeEncoder interface
type DecodeEncoder interface {
	Decode(*[]byte) interface{}
	Encode(*[]byte, interface{})
}

// Pop a decoder
func Pop(bts *[]byte) CodecLookup {
	lookup := CodecLookup{}
	codecs := []DecodeEncoder{}

	for len(*bts) > 0 {
		descriptorType := protocol.PopUint8(bts)
		id := protocol.PopUUID(bts)

		switch descriptorType {
		case setType:
			lookup[id] = popSetCodec(bts, id, codecs)
		case objectType:
			lookup[id] = popObjectCodec(bts, id, codecs)
		case baseScalarType:
			lookup[id] = getBaseScalarCodec(id)
		case scalarType:
			panic("scalar type descriptor not implemented") // todo
		case tupleType:
			lookup[id] = popTupleCodec(bts, id, codecs)
		case namedTupleType:
			lookup[id] = popNamedTupleCodec(bts, id, codecs)
		case arrayType:
			lookup[id] = popArrayCodec(bts, id, codecs)
		case enumType:
			panic("enum type descriptor not implemented") // todo
		default:
			panic(fmt.Sprintf("unknown descriptor type 0x%x:\n% x\n", descriptorType, bts))
		}
		codecs = append(codecs, lookup[id])
	}
	return lookup
}

func popSetCodec(bts *[]byte, id types.UUID, codecs []DecodeEncoder) DecodeEncoder {
	n := protocol.PopUint16(bts)
	return &Set{codecs[n]}
}

// Set is an EdgeDB set type codec.
type Set struct {
	child DecodeEncoder
}

// Decode a set
func (c *Set) Decode(bts *[]byte) interface{} {
	buf := protocol.PopBytes(bts)

	dimCount := protocol.PopUint32(&buf) // number of dimensions, either 0 or 1
	if dimCount == 0 {
		return types.Set{}
	}

	protocol.PopUint32(&buf) // reserved
	protocol.PopUint32(&buf) // reserved

	upper := int32(protocol.PopUint32(&buf))
	lower := int32(protocol.PopUint32(&buf))
	elmCount := int(upper - lower + 1)

	out := make(types.Set, elmCount)
	for i := 0; i < elmCount; i++ {
		out[i] = c.child.Decode(&buf)
	}

	return out
}

// Encode a set
func (c *Set) Encode(bts *[]byte, val interface{}) {
	panic("not implemented")
}

func popObjectCodec(bts *[]byte, id types.UUID, codecs []DecodeEncoder) DecodeEncoder {
	fields := []objectField{}

	elmCount := int(protocol.PopUint16(bts))
	for i := 0; i < elmCount; i++ {
		flags := protocol.PopUint8(bts)
		name := protocol.PopString(bts)
		index := protocol.PopUint16(bts)

		field := objectField{
			isImplicit:     flags&0b1 != 0,
			isLinkProperty: flags&0b10 != 0,
			isLink:         flags&0b100 != 0,
			name:           name,
			codec:          codecs[index],
		}

		fields = append(fields, field)
	}

	return &Object{fields}
}

// Object is an EdgeDB object type codec.
type Object struct {
	fields []objectField
}

type objectField struct {
	isImplicit     bool
	isLinkProperty bool
	isLink         bool
	name           string
	codec          DecodeEncoder
}

// Decode an object
func (c *Object) Decode(bts *[]byte) interface{} {
	buf := protocol.PopBytes(bts)

	elmCount := int(int32(protocol.PopUint32(&buf)))
	out := make(types.Object)

	for i := 0; i < elmCount; i++ {
		protocol.PopUint32(&buf) // reserved
		field := c.fields[i]

		switch int32(protocol.PeekUint32(&buf)) {
		case -1:
			// element length -1 means missing field
			// https://www.edgedb.com/docs/internals/protocol/dataformats#tuple-namedtuple-and-object
			protocol.PopUint32(&buf)
			out[field.name] = types.Set{}
		default:
			out[field.name] = field.codec.Decode(&buf)
		}
	}

	return out
}

// Encode an object
func (c *Object) Encode(bts *[]byte, val interface{}) {
	panic("objects can't be query parameters")
}

func getBaseScalarCodec(id types.UUID) DecodeEncoder {
	switch id {
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}:
		return &UUID{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}:
		return &String{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2}:
		return &Bytes{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3}:
		return &Int16{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 4}:
		return &Int32{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 5}:
		return &Int64{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 6}:
		return &Float32{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 7}:
		return &Float64{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 8}:
		panic("decimal type not implemented") // todo implement
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 9}:
		return &Bool{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xa}:
		return &DateTime{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xb}:
		panic("cal::local_datetime type not implemented") // todo implement
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xc}:
		panic("cal::local_date type not implemented") // todo implement
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xd}:
		panic("cal::local_time type not implemented") // todo implement
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xe}:
		return &Duration{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0xf}:
		return &JSON{}
	case types.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0x10}:
		panic("bigint type not implemented") // todo implement
	default:
		panic(fmt.Sprintf("unknown base scalar type descriptor id: % x", id))
	}
}

// UUID is an EdgeDB UUID type codec.
type UUID struct{}

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
type String struct{}

// Decode a string.
func (c *String) Decode(bts *[]byte) interface{} {
	return protocol.PopString(bts)
}

// Encode a string.
func (c *String) Encode(bts *[]byte, val interface{}) {
	protocol.PushString(bts, val.(string))
}

// Bytes is an EdgeDB bytes type codec.
type Bytes struct{}

// Decode []byte.
func (c *Bytes) Decode(bts *[]byte) interface{} {
	return protocol.PopBytes(bts)
}

// Encode []byte.
func (c *Bytes) Encode(bts *[]byte, val interface{}) {
	protocol.PushBytes(bts, val.([]byte))
}

// Int16 is an EdgeDB int64 type codec.
type Int16 struct{}

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
type Int32 struct{}

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
type Int64 struct{}

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
type Float32 struct{}

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
type Float64 struct{}

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
type Bool struct{}

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
type DateTime struct{}

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
type Duration struct{}

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
type JSON struct{}

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

func popTupleCodec(bts *[]byte, id types.UUID, codecs []DecodeEncoder) DecodeEncoder {
	fields := []DecodeEncoder{}

	elmCount := int(protocol.PopUint16(bts))
	for i := 0; i < elmCount; i++ {
		index := protocol.PopUint16(bts)
		fields = append(fields, codecs[index])
	}

	return &Tuple{fields}
}

// Tuple is an EdgeDB tuple type codec.
type Tuple struct {
	fields []DecodeEncoder
}

// Decode a tuple.
func (c *Tuple) Decode(bts *[]byte) interface{} {
	buf := protocol.PopBytes(bts)

	elmCount := int(int32(protocol.PopUint32(&buf)))
	out := make(types.Tuple, elmCount)

	for i := 0; i < elmCount; i++ {
		protocol.PopUint32(&buf) // reserved
		out[i] = c.fields[i].Decode(&buf)
	}

	return out
}

// Encode a tuple.
func (c *Tuple) Encode(bts *[]byte, val interface{}) {
	elmCount := len(c.fields)

	// special case for null tuple
	if elmCount == 0 {
		protocol.PushUint32(bts, 4) // data length
		protocol.PushUint32(bts, uint32(elmCount))
		return
	}

	tmp := []byte{}
	protocol.PushUint32(&tmp, uint32(elmCount))
	in := val.([]interface{})
	for i := 0; i < elmCount; i++ {
		protocol.PushUint32(&tmp, 0) // reserved
		c.fields[i].Encode(&tmp, in[i])
	}

	protocol.PushUint32(bts, uint32(len(tmp)))
	*bts = append(*bts, tmp...)
}

func popNamedTupleCodec(bts *[]byte, id types.UUID, codecs []DecodeEncoder) DecodeEncoder {
	fields := []namedTupleField{}

	elmCount := int(protocol.PopUint16(bts))
	for i := 0; i < elmCount; i++ {
		name := protocol.PopString(bts)
		index := protocol.PopUint16(bts)

		field := namedTupleField{
			name:  name,
			codec: codecs[index],
		}

		fields = append(fields, field)
	}

	return &NamedTuple{fields}
}

type namedTupleField struct {
	name  string
	codec DecodeEncoder
}

// NamedTuple is an EdgeDB namedtuple typep codec.
type NamedTuple struct {
	fields []namedTupleField
}

// Decode a named tuple.
func (c *NamedTuple) Decode(bts *[]byte) interface{} {
	buf := protocol.PopBytes(bts)

	elmCount := int(int32(protocol.PopUint32(&buf)))
	out := make(types.NamedTuple)

	for i := 0; i < elmCount; i++ {
		protocol.PopUint32(&buf) // reserved
		field := c.fields[i]
		out[field.name] = field.codec.Decode(&buf)
	}

	return out
}

// Encode a named tuple.
func (c *NamedTuple) Encode(bts *[]byte, val interface{}) {
	// don't know the data length yet
	// put everything in a new slice to get the length
	tmp := []byte{}

	elmCount := len(c.fields)
	protocol.PushUint32(&tmp, uint32(elmCount))

	args := val.([]interface{})
	if len(args) != 1 {
		panic(fmt.Sprintf("wrong number of arguments: %v", args))
	}

	in := args[0].(map[string]interface{})

	for i := 0; i < elmCount; i++ {
		protocol.PushUint32(&tmp, 0) // reserved
		field := c.fields[i]
		field.codec.Encode(&tmp, in[field.name])
	}

	protocol.PushUint32(bts, uint32(len(tmp)))
	*bts = append(*bts, tmp...)
}

func popArrayCodec(bts *[]byte, id types.UUID, codecs []DecodeEncoder) DecodeEncoder {
	index := protocol.PopUint16(bts) // element type descriptor index

	n := int(protocol.PopUint16(bts)) // number of array dimensions
	for i := 0; i < n; i++ {
		protocol.PopUint32(bts) //array dimension
	}

	return &Array{codecs[index]}
}

// Array is an EdgeDB array type codec.
type Array struct {
	child DecodeEncoder
}

// Decode an array.
func (c *Array) Decode(bts *[]byte) interface{} {
	buf := protocol.PopBytes(bts)

	// number of dimensions is 1 or 0
	dimCount := protocol.PopUint32(&buf)
	if dimCount == 0 {
		return types.Array{}
	}

	protocol.PopUint32(&buf) // reserved
	protocol.PopUint32(&buf) // reserved

	upper := int32(protocol.PopUint32(&buf))
	lower := int32(protocol.PopUint32(&buf))
	elmCount := int(upper - lower + 1)

	out := make(types.Array, elmCount)
	for i := 0; i < elmCount; i++ {
		out[i] = c.child.Decode(&buf)
	}

	return out
}

// Encode an array.
func (c *Array) Encode(bts *[]byte, val interface{}) {
	// the data length is not know until all values have been encoded
	// put the data in temporary slice to get the length
	tmp := []byte{}

	in := val.([]interface{})
	elmCount := len(in)

	protocol.PushUint32(&tmp, 1)                // number of dimensions
	protocol.PushUint32(&tmp, 0)                // reserved
	protocol.PushUint32(&tmp, 0)                // reserved
	protocol.PushUint32(&tmp, uint32(elmCount)) // dimension.upper
	protocol.PushUint32(&tmp, 1)                // dimension.lower

	for i := 0; i < elmCount; i++ {
		c.child.Encode(&tmp, in[i])
	}

	protocol.PushUint32(bts, uint32(len(tmp)))
	*bts = append(*bts, tmp...)
}
