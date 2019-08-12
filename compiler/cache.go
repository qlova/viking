package compiler

import (
	"bytes"
	"io"
)

//Cache is a storage container that contains code. It can be compiled at a later point in time.
type Cache struct {
	bytes.Buffer

	Filename   string
	LineNumber int

	compiler *Compiler
}

//CompileCacheWithContext compiles a cache with the new context.
func (compiler *Compiler) CompileCacheWithContext(cache Cache, context Context) error {
	compiler.PushContext(context)
	compiler.SetReader(&cache.Buffer)
	compiler.Filename = cache.Filename
	compiler.LineNumber = cache.LineNumber

	for {
		err := compiler.CompileStatement()
		if err != nil {
			//Return to the last frame.
			if len(compiler.Frames) > 0 {
				compiler.PopContext()

				if err != io.EOF {
					return err
				}

				return nil
			} else if err == io.EOF {
				return nil
			} else {
				return err
			}
		}

		compiler.Go.Write([]byte("\n"))
	}
}

//CacheBlock create a cache out of the next 'i' block.
func (compiler *Compiler) CacheBlock() Cache {
	var cache Cache

	cache.Filename = compiler.Filename
	cache.LineNumber = compiler.LineNumber

	var sort = compiler.Scan()

	var depth = 1

	if sort.Is(":") {
		var column = compiler.Column
		for {
			var token = compiler.Scan()

			if token.Is("\n") || token == nil {
				cache.Write(compiler.LastLine[column:])
				cache.WriteString("}")
				break
			}
		}
	} else {
		if sort.Is("{") {
			compiler.Scan()
		}

		cache.LineNumber++
		for {
			var token = compiler.Scan()

			if token.Is("\n") {
				cache.Write(compiler.LastLine)
			}

			if token.Is(":") || token.Is("}") || token == nil {
				depth--
				if depth == 0 {
					if len(compiler.Line) > 0 {
						cache.Write(compiler.Line[:len(compiler.Line)-1])
					}
					cache.WriteString("}")
					break
				}
			} else {
				switch token.String() {
				case "for", "if", "catch", "try", "{", "main":
					depth++
				}

			}
		}
	}

	return cache
}
