package main

import (
	"encoding/binary"
	"io"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func byteSliceEquals(a, b []byte) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// difference returns the elements in a that aren't in b
func difference(a, b []string) []string {
	mb := map[string]bool{}
	for _, x := range b {
		mb[x] = true
	}
	ab := []string{}
	for _, x := range a {
		if _, ok := mb[x]; !ok {
			ab = append(ab, x)
		}
	}
	return ab
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func writeMulti(w io.Writer, order binary.ByteOrder, data []interface{}) error {
	var err error
	for _, d := range data {
		err = binary.Write(w, order, d)
		if err != nil {
			return err
		}
	}
	return nil
}
