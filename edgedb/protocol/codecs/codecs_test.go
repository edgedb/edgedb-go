package codecs

import (
	"testing"
	"time"

	"github.com/fmoor/edgedb-golang/edgedb/types"
	"github.com/stretchr/testify/assert"
)

func TestDecodeObject(t *testing.T) {
	codec := &Object{[]objectField{
		objectField{false, false, false, "a", &String{}},
		objectField{false, false, false, "b", &Int32{}},
	}}

	bts := []byte{
		0, 0, 0, 32, // data length
		0, 0, 0, 2, // element count
		// field 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4, // data length
		102, 111, 117, 114, // utf-8 data
		// field 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4, // data length
		0, 0, 0, 4, // int32
	}

	result := codec.Decode(&bts)
	expected := types.Object{
		"a": "four",
		"b": int32(4),
	}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestDecodeUUID(t *testing.T) {
	bts := []byte{
		0, 0, 0, 16, // data length
		0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8,
	}

	id := (&UUID{}).Decode(&bts)

	expected := types.UUID{0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8}
	assert.Equal(t, expected, id)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeUUID(t *testing.T) {
	bts := []byte{}
	(&UUID{}).Encode(&bts, types.UUID{0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8})

	expected := []byte{
		0, 0, 0, 16, // data length
		0, 1, 2, 3, 3, 2, 1, 0, 8, 7, 6, 5, 5, 6, 7, 8,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeString(t *testing.T) {
	bts := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	result := (&String{}).Decode(&bts)

	assert.Equal(t, "hello", result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeString(t *testing.T) {
	bts := []byte{}
	(&String{}).Encode(&bts, "hello")

	expected := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeBytes(t *testing.T) {
	bts := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	result := (&Bytes{}).Decode(&bts)
	expected := []byte{104, 101, 108, 108, 111}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeBytes(t *testing.T) {
	bts := []byte{}
	(&Bytes{}).Encode(&bts, []byte{104, 101, 108, 108, 111})

	expected := []byte{
		0, 0, 0, 5, // data length
		104, 101, 108, 108, 111,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeInt16(t *testing.T) {
	bts := []byte{
		0, 0, 0, 2, // data length
		0, 7, // int16
	}

	result := (&Int16{}).Decode(&bts)

	assert.Equal(t, int16(7), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeInt16(t *testing.T) {
	bts := []byte{}
	(&Int16{}).Encode(&bts, int16(7))

	expected := []byte{
		0, 0, 0, 2, // data length
		0, 7, // int16
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeInt32(t *testing.T) {
	bts := []byte{
		0, 0, 0, 4, // data length
		0, 0, 0, 7, // int32
	}

	result := (&Int32{}).Decode(&bts)

	assert.Equal(t, int32(7), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeInt32(t *testing.T) {
	bts := []byte{}
	(&Int32{}).Encode(&bts, int32(7))

	expected := []byte{
		0, 0, 0, 4, // data length
		0, 0, 0, 7, // int32
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeInt64(t *testing.T) {
	bts := []byte{
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 7, // int64
	}

	result := (&Int64{}).Decode(&bts)

	assert.Equal(t, int64(7), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeInt64(t *testing.T) {
	bts := []byte{}
	(&Int64{}).Encode(&bts, int64(27))

	expected := []byte{
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 27, // int64
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeFloat32(t *testing.T) {
	bts := []byte{
		0, 0, 0, 4, // data length
		0xc2, 0, 0, 0,
	}

	result := (&Float32{}).Decode(&bts)

	assert.Equal(t, float32(-32), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeFloat32(t *testing.T) {
	bts := []byte{}
	(&Float32{}).Encode(&bts, float32(-32))

	expected := []byte{
		0, 0, 0, 4, // data length
		0xc2, 0, 0, 0,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeFloat64(t *testing.T) {
	bts := []byte{
		0, 0, 0, 8, // data length
		0xc0, 0x50, 0, 0, 0, 0, 0, 0,
	}

	result := (&Float64{}).Decode(&bts)

	assert.Equal(t, float64(-64), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeFloat64(t *testing.T) {
	bts := []byte{}
	(&Float64{}).Encode(&bts, float64(-64))

	expected := []byte{
		0, 0, 0, 8, // data length
		0xc0, 0x50, 0, 0, 0, 0, 0, 0,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeBool(t *testing.T) {
	bts := []byte{
		0, 0, 0, 1, // data length
		1,
	}

	result := (&Bool{}).Decode(&bts)

	assert.Equal(t, true, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeBool(t *testing.T) {
	bts := []byte{}
	(&Bool{}).Encode(&bts, true)

	expected := []byte{
		0, 0, 0, 1, // data length
		1,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeDateTime(t *testing.T) {
	bts := []byte{
		0, 0, 0, 8, // data length
		0xff, 0xfc, 0xa2, 0xfe, 0xc4, 0xc8, 0x20, 0x0,
	}

	result := (&DateTime{}).Decode(&bts)
	expected := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeDateTime(t *testing.T) {
	bts := []byte{}
	(&DateTime{}).Encode(&bts, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))

	expected := []byte{
		0, 0, 0, 8, // data length
		0xff, 0xfc, 0xa2, 0xfe, 0xc4, 0xc8, 0x20, 0x0,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeDuration(t *testing.T) {
	bts := []byte{
		0, 0, 0, 0x10, // data length
		0, 0, 0, 0, 0, 0xf, 0x42, 0x40,
		0, 0, 0, 0, // reserved
		0, 0, 0, 0, // reserved
	}

	result := (&Duration{}).Decode(&bts)

	assert.Equal(t, time.Duration(1_000_000_000), result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeDuration(t *testing.T) {
	bts := []byte{}
	(&Duration{}).Encode(&bts, time.Duration(1_000_000_000))

	expected := []byte{
		0, 0, 0, 0x10, // data length
		0, 0, 0, 0, 0, 0xf, 0x42, 0x40,
		0, 0, 0, 0, // reserved
		0, 0, 0, 0, // reserved
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeJSON(t *testing.T) {
	bts := []byte{
		0, 0, 0, 0x12, // data length
		1, // json format
		0x7b, 0x22, 0x68, 0x65,
		0x6c, 0x6c, 0x6f, 0x22,
		0x3a, 0x22, 0x77, 0x6f,
		0x72, 0x6c, 0x64, 0x22,
		0x7d,
	}

	result := (&JSON{}).Decode(&bts)
	expected := map[string]interface{}{"hello": "world"}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeJSON(t *testing.T) {
	bts := []byte{}
	(&JSON{}).Encode(&bts, map[string]string{"hello": "world"})

	expected := []byte{
		0, 0, 0, 0x12, // data length
		1, // json format
		0x7b, 0x22, 0x68, 0x65,
		0x6c, 0x6c, 0x6f, 0x22,
		0x3a, 0x22, 0x77, 0x6f,
		0x72, 0x6c, 0x64, 0x22,
		0x7d,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeTuple(t *testing.T) {
	bts := []byte{
		0, 0, 0, 36, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 2,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 3,
	}

	codec := &Tuple{[]DecodeEncoder{&Int64{}, &Int64{}}}
	result := codec.Decode(&bts)
	expected := types.Tuple{int64(2), int64(3)}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeNullTuple(t *testing.T) {
	bts := []byte{}
	(&Tuple{}).Encode(&bts, []interface{}{})

	expected := []byte{
		0, 0, 0, 4, // data length
		0, 0, 0, 0, // number of elements
	}

	assert.Equal(t, expected, bts)
}

func TestEncodeTuple(t *testing.T) {
	bts := []byte{}

	codec := &Tuple{[]DecodeEncoder{&Int64{}, &Int64{}}}
	codec.Encode(&bts, []interface{}{int64(2), int64(3)})

	expected := []byte{
		0, 0, 0, 36, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 2,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 3,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeNamedTuple(t *testing.T) {
	bts := []byte{
		0, 0, 0, 28, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 5,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 6,
	}

	codec := &NamedTuple{[]namedTupleField{
		namedTupleField{"a", &Int32{}},
		namedTupleField{"b", &Int32{}},
	}}

	result := codec.Decode(&bts)
	expected := types.NamedTuple{
		"a": int32(5),
		"b": int32(6),
	}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeNamedTuple(t *testing.T) {
	codec := &NamedTuple{[]namedTupleField{
		namedTupleField{"a", &Int32{}},
		namedTupleField{"b", &Int32{}},
	}}

	bts := []byte{}
	codec.Encode(&bts, map[string]interface{}{
		"a": int32(5),
		"b": int32(6),
	})

	expected := []byte{
		0, 0, 0, 28, // data length
		0, 0, 0, 2, // number of elements
		// element 0
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 5,
		// element 1
		0, 0, 0, 0, // reserved
		0, 0, 0, 4,
		0, 0, 0, 6,
	}

	assert.Equal(t, expected, bts)
}

func TestDecodeArray(t *testing.T) {
	bts := []byte{
		0, 0, 0, 38, // data length
		0, 0, 0, 1, // dimension count
		0, 0, 0, 0, // reserved
		0, 0, 0, 0x14, // reserved
		0, 0, 0, 3, // dimension.upper
		0, 0, 0, 1, // dimension.lower
		// element 0
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 3, // ing64
		// element 1
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 5, // int64
		// element 2
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 8, // int64
	}

	result := (&Array{&Int64{}}).Decode(&bts)
	expected := types.Array{int64(3), int64(5), int64(8)}

	assert.Equal(t, expected, result)
	assert.Equal(t, []byte{}, bts)
}

func TestEncodeArray(t *testing.T) {
	bts := []byte{}
	(&Array{&Int64{}}).Encode(&bts, []interface{}{int64(3), int64(5), int64(8)})

	expected := []byte{
		0, 0, 0, 0x38, // data length
		0, 0, 0, 1, // dimension count
		0, 0, 0, 0, // reserved
		0, 0, 0, 0, // reserved
		0, 0, 0, 3, // dimension.upper
		0, 0, 0, 1, // dimension.lower
		// element 0
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 3, // ing64
		// element 1
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 5, // int64
		// element 2
		0, 0, 0, 8, // data length
		0, 0, 0, 0, 0, 0, 0, 8, // int64
	}

	assert.Equal(t, expected, bts)
}
