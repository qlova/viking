package scanner

import "bytes"

type Token []byte

func (token Token) String() string {
	return string(token)
}

func (token Token) Is(constant string) bool {
	return bytes.Equal(token, []byte(constant))
}
