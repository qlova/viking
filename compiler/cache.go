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
		if err == io.EOF {

			//Return to the last frame.
			if len(compiler.Frames) > 0 {
				var context = compiler.Frames[len(compiler.Frames)-1]
				compiler.Context = context
				compiler.Frames = compiler.Frames[:len(compiler.Frames)-1]
				return nil
			}

			return nil
		} else if err != nil {
			return err
		}

		compiler.Write([]byte("\n"))
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
		for {
			var token = compiler.Scan()

			if token.Is("\n") {
				break
			}

			cache.WriteByte(' ')
			cache.Write(token)
			cache.WriteByte(' ')
		}
		cache.WriteByte('}')
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
				case "for", "if", "catch", "try":
					depth++
				}

			}

			cache.WriteByte(' ')
			cache.Write(token)
			cache.WriteByte(' ')
		}
		cache.WriteByte('}')
	}

	return cache
}
