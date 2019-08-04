//Package scanner provides a 'i' syntax scanner.
package scanner

import "bytes"

//Token is a token read from a scanner.
type Token []byte

func (token Token) String() string {
	return string(token)
}

//Is can be used to comare tokens for equality.
func (token Token) Is(constant string) bool {
	return bytes.Equal(token, []byte(constant))
}
