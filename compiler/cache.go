package compiler

import "bytes"

//A cache is a storage container that contains code. It can be compiled at a later point in time.
type Cache struct {
	bytes.Buffer

	Filename   string
	LineNumber int

	compiler *Compiler
}

func (compiler *Compiler) CacheBlock() Cache {
	var cache Cache

	cache.Filename = compiler.Filename
	cache.LineNumber = compiler.LineNumber

	var sort = compiler.Scan()

	var depth = 1

	if sort.Is(":") {
		for {
			var token = compiler.Scan()

			if token.Is("\n") {
				break
			}

			cache.WriteByte(' ')
			cache.Write(token)
			cache.WriteByte(' ')
		}
	} else {
		for {
			var token = compiler.Scan()

			if token.Is("}") {
				depth--
				if depth == 0 {
					break
				}
			} else {
				switch token.String() {
				case "for", "if":
					depth++
				}

			}

			cache.WriteByte(' ')
			cache.Write(token)
			cache.WriteByte(' ')
		}
	}

	return cache
}
