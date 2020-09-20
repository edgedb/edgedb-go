package codecs

import (
	"fmt"
	"math"

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

// DecoderLookup ...
type DecoderLookup map[protocol.UUID]Decoder

// Decoder interface
type Decoder interface {
	Decode(*[]byte) string
}

// Get a decoder
func Get(bts *[]byte) DecoderLookup {
	lookup := DecoderLookup{}
	decoders := []Decoder{}

	for len(*bts) > 0 {
		descriptorType := protocol.PopUint8(bts)
		id := protocol.PopUUID(bts)

		switch descriptorType {
		case objectType:
			lookup[id] = popObjectCodec(bts, id, decoders)
		case baseScalarType:
			lookup[id] = getBaseScalarCodec(id)
		case tupleType:
			lookup[id] = popTupleCodec(bts, id, decoders)
		case arrayType:
			lookup[id] = popArrayCodec(bts, id, decoders)
		default:
			panic(fmt.Sprintf("unknown descriptor type %x:\n% x\n", descriptorType, bts))
		}
		decoders = append(decoders, lookup[id])
	}
	fmt.Println(lookup)
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
func (c *Object) Decode(bts *[]byte) string {
	protocol.PopUint32(bts) // data length
	elmCount := int(protocol.PopInt32(bts))
	out := "{"
	for i := 0; i < elmCount; i++ {
		if i > 0 {
			out += ", "
		}
		protocol.PopUint32(bts) // reserved
		field := c.fields[i]
		out += field.name + ": "
		out += field.codec.Decode(bts)
	}
	return out + "}"
}

func getBaseScalarCodec(uuid protocol.UUID) Decoder {
	switch uuid {
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
		panic("decimal type not implemented")
	case "00000000-0000-0000-0000-0000-00000109":
		return &Bool{}
	case "00000000-0000-0000-0000-0000-00000110":
		panic("bigint type not implemented")
	default:
		panic(fmt.Sprintf("unknown base scalar type descriptor: %v", uuid))
	}
}

// UUID decoder
type UUID struct{}

// Decode a uuid from bytes
func (s *UUID) Decode(bts *[]byte) string {
	id := protocol.PopUUID(bts)
	return fmt.Sprintf("%q", id)
}

// String decoder
type String struct{}

// Decode a utf-8 string from bytes
func (s *String) Decode(bts *[]byte) string {
	str, _ := protocol.PopString(bts)
	return fmt.Sprintf("%q", str)
}

// Bytes decoder
type Bytes struct{}

// Decode bytes from bytes
func (s *Bytes) Decode(bts *[]byte) string {
	str, _ := protocol.PopBytes(bts)
	return fmt.Sprintf("b%q", str)
}

// Int16 decoder
type Int16 struct{}

// Decode int16
func (c *Int16) Decode(bts *[]byte) string {
	protocol.PopUint32(bts) // data length
	return fmt.Sprintf("%v", int16(protocol.PopUint16(bts)))
}

// Int32 decoder
type Int32 struct{}

// Decode int32
func (c *Int32) Decode(bts *[]byte) string {
	protocol.PopUint32(bts) // data length
	return fmt.Sprintf("%v", protocol.PopInt32(bts))
}

// Int64 decoder
type Int64 struct{}

// Decode int64 from bytes
func (c *Int64) Decode(bts *[]byte) string {
	protocol.PopUint32(bts) // data length
	return fmt.Sprintf("%v", protocol.PopInt64(bts))
}

// Float32 decoder
type Float32 struct{}

// Decode float32 from bytes
func (c *Float32) Decode(bts *[]byte) string {
	protocol.PopUint32(bts) // data length
	bits := protocol.PopUint32(bts)
	return fmt.Sprintf("%v", math.Float32frombits(bits))
}

// Float64 decoder
type Float64 struct{}

// Decode float64 from bytes
func (c *Float64) Decode(bts *[]byte) string {
	protocol.PopUint32(bts) // data length
	bits := protocol.PopUint64(bts)
	return fmt.Sprintf("%v", math.Float64frombits(bits))
}

// Bool decoder
type Bool struct{}

// Decode bool from bytes
func (c *Bool) Decode(bts *[]byte) string {
	protocol.PopUint32(bts) // data length
	val := protocol.PopUint8(bts)
	if val > 1 {
		panic(fmt.Sprintf("invalid bool byte must be 00 or 01, got: %x", val))
	}
	return fmt.Sprintf("%v", val != 0)
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
func (t *Tuple) Decode(bts *[]byte) string {
	protocol.PopUint32(bts) // data length
	elmCount := int(protocol.PopInt32(bts))
	out := "("
	for i := 0; i < elmCount; i++ {
		if i > 0 {
			out += ", "
		}
		protocol.PopUint32(bts) // reserved
		out += t.fields[i].Decode(bts)
	}
	return out + ")"
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
func (a *Array) Decode(bts *[]byte) string {
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
	out := "["
	for i := 0; i < elmCount; i++ {
		if i > 0 {
			out += ", "
		}
		out += a.child.Decode(bts)
	}
	return out + "]"
}

type dimension struct {
	upper int32
	lower int32
}
