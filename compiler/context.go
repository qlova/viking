package compiler

import "github.com/qlova/viking/compiler/scanner"

//Token is a scanner Token.
type Token = scanner.Token

//Context contains current context information for the compiler.
type Context struct {
	scanner.Scanner
	Directory string

	Export bool

	//Type definitions.
	TypeName             Token
	InsideTypeDefinition bool
	TypeDefinition       Type

	Returns *Type
}

//NewContext pushes a new context to the compiler.
func (compiler *Compiler) NewContext() {
	var ctx Context
	ctx.Returns = &Type{}
	compiler.PushContext(ctx)
}

//PushContext pushes the specified context to the compiler.
func (compiler *Compiler) PushContext(context Context) {
	var directory = compiler.Directory
	compiler.Frames = append(compiler.Frames, compiler.Context)
	compiler.Context = context
	compiler.Directory = directory
}
