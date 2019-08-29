package compiler

import "github.com/qlova/viking/compiler/scanner"

//Token is a scanner Token.
type Token = scanner.Token

//Context contains current context information for the compiler.
//A new context is created for each file/package being compiled.
type Context struct {
	scanner.Scanner
	Directory string

	Export bool

	//Type definitions.
	TypeName             Token
	InsideTypeDefinition bool
	TypeDefinition       Type

	//Functions is the defined global functions available to this context.
	Functions map[string]*Type
	Concepts  map[string]Concept

	Depth  int
	Depths []int

	Scope  []Scope
	Scopes [][]Scope

	Aliases map[string]Alias

	Returns *Type

	//Does the current statement throw?
	Throws bool
}

//NewContext pushes a new context to the compiler.
func (compiler *Compiler) NewContext() Context {
	var ctx Context
	ctx.Returns = &Type{}
	ctx.Concepts = compiler.Concepts
	ctx.Functions = compiler.Functions
	ctx.Aliases = compiler.Aliases
	ctx.Directory = compiler.Directory
	return ctx
}

//NewPackageContext pushes a new context to the compiler with unique functions and concept maps.
func (compiler *Compiler) NewPackageContext() Context {
	var ctx Context
	ctx.Returns = &Type{}
	ctx.Concepts = make(map[string]Concept)
	ctx.Functions = make(map[string]*Type)
	ctx.Aliases = make(map[string]Alias)
	ctx.Directory = compiler.Directory
	return ctx
}

//PushContext pushes the specified context to the compiler.
func (compiler *Compiler) PushContext(context Context) {
	compiler.Frames = append(compiler.Frames, compiler.Context)
	compiler.Context = context

}

//PopContext pops the current context of the compiler and returns the previous context.
func (compiler *Compiler) PopContext() {
	var context = compiler.Frames[len(compiler.Frames)-1]
	compiler.Context = context
	compiler.Frames = compiler.Frames[:len(compiler.Frames)-1]
}

//GainScope gains a new scope level.
func (compiler *Context) GainScope() {
	compiler.Depth++
	compiler.Scope = append(compiler.Scope, NewScope())
}
