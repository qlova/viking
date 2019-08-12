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

	//Functions is the defined global functions available to this context.
	Functions map[string]struct{}
	Concepts  map[string]Concept

	Depth  int
	Depths []int

	Scope  []Scope
	Scopes [][]Scope

	Returns *Type
}

//NewContext pushes a new context to the compiler.
func (compiler *Compiler) NewContext() Context {
	var ctx Context
	ctx.Returns = &Type{}
	ctx.Concepts = compiler.Concepts
	ctx.Functions = compiler.Functions
	return ctx
}

//PushContext pushes the specified context to the compiler.
func (compiler *Compiler) PushContext(context Context) {
	var directory = compiler.Directory
	compiler.Frames = append(compiler.Frames, compiler.Context)
	compiler.Context = context
	compiler.Directory = directory
}

//GainScope gains a new scope level.
func (compiler *Context) GainScope() {
	compiler.Depth++
	compiler.Scope = append(compiler.Scope, NewScope())
}
