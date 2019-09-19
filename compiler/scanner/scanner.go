//Package scanner provides a 'i' syntax scanner.
package scanner

import (
	"bufio"
	"io"
)

//Scanner is an 'i' token scanner.
type Scanner struct {
	Reader  *bufio.Reader
	Readers []*bufio.Reader

	NextToken Token
	LastToken Token

	LastLine           []byte
	Line               []byte
	LineNumber, Column int
	Filename           string
}

//PushReader pushes a reader onto the stack.
func (scanner *Scanner) PushReader(reader io.Reader) {
	scanner.Readers = append(scanner.Readers, scanner.Reader)
	scanner.Reader = bufio.NewReader(reader)
}

//SetReader sets the reader for the scanner.
func (scanner *Scanner) SetReader(reader io.Reader) {
	scanner.Reader = bufio.NewReader(reader)
}

//Token returns the last scanned token.
func (scanner *Scanner) Token() Token {
	return scanner.LastToken
}

//Peek returns the next token without advancing the scanner.
func (scanner *Scanner) Peek() Token {
	next := scanner.scan()
	if next == nil {
		if len(scanner.Readers) > 0 {
			scanner.Reader = scanner.Readers[len(scanner.Readers)-1]
			scanner.Readers = scanner.Readers[:len(scanner.Readers)-1]
			return scanner.Peek()
		}
	}
	scanner.NextToken = next
	return scanner.NextToken
}

//ScanIf returns true and scans if the next token matches 'b'
func (scanner *Scanner) ScanIf(b byte) bool {
	var peek = scanner.Peek()
	//fmt.Println("peek", peek)
	if peek != nil && len(peek) > 0 && peek[0] == b {
		scanner.Scan()
		return true
	}
	return false
}

//ScanAndIgnoreNewLines advances the scanner and returns the next token (ignoring newline characters).
func (scanner *Scanner) ScanAndIgnoreNewLines() Token {
	for scanner.ScanIf('\n') {

	}
	return scanner.Scan()
}

//Scan advances the scanner and returns the next token.
func (scanner *Scanner) Scan() Token {
	var token = scanner.scan()
	//fmt.Println("scan", token)
	if token == nil {
		if len(scanner.Readers) > 0 {
			scanner.Reader = scanner.Readers[len(scanner.Readers)-1]
			scanner.Readers = scanner.Readers[:len(scanner.Readers)-1]
			return scanner.Scan()
		}
	}
	scanner.LastToken = token
	return token
}

func (scanner *Scanner) readByteRaw(ignorespace bool) error {
	//Record line numbers, character position and the last line.
	b, err := scanner.Reader.ReadByte()

	//Line numbers should always start at one.
	if scanner.LineNumber == 0 {
		scanner.LineNumber = 1
	}

	if b != '\t' || ignorespace {
		scanner.Column++
		scanner.Line = append(scanner.Line, b)
	}
	if b == '\n' && !ignorespace {
		scanner.Column = 0
		scanner.LineNumber++
		scanner.LastLine = scanner.Line
		scanner.Line = nil
	}

	return err
}

func (scanner *Scanner) readByte() error {
	return scanner.readByteRaw(false)
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

func (scanner *Scanner) readLiteral() ([]byte, error) {
	var result = []byte{'`'}

	for {
		b, err := scanner.Reader.Peek(1)
		if err != nil {
			return nil, err
		}

		if err := scanner.readByteRaw(true); err != nil {
			return nil, err
		}

		result = append(result, b[0])
		if b[0] == '`' {
			break
		}
	}

	return result, nil
}

func (scanner *Scanner) readNumber() ([]byte, error) {
	var result = []byte{}

	for {
		b, err := scanner.Reader.Peek(1)
		if err != nil {
			return result, err
		}

		switch b[0] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'x':
		default:
			return result, nil
		}

		if err := scanner.readByte(); err != nil {
			return result, err
		}

		result = append(result, b[0])
	}
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
			return token
		}

		//Ignore whitespace
		if peek[0] == ' ' || peek[0] == '\t' {
			if err := scanner.readByte(); err != nil {
				return token
			}
			if len(token) > 0 {
				return token
			}
		} else {

			switch peek[0] {

			//Numerics
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				s, err := scanner.readNumber()
				if err != nil {
					return s
				}
				return s
			case '/':
				b, err := scanner.Reader.Peek(2)
				if err != nil {
					return token
				}
				if b[1] == '/' {
					break
				}
				fallthrough

			//These symbols break a token.
			case ':', '\n', '(', ')', '{', '}', '[', ']', '.', ',', '$', '#',
				'+', '-', '*', '%', '=', '|', '^', '&', '!', '?', '<', '>', '~', '_', ';':
				if len(token) > 0 {
					return token //This is an endquote.
				}

				if err := scanner.readByte(); err != nil {
					return token
				}

				return Token{peek[0]}

			//Quotes
			case '"':
				if len(token) > 0 {
					return token //This is an endquote.
				}

				if err := scanner.readByte(); err != nil {
					return token
				}
				s, err := scanner.readString()
				if err != nil {
					return token
				}

				return s

			//Literals
			case '`':
				if len(token) > 0 {
					return token //This is an endquote.
				}

				if err := scanner.readByte(); err != nil {
					return token
				}
				s, err := scanner.readLiteral()
				if err != nil {
					return token
				}

				return s

			//Symbols
			case '\'':
				if len(token) > 0 {
					return token //This is an endquote.
				}

				if err := scanner.readByte(); err != nil {
					return token
				}
				s, err := scanner.readSymbol()
				if err != nil {
					return token
				}

				return s
			}

			if peek[0] == '/' {
				peek, err := scanner.Reader.Peek(2)
				if err != nil {
					return token
				}

				//Comments
				if peek[1] == '/' {
					s, err := scanner.readLine()
					if err != nil {
						return token
					}

					return s[:len(s)-1]
				}
			}

			if err := scanner.readByte(); err != nil {
				return token
			}
			token = append(token, peek[0])
		}
	}
}
