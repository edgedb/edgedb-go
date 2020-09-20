package protocol

import (
	"bytes"
	"testing"
)

func TestPopUint8(t *testing.T) {
	bts := []byte{10, 12}
	val := PopUint8(&bts)
	if val != 10 {
		t.Errorf("%v != 10", val)
	}
	if !bytes.Equal(bts, []byte{12}) {
		t.Errorf("%v != [12]", bts)
	}
}

func TestPopUint32(t *testing.T) {
	bts := []byte{0, 0, 0, 37, 10}
	val := PopUint32(&bts)
	if val != 37 {
		t.Errorf("expected 37 got: %v", val)
	}
	if !bytes.Equal(bts, []byte{10}) {
		t.Errorf("expected []byte{10} got: %v", bts)
	}
}

func TestPopBytes(t *testing.T) {
	bts := []byte{0, 0, 0, 1, 32, 2}
	val, n := PopBytes(&bts)
	if !bytes.Equal(val, []byte{32}) {
		t.Errorf("[]byte{32} != %v", val)
	}
	if n != 5 {
		t.Errorf("%v != 5", n)
	}
	if !bytes.Equal(bts, []byte{2}) {
		t.Errorf("[]byte{2} != %v", bts)
	}
}

func TestPopMessage(t *testing.T) {
	bts := []byte{32, 0, 0, 0, 5, 6, 7}
	msg := PopMessage(&bts)
	expected := []byte{32, 0, 0, 0, 5, 6}
	if !bytes.Equal(msg, expected) {
		t.Errorf("%v != %v", msg, expected)
	}
	if !bytes.Equal(bts, []byte{7}) {
		t.Errorf("%v != [7]", bts)
	}
}
