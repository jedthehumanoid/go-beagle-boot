package main

import "fmt"

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func toHexString(data []byte) string {

	ret := ""

	for _, char := range data {
		s := fmt.Sprintf("%02x ", char)
		ret = ret + s
	}
	return ret
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
