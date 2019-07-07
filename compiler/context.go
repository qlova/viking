package compiler

import "github.com/qlova/viking/compiler/scanner"

type Context struct {
	scanner.Scanner
	Directory string

	Export bool

	//Type definitions.
	TypeName             scanner.Token
	InsideTypeDefinition bool
	TypeDefinition       Type
}

func (compiler *Compiler) NewContext() {
	var directory = compiler.Directory
	compiler.Frames = append(compiler.Frames, compiler.Context)
	compiler.Context = Context{}
	compiler.Directory = directory
}
