package scanner

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

type Scanner struct {
	Reader    *bufio.Reader
	NextToken Token
	LastToken Token

	Line               []byte
	LineNumber, Column int
	Filename           string
}

func (scanner *Scanner) SetReader(reader io.Reader) {
	scanner.Reader = bufio.NewReader(reader)
}

func (scanner *Scanner) Peek() Token {
	scanner.NextToken = scanner.scan()
	return scanner.NextToken
}

func (scanner *Scanner) Unexpected() error {
	return errors.New("Unexpected " + string(scanner.Peek()))
}

func (scanner *Scanner) ScanIf(b byte) bool {
	var peek = scanner.Peek()
	if peek != nil && len(peek) > 0 && peek[0] == b {
		scanner.Scan()
		return true
	}
	return false
}

func (scanner *Scanner) Scan() Token {
	var token = scanner.scan()
	scanner.LastToken = token

	//Record line numbers, character position and the last line.
	scanner.Column += len(token)
	scanner.Line = append(scanner.Line, token...)
	if token.Is("\n") {
		scanner.Column = 0
		scanner.LineNumber++
		scanner.Line = nil
	}

	return token
}

func (scanner *Scanner) scan() Token {
	var token Token

	if scanner.NextToken != nil {
		defer func() {
			scanner.NextToken = nil
		}()
		return scanner.NextToken
	}

	for {
		peek, err := scanner.Reader.Peek(1)
		if err != nil {
			return nil
		}

		//Ignore whitespace
		if peek[0] == ' ' || peek[0] == '\t' {
			_, err = scanner.Reader.ReadByte()
			if err != nil {
				return nil
			}
			if len(token) > 0 {
				return token
			}
		} else {

			switch peek[0] {

			//Numerics
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				if len(token) > 0 {

					if _, err := strconv.Atoi(string(token)); err == nil {
						_, err = scanner.Reader.ReadByte()
						if err != nil {
							return nil
						}
						token = append(token, peek[0])
						continue
					} else {
						return token
					}
				} else {

					_, err = scanner.Reader.ReadByte()
					if err != nil {
						return nil
					}

					token = append(token, peek[0])
					continue
				}

			case ':', '\n', '(', ')', '{', '}', '[', ']', '.', ',':
				if len(token) > 0 {
					return token
				} else {

					_, err = scanner.Reader.ReadByte()
					if err != nil {
						return nil
					}

					return Token{peek[0]}
				}

			//Quotes
			case '"':
				if len(token) > 0 {
					return token
				} else {

					_, err = scanner.Reader.ReadByte()
					if err != nil {
						return nil
					}
					s, err := scanner.Reader.ReadBytes('"')
					if err != nil {
						return nil
					}

					return append(Token{'"'}, s...)
				}

			//Quotes
			case '\'':
				if len(token) > 0 {
					return token
				} else {

					_, err = scanner.Reader.ReadByte()
					if err != nil {
						return nil
					}
					s, err := scanner.Reader.ReadBytes('\'')
					if err != nil {
						return nil
					}

					return append(Token{'\''}, s...)
				}
			}

			if peek[0] == '/' {
				peek, err := scanner.Reader.Peek(2)
				if err != nil {
					return nil
				}

				//Comments
				if peek[1] == '/' {
					s, err := scanner.Reader.ReadBytes('\n')
					if err != nil {
						return nil
					}

					return s[:len(s)-1]
				}
			}

			_, err = scanner.Reader.ReadByte()
			if err != nil {
				return nil
			}
			token = append(token, peek[0])
		}
	}
}
