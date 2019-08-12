package compiler

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"viking/compiler/target"
)

//ReservedWords are not available for use as names.
var ReservedWords = []string{"if", "for", "return", "break", "go", "in"}

//Compiler is an 'i' compiler.
type Compiler struct {
	Context
	target.Buffer

	Target target.Buffer

	Depth  int
	Depths []int

	ExpectedOutput []byte
	ProvidedInput  []byte

	Imports      map[string]struct{}
	Dependencies map[string]struct{}
	Functions    map[string]struct{}

	Scope  []Scope
	Scopes [][]Scope

	Frames []Context

	Buffers []target.Buffer

	Concepts map[string]Concept
}

//NewScope creates and returns a new compiler scope.
func NewScope() Scope {
	return Scope{
		Table: make(map[string]Type),
	}
}

//SetTarget sets the compiler target.
func (compiler *Compiler) SetTarget(target target.Buffer) {
	compiler.Target = target
	compiler.Buffer = target
}

//PushScope pushes a new scopeset to the compiler.
func (compiler *Compiler) PushScope() {
	compiler.Scopes = append(compiler.Scopes, compiler.Scope)
	compiler.Scope = nil

	compiler.Depths = append(compiler.Depths, compiler.Depth)
	compiler.Depth = 0
}

//PopScope pops the last scopeset from the compiler.
func (compiler *Compiler) PopScope() {
	compiler.Scope = compiler.Scopes[len(compiler.Scopes)-1]
	compiler.Scopes = compiler.Scopes[:len(compiler.Scopes)-1]

	compiler.Depth = compiler.Depths[len(compiler.Depths)-1]
	compiler.Depths = compiler.Depths[:len(compiler.Depths)-1]
}

//Scope can contain variables and Cleanup routines.
type Scope struct {
	Table    map[string]Type
	Cleanups []func()
}

//DeferCleanup schedules the function to run at the end of the current scope.
func (compiler *Compiler) DeferCleanup(f func()) {
	var scope = len(compiler.Scope) - 1
	compiler.Scope[scope].Cleanups = append(compiler.Scope[scope].Cleanups, f)
}

//FlipBuffer creates a new buffer that future writes will write to, returns the old buffer.
func (compiler *Compiler) FlipBuffer() *target.Buffer {
	var current = compiler.Buffer

	compiler.Buffers = append(compiler.Buffers, current)

	compiler.Buffer = compiler.Target

	return &compiler.Buffers[len(compiler.Buffers)-1]
}

//DumpBuffer collapses the current buffer onto the old one.
func (compiler *Compiler) DumpBuffer() {
	var last = compiler.Buffers[len(compiler.Buffers)-1]

	last.Go.Head.Write(compiler.Go.Head.Bytes())
	last.Go.Neck.Write(compiler.Go.Neck.Bytes())
	last.Go.Write(compiler.Go.Bytes())
	last.Go.Tail.Write(compiler.Go.Tail.Bytes())

	compiler.Buffer = last

	compiler.Buffers = compiler.Buffers[:len(compiler.Buffers)-1]
}

//DumpBufferHead collapses the current buffer onto the old one, writing the body onto the head.
func (compiler *Compiler) DumpBufferHead(split []byte) {

	var last = compiler.Buffers[len(compiler.Buffers)-1]

	last.Go.Head.Write(compiler.Go.Head.Bytes())
	last.Go.Neck.Write(compiler.Go.Neck.Bytes())
	last.Go.Neck.Write(split)
	last.Go.Neck.Write(compiler.Go.Bytes())
	last.Go.Tail.Write(compiler.Go.Tail.Bytes())

	compiler.Buffer = last

	compiler.Buffers = compiler.Buffers[:len(compiler.Buffers)-1]
}

//GainScope gains a new scope level.
func (compiler *Compiler) GainScope() {
	compiler.Depth++
	compiler.Scope = append(compiler.Scope, NewScope())
}

//LoseScope loses a scope level.
func (compiler *Compiler) LoseScope() {

	//Cleanup
	if compiler.InsideTypeDefinition {
		compiler.Indent()
		compiler.Go.Write(s("return " + compiler.TypeName.String() + "{}\n"))

		//We need to create the Go code for this type definition.
		var T = compiler.TypeDefinition

		compiler.Go.Neck.Write(s("type " + compiler.TypeName.String() + " struct {\n"))
		compiler.Indent(&compiler.Go.Neck)
		for _, field := range T.Fields {
			compiler.Go.Neck.Write(s(field.Name + " "))
			compiler.Go.Neck.Write(GoTypeOf(field.Type))
			compiler.Go.Neck.Write(s("\n"))
		}
		compiler.Go.Neck.Write(s("}\n\n"))

		compiler.TypeDefinition = Type{}
	}

	var scope = compiler.Scope[len(compiler.Scope)-1]
	for _, cleanup := range scope.Cleanups {
		cleanup()
	}

	compiler.Scope = compiler.Scope[:len(compiler.Scope)-1]
	compiler.Depth--
}

//Import imports a Go package to the Buffer's Neck if it hasn't already been imported.
func (compiler *Compiler) Import(pkg string) {
	if compiler.Imports == nil {
		compiler.Imports = make(map[string]struct{})
	}

	if _, ok := compiler.Imports[pkg]; !ok {
		compiler.Imports[pkg] = struct{}{}
		compiler.Go.Head.Write([]byte(`import "` + pkg + `"`))
		compiler.Go.Head.Write([]byte("\n"))
	}
}

//Require writes the provided code to the neck of the buffer if it hasn't been required already.
func (compiler *Compiler) Require(code string) {
	if compiler.Dependencies == nil {
		compiler.Dependencies = make(map[string]struct{})
	}

	if _, ok := compiler.Dependencies[code]; !ok {
		compiler.Dependencies[code] = struct{}{}

		compiler.FlipBuffer()
		compiler.Go.Neck.Write([]byte(code))
		compiler.DumpBuffer()
	}
}

//Indent writes indentation to the body of the compiler's output.
func (compiler *Compiler) Indent(writers ...io.Writer) {
	if len(writers) == 0 {
		for i := 0; i < compiler.Depth; i++ {
			compiler.Go.Write([]byte{'\t'})
		}
	} else {
		var writer = writers[0]
		for i := 0; i < compiler.Depth; i++ {
			writer.Write([]byte{'\t'})
		}
	}
}

//ScanLine attempts to scan a newline, an error is returned if no newline is found.
func (compiler *Compiler) ScanLine() error {
	var token = compiler.Scan()
	if token.Is("\n") {
		return nil
	}
	if len(token) >= 2 && (token[0] == '/' && token[1] == '/') {
		compiler.Go.Write(s(" "))
		compiler.Go.Write(token)
		return nil
	}
	return errors.New("newline expected but found: " + string(token))
}

//Compile package located at Compiler.Dir or current working directory if empty.
func (compiler *Compiler) Compile() error {
	files, err := ioutil.ReadDir(compiler.Directory)
	if err != nil {
		return Error{compiler, err}
	}

	compiler.Go.Head.Write([]byte("package main\n\n"))

	for _, file := range files {
		if path.Ext(file.Name()) == ".i" {
			err := compiler.CompileFile(file.Name())
			if err != nil {
				return Error{compiler, err}
			}
		}
	}

	return nil
}

//CompileBlock compiles an 'i' code block.
func (compiler *Compiler) CompileBlock() error {
	if compiler.ScanIf(':') {
		defer func() {
			compiler.LoseScope()
			compiler.Go.Write([]byte("}"))
		}()
		return compiler.CompileStatement()
	}

	if !compiler.ScanIf('\n') {
		return errors.New("block must start with a newline")
	}

	return nil
}

//CompileFile compiles a file.
func (compiler *Compiler) CompileFile(location string) error {
	file, err := os.Open(path.Join(compiler.Directory, location))
	if err != nil {
		return err
	}
	defer file.Close()

	return compiler.CompileReader(file)
}

//CompileReader compiles a reader.
func (compiler *Compiler) CompileReader(reader io.Reader) error {
	if reader == nil {
		return errors.New("null reader")
	}

	compiler.SetReader(reader)

	for {
		err := compiler.CompileStatement()
		if err != nil {
			//Return to the last frame.
			if len(compiler.Frames) > 0 {
				var context = compiler.Frames[len(compiler.Frames)-1]
				compiler.Context = context
				compiler.Frames = compiler.Frames[:len(compiler.Frames)-1]

				if err != io.EOF {
					return err
				}

				continue
			} else {
				return nil
			}
		}

		compiler.Go.Write([]byte("\n"))
	}
}

//WriteTo writes the compiler's output buffer to the specified buffer.
func (compiler *Compiler) WriteTo(writer io.Writer) (int64, error) {
	var sum int
	if n, err := writer.Write(compiler.Go.Head.Bytes()); err != nil {
		return int64(n), err
	} else {
		sum += n
	}
	if n, err := writer.Write(compiler.Go.Neck.Bytes()); err != nil {
		return int64(n), err
	} else {
		sum += n
	}
	if n, err := writer.Write(compiler.Go.Bytes()); err != nil {
		return int64(n), err
	} else {
		sum += n
	}
	if n, err := writer.Write(compiler.Go.Tail.Bytes()); err != nil {
		return int64(n), err
	} else {
		sum += n
	}
	return int64(sum), nil
}
