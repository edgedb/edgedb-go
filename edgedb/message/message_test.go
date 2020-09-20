package message

import (
	"bytes"
	"testing"
)

func TestPushUint16(t *testing.T) {
	msg := Make(1)
	msg.PushUint16(9)
	bts := msg.ToBytes()
	expected := []byte{1, 0, 0, 0, 6, 0, 9}
	if !bytes.Equal(bts, expected) {
		t.Errorf("%v != %v", bts, expected)
	}
}

func TestPushUint32(t *testing.T) {
	msg := Make(1)
	msg.PushUint32(13)
	bts := msg.ToBytes()
	expected := []byte{1, 0, 0, 0, 8, 0, 0, 0, 13}
	if !bytes.Equal(bts, expected) {
		t.Errorf("%v != %v", bts, expected)
	}
}

func TestPushString(t *testing.T) {
	msg := Make(1)
	msg.PushString("user")
	bts := msg.ToBytes()
	expected := []byte{1, 0, 0, 0, 12, 0, 0, 0, 4, 117, 115, 101, 114}
	if !bytes.Equal(bts, expected) {
		t.Errorf("%v != %v", bts, expected)
	}
}

func TestPushZeroParams(t *testing.T) {
	msg := Make(1)
	msg.PushParams(Params{})
	bts := msg.ToBytes()
	expected := []byte{1, 0, 0, 0, 6, 0, 0}
	if !bytes.Equal(bts, expected) {
		t.Errorf("%v != %v", bts, expected)
	}
}

func TestPushParams(t *testing.T) {
	msg := Make(1)
	msg.PushParams(Params{"user": "edgedb"})
	bts := msg.ToBytes()
	expected := []byte{
		1,
		0, 0, 0, 24,
		0, 1,
		0, 0, 0, 4, 117, 115, 101, 114,
		0, 0, 0, 6, 101, 100, 103, 101, 100, 98,
	}
	if !bytes.Equal(bts, expected) {
		t.Errorf("%v != %v", bts, expected)
	}
}
