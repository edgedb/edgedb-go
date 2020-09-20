package payload

import (
	"bytes"
	"testing"
)

func TestMake(t *testing.T) {
	pyld := Make([]byte{1, 2, 3, 4})
	bts := pyld.ToBytes()
	expected := []byte{1, 2, 3, 4}
	if !bytes.Equal(bts, expected) {
		t.Errorf("%v != %v", bts, expected)
	}
}

func TestPush(t *testing.T) {
	pyld := Make([]byte{1, 2, 3})
	pyld.Push([]byte{4, 5, 6})
	bts := pyld.ToBytes()
	expected := []byte{1, 2, 3, 4, 5, 6}
	if !bytes.Equal(bts, expected) {
		t.Errorf("%v != %v", bts, expected)
	}
}

func TestPopEmpty(t *testing.T) {
	pyld := Make([]byte{})
	bts, err := pyld.Pop()
	if bts != nil {
		t.Errorf("bts should be nil got: %v", bts)
	}
	if err != nil {
		t.Errorf("err should be nil got: %v", err)
	}
}

func TestPop(t *testing.T) {
	pyld := Make([]byte{1, 0, 0, 0, 6, 7, 8})
	bts, err := pyld.Pop()
	if err != nil {
		t.Errorf("err should be nil got: %v", err)
	}

	expected := []byte{1, 0, 0, 0, 6, 7, 8}
	if !bytes.Equal(bts, expected) {
		t.Errorf("%v != %v", bts, expected)
	}
}

func TestPopEmptySecond(t *testing.T) {
	pyld := Make([]byte{1, 0, 0, 0, 6, 7, 8})
	pyld.Pop()
	bts, err := pyld.Pop()
	if bts != nil {
		t.Errorf("bts should be nil got: %v", bts)
	}
	if err != nil {
		t.Errorf("err should be nil got: %v", err)
	}
}

func TestPopStrayBytes(t *testing.T) {
	cases := [][]byte{
		{1},
		{1, 2},
		{1, 2, 3},
		{1, 2, 3, 4},
	}

	for _, inPut := range cases {
		pyld := Make(inPut)
		bts, err := pyld.Pop()
		if bts != nil {
			t.Errorf("bts should be nil for %v got: %v", inPut, bts)
		}
		if err == nil {
			t.Errorf("err should not be nil for %v", inPut)
		}
	}
}
