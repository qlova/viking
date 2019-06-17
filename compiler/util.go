package compiler

import "bytes"

func equal(b []byte, s string) bool {
	return bytes.Equal(b, []byte(s))
}

func s(s string) []byte {
	return []byte(s)
}
