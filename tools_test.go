package main

import (
	"errors"
	"testing"
)

func TestCheck(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Errorf("check should not panic")
		}
	}()
	check(nil)
}

func TestCheckWithErr(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("check should panic")
		}
	}()
	check(errors.New("error"))
}

func TestByteSliceWrong(t *testing.T) {

	if byteSliceEquals([]byte{7, 2, 3, 4, 5, 6}, []byte{1, 2, 3, 4, 5, 6}) {
		t.Error("Expected {1,2,3,4,5,6}")
	}
}

func TestByteSliceNil(t *testing.T) {

	if byteSliceEquals([]byte{7, 2, 3, 4, 5, 6}, nil) {
		t.Error("Expected false")
	}
}

func TestByteSliceDifferentLength(t *testing.T) {

	if byteSliceEquals([]byte{7, 2, 3, 4, 5, 6}, []byte{1, 2, 3}) {
		t.Error("Expected false")
	}
}

func TestByteSliceBothNil(t *testing.T) {

	if !byteSliceEquals(nil, nil) {
		t.Error("Expected true")
	}
}

func TestToHexString(t *testing.T) {
	ret := toHexString([]byte{0, 1, 77, 144, 255})

	if ret != "00 01 4d 90 ff " {
		t.Error("wrong string, got", ret)
	}
}
