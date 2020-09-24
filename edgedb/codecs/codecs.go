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
	nameTupleType // todo implement
	arrayType
	enumType // todo implement
)

// CodecLookup ...
type CodecLookup map[protocol.UUID]Decoder

// Decoder interface
type Decoder interface {
	Decode(*[]byte) interface{}
}

// Get a decoder
func Get(bts *[]byte) CodecLookup {
	lookup := CodecLookup{}
	codecs := []Decoder{}

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
		case arrayType:
			lookup[id] = popArrayCodec(bts, id, codecs)
		default:
			panic(fmt.Sprintf("unknown descriptor type %x:\n% x\n", descriptorType, bts))
		}
		codecs = append(codecs, lookup[id])
	}
	return lookup
}

func popObjectCodec(bts *[]byte, id protocol.UUID, codecs []Decoder) Decoder {
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
	codec          Decoder
}

// Decode an object
func (c *Object) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	elmCount := int(protocol.PopInt32(bts))
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

func getBaseScalarCodec(id protocol.UUID) Decoder {
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

// String codec
type String struct{}

// Decode string
func (c *String) Decode(bts *[]byte) interface{} {
	return protocol.PopString(bts)
}

// Bytes codec
type Bytes struct{}

// Decode []byte
func (c *Bytes) Decode(bts *[]byte) interface{} {
	return protocol.PopBytes(bts)
}

// Int16 codec
type Int16 struct{}

// Decode int16
func (c *Int16) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	return int16(protocol.PopUint16(bts))
}

// Int32 codec
type Int32 struct{}

// Decode int32
func (c *Int32) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	return protocol.PopInt32(bts)
}

// Int64 codec
type Int64 struct{}

// Decode int64
func (c *Int64) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	return protocol.PopInt64(bts)
}

// Float32 codec
type Float32 struct{}

// Decode float32
func (c *Float32) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	bits := protocol.PopUint32(bts)
	return math.Float32frombits(bits)
}

// Float64 codec
type Float64 struct{}

// Decode float64
func (c *Float64) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	bits := protocol.PopUint64(bts)
	return math.Float64frombits(bits)
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

// DateTime codec
type DateTime struct{}

// Decode datetime
func (c *DateTime) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	val := protocol.PopInt64(bts)
	return time.Unix(0, 1_000*(val+946_684_800_000_000)).UTC()
}

// LocalDateTime codec
type LocalDateTime struct{}

// Decode local datetime
func (c *LocalDateTime) Decode(bts *[]byte) interface{} {
	// todo this should return a date and time without a timezone
	protocol.PopUint32(bts) // data length
	val := protocol.PopInt64(bts)
	return time.Unix(0, 1_000*(val+946_684_800_000_000)).UTC()
}

// LocalDate codec
type LocalDate struct{}

// Decode local date
func (c *LocalDate) Decode(bts *[]byte) interface{} {
	// todo this should return a date without a time or timezone
	protocol.PopUint32(bts) // data length
	val := protocol.PopInt32(bts)
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

// LocalTime codec
type LocalTime struct{}

// Decode local time
func (c *LocalTime) Decode(bts *[]byte) interface{} {
	// todo this should probably return a different type
	protocol.PopUint32(bts) // data length
	val := protocol.PopInt64(bts)
	str := fmt.Sprintf("%vus", val)
	duration, err := time.ParseDuration(str)
	if err != nil {
		panic(err)
	}
	return duration
}

// Duration codec
type Duration struct{}

// Decode duration
func (c *Duration) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	microSeconds := protocol.PopInt64(bts)
	protocol.PopUint32(bts) // reserved
	protocol.PopUint32(bts) // reserved
	duration, err := time.ParseDuration(fmt.Sprintf("%vus", microSeconds))
	if err != nil {
		panic(err)
	}
	return duration
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

func popTupleCodec(bts *[]byte, id protocol.UUID, codecs []Decoder) Decoder {
	fields := []Decoder{}

	elmCount := int(protocol.PopUint16(bts))
	for i := 0; i < elmCount; i++ {
		index := protocol.PopUint16(bts)
		fields = append(fields, codecs[index])
	}
	return &Tuple{fields}
}

// Tuple codec
type Tuple struct {
	fields []Decoder
}

// Decode a tuple
func (t *Tuple) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	elmCount := int(protocol.PopInt32(bts))
	out := []interface{}{}
	for i := 0; i < elmCount; i++ {
		protocol.PopUint32(bts) // reserved
		out = append(out, t.fields[i].Decode(bts))
	}
	return out
}

func popArrayCodec(bts *[]byte, id protocol.UUID, codecs []Decoder) Decoder {
	index := protocol.PopUint16(bts)  // element type descriptor index
	n := int(protocol.PopUint16(bts)) // number of array dimensions
	for i := 0; i < n; i++ {
		protocol.PopUint32(bts) //array dimension
	}
	return &Array{codecs[index]}
}

// Array codec
type Array struct {
	child Decoder
}

// Decode and array
func (a *Array) Decode(bts *[]byte) interface{} {
	protocol.PopUint32(bts) // data length
	dimensions := []dimension{}
	dimCount := int(protocol.PopInt32(bts)) // number of dimensions
	protocol.PopUint32(bts)                 // reserved
	protocol.PopUint32(bts)                 // reserved
	for i := 0; i < dimCount; i++ {
		upper := protocol.PopInt32(bts)
		lower := protocol.PopInt32(bts)
		dimensions = append(dimensions, dimension{upper, lower})
	}
	elmCount := 0
	for _, dim := range dimensions {
		elmCount += int(dim.upper - dim.lower + 1)
	}
	out := []interface{}{}
	for i := 0; i < elmCount; i++ {
		out = append(out, a.child.Decode(bts))
	}
	return out
}

type dimension struct {
	upper int32
	lower int32
}
