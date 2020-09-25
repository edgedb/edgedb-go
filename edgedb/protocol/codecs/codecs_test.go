package codecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeString(t *testing.T) {
	bts := []byte{}
	codec := &String{}
	codec.Encode(&bts, "hello")
	expected := []byte{0, 0, 0, 5, 104, 101, 108, 108, 111}
	assert.Equal(t, expected, bts)
}

func TestEncodeInt64(t *testing.T) {
	bts := []byte{}
	codec := &Int64{}
	codec.Encode(&bts, int64(27))
	expected := []byte{0, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0, 27}
	assert.Equal(t, expected, bts)
}

func TestEncodeBool(t *testing.T) {
	bts := []byte{}
	codec := &Bool{}
	codec.Encode(&bts, true)
	expected := []byte{0, 0, 0, 1, 1}
	assert.Equal(t, expected, bts)
}

func TestEncodeNullTuple(t *testing.T) {
	bts := []byte{}
	codec := &Tuple{}
	codec.Encode(&bts, []interface{}{})
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

func TestEncodeNamedTuple(t *testing.T) {
	bts := []byte{}
	codec := &NamedTuple{[]namedTupleField{
		namedTupleField{"a", &Int32{}},
		namedTupleField{"b", &Int32{}},
	}}
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
