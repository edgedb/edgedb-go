package codecs

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/fmoor/edgedb-golang/edgedb/protocol"
)

const (
	setType = iota
	objectType
	baseScalarType
	scalarType
	tupleType
	nameTupleType
	arrayType
	enumType
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
	decoders := []Decoder{}

	for len(*bts) > 0 {
		descriptorType := protocol.PopUint8(bts)
		id := protocol.PopUUID(bts)

		switch descriptorType {
		case objectType:
			lookup[id] = popObjectCodec(bts, id, decoders)
		case baseScalarType:
			lookup[id] = &BaseScalarCodec{id}
		case tupleType:
			lookup[id] = popTupleCodec(bts, id, decoders)
		case arrayType:
			lookup[id] = popArrayCodec(bts, id, decoders)
		default:
			panic(fmt.Sprintf("unknown descriptor type %x:\n% x\n", descriptorType, bts))
		}
		decoders = append(decoders, lookup[id])
	}
	return lookup
}

func popObjectCodec(bts *[]byte, id protocol.UUID, codecs []Decoder) Decoder {
	fields := []objectField{}

	elmCount := int(protocol.PopUint16(bts))
	for i := 0; i < elmCount; i++ {
		flags := protocol.PopUint8(bts)
		name, _ := protocol.PopString(bts)
		index := protocol.PopUint16(bts)

		field := objectField{
			isImplicit:     0 != flags&0b1,
			isLinkProperty: 0 != flags&0b10,
			isLink:         0 != flags&0b100,
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

// Decode an object from bytes
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

type BaseScalarCodec struct {
	id protocol.UUID
}

func (b *BaseScalarCodec) Decode(bts *[]byte) interface{} {
	switch b.id {
	case "00000000-0000-0000-0000-0000-00000100":
		protocol.PopUint32(bts) // data length
		return protocol.PopUUID(bts)
	case "00000000-0000-0000-0000-0000-00000101":
		val, _ := protocol.PopString(bts)
		return val
	case "00000000-0000-0000-0000-0000-00000102":
		val, _ := protocol.PopBytes(bts)
		return val
	case "00000000-0000-0000-0000-0000-00000103":
		protocol.PopUint32(bts) // data length
		return int16(protocol.PopUint16(bts))
	case "00000000-0000-0000-0000-0000-00000104":
		protocol.PopUint32(bts) // data length
		return protocol.PopInt32(bts)
	case "00000000-0000-0000-0000-0000-00000105":
		protocol.PopUint32(bts) // data length
		return protocol.PopInt64(bts)
	case "00000000-0000-0000-0000-0000-00000106":
		protocol.PopUint32(bts) // data length
		bits := protocol.PopUint32(bts)
		return math.Float32frombits(bits)
	case "00000000-0000-0000-0000-0000-00000107":
		protocol.PopUint32(bts) // data length
		bits := protocol.PopUint64(bts)
		return math.Float64frombits(bits)
	case "00000000-0000-0000-0000-0000-00000108":
		panic("decimal type not implemented")
	case "00000000-0000-0000-0000-0000-00000109":
		protocol.PopUint32(bts) // data length
		val := protocol.PopUint8(bts)
		if val > 1 {
			panic(fmt.Sprintf("invalid bool byte, must be 0 or 1, got: %x", val))
		}
		return val != 0
	case "00000000-0000-0000-0000-0000-0000010a":
		protocol.PopUint32(bts) // data length
		val := protocol.PopInt64(bts)
		return time.Unix(0, 1_000*(val+946_684_800_000_000)).UTC()
	case "00000000-0000-0000-0000-0000-0000010b":
		// todo this should return a date and time without a timezone
		protocol.PopUint32(bts) // data length
		val := protocol.PopInt64(bts)
		return time.Unix(0, 1_000*(val+946_684_800_000_000)).UTC()
	case "00000000-0000-0000-0000-0000-0000010c":
		// todo this should return a date without a time or timezone
		protocol.PopUint32(bts) // data length
		val := protocol.PopInt32(bts)
		delta, _ := time.ParseDuration(fmt.Sprintf("%vh", 24*val))
		location, _ := time.LoadLocation("UTC")
		return time.Date(2000, 1, 1, 0, 0, 0, 0, location).Add(delta)
	case "00000000-0000-0000-0000-0000-0000010d":
		// todo this should probably return a different type
		protocol.PopUint32(bts) // data length
		val := protocol.PopInt64(bts)
		str := fmt.Sprintf("%vus", val)
		duration, _ := time.ParseDuration(str)
		return duration
	case "00000000-0000-0000-0000-0000-0000010e":
		protocol.PopUint32(bts) // data length
		microSeconds := protocol.PopInt64(bts)
		protocol.PopUint32(bts) // reserved
		protocol.PopUint32(bts) // reserved
		duration, _ := time.ParseDuration(fmt.Sprintf("%vus", microSeconds))
		return duration
	case "00000000-0000-0000-0000-0000-0000010f":
		n := protocol.PopUint32(bts) // data length
		protocol.PopUint8(bts)       // json format, always 1
		var val interface{}
		json.Unmarshal((*bts)[:n-1], &val)
		*bts = (*bts)[n-1:]
		return val
	case "00000000-0000-0000-0000-0000-00000110":
		panic("bigint type not implemented")
	default:
		panic(fmt.Sprintf("unknown base scalar type descriptor: %v", b.id))
	}
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

// Decode a tuple from bytes
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

// Decode and array from bytes
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
