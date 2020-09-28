package codecs

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/fmoor/edgedb-golang/edgedb/protocol"
)

const (
	setType = iota // todo implement
	objectType
	baseScalarType
	scalarType // todo implement
	tupleType
	namedTupleType // todo implement
	arrayType
	enumType // todo implement
)

// CodecLookup ...
type CodecLookup map[protocol.UUID]DecodeEncoder

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
		case objectType:
			lookup[id] = popObjectCodec(bts, id, codecs)
		case baseScalarType:
			lookup[id] = getBaseScalarCodec(id)
		case tupleType:
			lookup[id] = popTupleCodec(bts, id, codecs)
		case namedTupleType:
			lookup[id] = popNamedTupleCodec(bts, id, codecs)
		case arrayType:
			lookup[id] = popArrayCodec(bts, id, codecs)
		default:
			panic(fmt.Sprintf("unknown descriptor type %x:\n% x\n", descriptorType, bts))
		}
		codecs = append(codecs, lookup[id])
	}
	return lookup
}

func popObjectCodec(bts *[]byte, id protocol.UUID, codecs []DecodeEncoder) DecodeEncoder {
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

// Object codec
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
	protocol.PopUint32(bts) // data length
	elmCount := int(int32(protocol.PopUint32(bts)))
	out := map[string]interface{}{}
	for i := 0; i < elmCount; i++ {
		protocol.PopUint32(bts) // reserved
		field := c.fields[i]
		if field.name == "__tid__" {
			field.codec.Decode(bts)
		} else {
			out[field.name] = field.codec.Decode(bts)
		}
	}
	return out
}

// Encode an object
func (c *Object) Encode(bts *[]byte, val interface{}) {
	panic("objects can't be query parameters")
}

func getBaseScalarCodec(id protocol.UUID) DecodeEncoder {
	switch id {
	case "00000000-0000-0000-0000-0000-00000100":
		return &UUID{}
	case "00000000-0000-0000-0000-0000-00000101":
		return &String{}
	case "00000000-0000-0000-0000-0000-00000102":
		return &Bytes{}
	case "00000000-0000-0000-0000-0000-00000103":
		return &Int16{}
	case "00000000-0000-0000-0000-0000-00000104":
		return &Int32{}
	case "00000000-0000-0000-0000-0000-00000105":
		return &Int64{}
	case "00000000-0000-0000-0000-0000-00000106":
		return &Float32{}
	case "00000000-0000-0000-0000-0000-00000107":
		return &Float64{}
	case "00000000-0000-0000-0000-0000-00000108":
		panic("decimal type not implemented") // todo implement
	case "00000000-0000-0000-0000-0000-00000109":
		return &Bool{}
	case "00000000-0000-0000-0000-0000-0000010a":
		return &DateTime{}
	case "00000000-0000-0000-0000-0000-0000010b":
		return &LocalDateTime{}
	case "00000000-0000-0000-0000-0000-0000010c":
		return &LocalDate{}
	case "00000000-0000-0000-0000-0000-0000010d":
		return &LocalTime{}
	case "00000000-0000-0000-0000-0000-0000010e":
		return &Duration{}
	case "00000000-0000-0000-0000-0000-0000010f":
		return &JSON{}
	case "00000000-0000-0000-0000-0000-00000110":
		panic("bigint type not implemented") // todo implement
	default:
		panic(fmt.Sprintf("unknown base scalar type descriptor id: %q", id))
	}
}

// UUID codec
type UUID struct{}

// Decode a UUID
func (c *UUID) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	return protocol.PopUUID(bts)
}

// Encode UUID
func (c *UUID) Encode(bts *[]byte, val interface{}) {
	panic("not implemented, todo")
}

// String codec
type String struct{}

// Decode string
func (c *String) Decode(bts *[]byte) interface{} {
	return protocol.PopString(bts)
}

// Encode string
func (c *String) Encode(bts *[]byte, val interface{}) {
	protocol.PushString(bts, val.(string))
}

// Bytes codec
type Bytes struct{}

// Decode []byte
func (c *Bytes) Decode(bts *[]byte) interface{} {
	return protocol.PopBytes(bts)
}

// Encode []byte
func (c *Bytes) Encode(bts *[]byte, val interface{}) {
	protocol.PushBytes(bts, val.([]byte))
}

// Int16 codec
type Int16 struct{}

// Decode int16
func (c *Int16) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	return int16(protocol.PopUint16(bts))
}

// Encode int16
func (c *Int16) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 2) // data length
	protocol.PushUint16(bts, uint16(val.(int16)))
}

// Int32 codec
type Int32 struct{}

// Decode int32
func (c *Int32) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	return int32(protocol.PopUint32(bts))
}

// Encode int32
func (c *Int32) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 4) // data length
	protocol.PushUint32(bts, uint32(val.(int32)))
}

// Int64 codec
type Int64 struct{}

// Decode int64
func (c *Int64) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	return int64(protocol.PopUint64(bts))
}

// Encode int64
func (c *Int64) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 8) // data length
	protocol.PushUint64(bts, uint64(val.(int64)))
}

// Float32 codec
type Float32 struct{}

// Decode float32
func (c *Float32) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	bits := protocol.PopUint32(bts)
	return math.Float32frombits(bits)
}

// Encode float32
func (c *Float32) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 4)
	protocol.PushUint32(bts, math.Float32bits(val.(float32)))
}

// Float64 codec
type Float64 struct{}

// Decode float64
func (c *Float64) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	bits := protocol.PopUint64(bts)
	return math.Float64frombits(bits)
}

// Encode float64
func (c *Float64) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 8)
	protocol.PushUint64(bts, math.Float64bits(val.(float64)))
}

// Bool codec
type Bool struct{}

// Decode bool
func (c *Bool) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	val := protocol.PopUint8(bts)
	if val > 1 {
		panic(fmt.Sprintf("invalid bool byte, must be 0 or 1, got: 0x%x", val))
	}
	return val != 0
}

// Encode bool
func (c *Bool) Encode(bts *[]byte, val interface{}) {
	protocol.PushUint32(bts, 1) // data length

	// convert bool to uint8
	var out uint8
	if val.(bool) {
		out = 1
	}

	protocol.PushUint8(bts, out)
}

// DateTime codec
type DateTime struct{}

// Decode datetime
func (c *DateTime) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	val := int64(protocol.PopUint64(bts))
	seconds := val / 1_000_000
	microseconds := val % 1_000_000
	return time.Unix(946_684_800+seconds, 1_000*microseconds).UTC()
}

// Encode date time
func (c *DateTime) Encode(bts *[]byte, val interface{}) {
	date := val.(time.Time)
	seconds := date.Unix() - 946_684_800
	nanoseconds := int64(date.Sub(time.Unix(date.Unix(), 0)))
	microseconds := seconds*1_000_000 + nanoseconds/1_000
	protocol.PushUint32(bts, 8) // data length
	protocol.PushUint64(bts, uint64(microseconds))
}

// LocalDateTime codec
type LocalDateTime struct{}

// Decode local datetime
func (c *LocalDateTime) Decode(bts *[]byte) interface{} {
	// todo this should return a date and time without a timezone
	protocol.PopUint32(bts) // data length
	val := int64(protocol.PopUint64(bts))
	return time.Unix(0, 1_000*(val+946_684_800_000_000)).UTC()
}

// Encode local date time
func (c *LocalDateTime) Encode(bts *[]byte, val interface{}) {
	panic("not implemented, todo")
}

// LocalDate codec
type LocalDate struct{}

// Decode local date
func (c *LocalDate) Decode(bts *[]byte) interface{} {
	// todo this should return a date without a time or timezone
	protocol.PopUint32(bts) // data length
	val := int32(protocol.PopUint32(bts))
	delta, err := time.ParseDuration(fmt.Sprintf("%vh", 24*val))
	if err != nil {
		panic(err)
	}
	location, err := time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}
	return time.Date(2000, 1, 1, 0, 0, 0, 0, location).Add(delta)
}

// Encode local date
func (c *LocalDate) Encode(bts *[]byte, val interface{}) {
	panic("not implemented, todo")
}

// LocalTime codec
type LocalTime struct{}

// Decode local time
func (c *LocalTime) Decode(bts *[]byte) interface{} {
	// todo this should probably return a different type
	protocol.PopUint32(bts) // data length
	val := int64(protocol.PopUint64(bts))
	str := fmt.Sprintf("%vus", val)
	duration, err := time.ParseDuration(str)
	if err != nil {
		panic(err)
	}
	return duration
}

// Encode local time
func (c *LocalTime) Encode(bts *[]byte, val interface{}) {
	panic("not implemented, todo")
}

// Duration codec
type Duration struct{}

// Decode duration
func (c *Duration) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	microseconds := int64(protocol.PopUint64(bts))
	protocol.PopUint32(bts) // reserved
	protocol.PopUint32(bts) // reserved
	return time.Duration(microseconds * 1_000)
}

// Encode a duration
func (c *Duration) Encode(bts *[]byte, val interface{}) {
	duration := val.(time.Duration)
	protocol.PushUint32(bts, 16) // data length
	protocol.PushUint64(bts, uint64(duration/1_000))
	protocol.PushUint32(bts, 0) // reserved
	protocol.PushUint32(bts, 0) // reserved
}

// JSON codec
type JSON struct{}

// Decode json
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

// Encode json
func (c *JSON) Encode(bts *[]byte, val interface{}) {
	panic("not implemented, todo")
}

func popTupleCodec(bts *[]byte, id protocol.UUID, codecs []DecodeEncoder) DecodeEncoder {
	fields := []DecodeEncoder{}

	elmCount := int(protocol.PopUint16(bts))
	for i := 0; i < elmCount; i++ {
		index := protocol.PopUint16(bts)
		fields = append(fields, codecs[index])
	}
	return &Tuple{fields}
}

// Tuple codec
type Tuple struct {
	fields []DecodeEncoder
}

// Decode a tuple
func (c *Tuple) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	elmCount := int(int32(protocol.PopUint32(bts)))
	out := []interface{}{}
	for i := 0; i < elmCount; i++ {
		protocol.PopUint32(bts) // reserved
		out = append(out, c.fields[i].Decode(bts))
	}
	return out
}

// Encode a tuple
func (c *Tuple) Encode(bts *[]byte, val interface{}) {
	tmp := []byte{}
	elmCount := len(c.fields)

	// special case for null tuple
	// todo this should not be needed
	if elmCount == 0 {
		protocol.PushUint32(bts, 4) // data length
		protocol.PushUint32(bts, uint32(elmCount))
		return
	}

	protocol.PushUint32(&tmp, uint32(elmCount))
	in := val.([]interface{})
	for i := 0; i < elmCount; i++ {
		protocol.PushUint32(&tmp, 0) // reserved
		c.fields[i].Encode(&tmp, in[i])
	}

	protocol.PushUint32(bts, uint32(len(tmp)))
	*bts = append(*bts, tmp...)
}

func popNamedTupleCodec(bts *[]byte, id protocol.UUID, codecs []DecodeEncoder) DecodeEncoder {
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

// NamedTuple codec
type NamedTuple struct {
	fields []namedTupleField
}

// Decode a named tuple
func (c *NamedTuple) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	elmCount := int(int32(protocol.PopUint32(bts)))
	out := map[string]interface{}{}
	for i := 0; i < elmCount; i++ {
		protocol.PopUint32(bts) // reserved
		field := c.fields[i]
		out[field.name] = field.codec.Decode(bts)
	}
	return out
}

// Encode a named tuple
func (c *NamedTuple) Encode(bts *[]byte, val interface{}) {
	// don't know the data length yet
	// put everything in a new slice to get the length
	tmp := []byte{}

	elmCount := len(c.fields)
	protocol.PushUint32(&tmp, uint32(elmCount))
	in := val.(map[string]interface{})
	for i := 0; i < elmCount; i++ {
		protocol.PushUint32(&tmp, 0) // reserved
		field := c.fields[i]
		field.codec.Encode(&tmp, in[field.name])
	}

	protocol.PushUint32(bts, uint32(len(tmp)))
	*bts = append(*bts, tmp...)
}

func popArrayCodec(bts *[]byte, id protocol.UUID, codecs []DecodeEncoder) DecodeEncoder {
	index := protocol.PopUint16(bts)  // element type descriptor index
	n := int(protocol.PopUint16(bts)) // number of array dimensions
	for i := 0; i < n; i++ {
		protocol.PopUint32(bts) //array dimension
	}
	return &Array{codecs[index]}
}

// Array codec
type Array struct {
	child DecodeEncoder
}

type dimension struct {
	upper int32
	lower int32
}

// Decode an array
func (c *Array) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	dimensions := []dimension{}
	dimCount := int(int32(protocol.PopUint32(bts))) // number of dimensions
	protocol.PopUint32(bts)                         // reserved
	protocol.PopUint32(bts)                         // reserved
	for i := 0; i < dimCount; i++ {
		upper := int32(protocol.PopUint32(bts))
		lower := int32(protocol.PopUint32(bts))
		dimensions = append(dimensions, dimension{upper, lower})
	}
	elmCount := 0
	for _, dim := range dimensions {
		elmCount += int(dim.upper - dim.lower + 1)
	}
	out := []interface{}{}
	for i := 0; i < elmCount; i++ {
		out = append(out, c.child.Decode(bts))
	}
	return out
}

// Encode an array
func (c *Array) Encode(bts *[]byte, val interface{}) {
	// the data length is not know until all values have been encoded
	// put the data in temporary slice to get the length
	tmp := []byte{}

	protocol.PushUint32(&tmp, 1) // number of dimensions
	protocol.PushUint32(&tmp, 0) // reserved
	protocol.PushUint32(&tmp, 0) // reserved
	protocol.PushUint32(&tmp, 3) // dimension.upper
	protocol.PushUint32(&tmp, 1) // dimension.lower

	in := val.([]interface{})
	elmCount := len(in)
	for i := 0; i < elmCount; i++ {
		c.child.Encode(&tmp, in[i])
	}

	protocol.PushUint32(bts, uint32(len(tmp)))
	*bts = append(*bts, tmp...)
}
