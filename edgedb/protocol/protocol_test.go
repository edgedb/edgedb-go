package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPopUint8(t *testing.T) {
	bts := []byte{10, 12}
	val := PopUint8(&bts)
	assert.Equal(t, uint8(10), val)
	assert.Equal(t, []byte{12}, bts)
}

func TestPopUint32(t *testing.T) {
	bts := []byte{0, 0, 0, 37, 10}
	val := PopUint32(&bts)
	assert.Equal(t, uint32(37), val)
	assert.Equal(t, []byte{10}, bts)
}

func TestPopBytes(t *testing.T) {
	bts := []byte{0, 0, 0, 1, 32, 2}
	val, n := PopBytes(&bts)
	assert.Equal(t, []byte{32}, val)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte{2}, bts)
}

func TestPopString(t *testing.T) {
	bts := []byte{0, 0, 0, 3, 102, 111, 111}
	str, n := PopString(&bts)
	assert.Equal(t, "foo", str)
	assert.Equal(t, 7, n)
	assert.Equal(t, []byte{}, bts)
}

func TestPopMessage(t *testing.T) {
	bts := []byte{32, 0, 0, 0, 5, 6, 7}
	msg := PopMessage(&bts)
	expected := []byte{32, 0, 0, 0, 5, 6}
	assert.Equal(t, expected, msg)
	assert.Equal(t, []byte{7}, bts)
}
