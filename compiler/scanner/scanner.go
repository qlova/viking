//Package scanner provides a 'i' syntax scanner.
package scanner

import (
	"bufio"
	"io"
	"strconv"
)

//Scanner is an 'i' token scanner.
type Scanner struct {
	Reader *bufio.Reader

	NextToken Token
	LastToken Token

	LastLine           []byte
	Line               []byte
	LineNumber, Column int
	Filename           string
}

//SetReader sets the reader for the scanner.
func (scanner *Scanner) SetReader(reader io.Reader) {
	scanner.Reader = bufio.NewReader(reader)
}

//Peek returns the next token without advancing the scanner.
func (scanner *Scanner) Peek() Token {
	scanner.NextToken = scanner.scan()
	return scanner.NextToken
}

//ScanIf returns true and scans if the next token matches 'b'
func (scanner *Scanner) ScanIf(b byte) bool {
	var peek = scanner.Peek()
	if peek != nil && len(peek) > 0 && peek[0] == b {
		scanner.Scan()
		return true
	}
	return false
}

//Scan advances the scanner and returns the next token.
func (scanner *Scanner) Scan() Token {
	var token = scanner.scan()
	scanner.LastToken = token
	return token
}

func (scanner *Scanner) readByte() error {

	//Record line numbers, character position and the last line.
	b, err := scanner.Reader.ReadByte()

	//Line numbers should always start at one.
	if scanner.LineNumber == 0 {
		scanner.LineNumber = 1
	}

	if b != '\t' {
		scanner.Column++
		scanner.Line = append(scanner.Line, b)
	}
	if b == '\n' {
		scanner.Column = 0
		scanner.LineNumber++
		scanner.LastLine = scanner.Line
		scanner.Line = nil
	}

	return err
}

func (scanner *Scanner) readString() ([]byte, error) {
	var result = []byte{'"'}

	for {
		b, err := scanner.Reader.Peek(1)
		if err != nil {
			return nil, err
		}

		if err := scanner.readByte(); err != nil {
			return nil, err
		}

		result = append(result, b[0])
		if b[0] == '"' {
			break
		}
	}

	return result, nil
}

func (scanner *Scanner) readLine() ([]byte, error) {
	var result []byte

	for {
		b, err := scanner.Reader.Peek(1)
		if err != nil {
			return nil, err
		}

		if err := scanner.readByte(); err != nil {
			return nil, err
		}

		result = append(result, b[0])
		if b[0] == '\n' {
			break
		}
	}

	return result, nil
}

func (scanner *Scanner) readSymbol() ([]byte, error) {
	var result = []byte{'\''}

	for {
		b, err := scanner.Reader.Peek(1)
		if err != nil {
			return nil, err
		}

		if err := scanner.readByte(); err != nil {
			return nil, err
		}

		result = append(result, b[0])
		if b[0] == '\'' {
			break
		}
	}

	return result, nil
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
			if err := scanner.readByte(); err != nil {
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
						if err := scanner.readByte(); err != nil {
							return nil
						}
						token = append(token, peek[0])
						continue
					} else {
						return token
					}
				} else {

					if err := scanner.readByte(); err != nil {
						return nil
					}

					token = append(token, peek[0])
					continue
				}
			case '/':
				b, err := scanner.Reader.Peek(2)
				if err != nil {
					return nil
				}
				if b[1] == '/' {
					break
				}
				fallthrough

			//These symbols break a token.
			case ':', '\n', '(', ')', '{', '}', '[', ']', '.', ',', '$', '#',
				'+', '-', '*', '%', '=':
				if len(token) > 0 {
					return token //This is an endquote.
				}

				if err := scanner.readByte(); err != nil {
					return nil
				}

				return Token{peek[0]}

			//Quotes
			case '"':
				if len(token) > 0 {
					return token //This is an endquote.
				}

				if err := scanner.readByte(); err != nil {
					return nil
				}
				s, err := scanner.readString()
				if err != nil {
					return nil
				}

				return s

			//Symbols
			case '\'':
				if len(token) > 0 {
					return token //This is an endquote.
				}

				if err := scanner.readByte(); err != nil {
					return nil
				}
				s, err := scanner.readSymbol()
				if err != nil {
					return nil
				}

				return s
			}

			if peek[0] == '/' {
				peek, err := scanner.Reader.Peek(2)
				if err != nil {
					return nil
				}

				//Comments
				if peek[1] == '/' {
					s, err := scanner.readLine()
					if err != nil {
						return nil
					}

					return s[:len(s)-1]
				}
			}

			if err := scanner.readByte(); err != nil {
				return nil
			}
			token = append(token, peek[0])
		}
	}
}
