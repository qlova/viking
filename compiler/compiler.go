package compiler

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
)

//ReservedWords are not available for use as names.
var ReservedWords = []string{"if", "for", "return", "break", "go", "in"}

//Compiler is an 'i' compiler.
type Compiler struct {
	Context
	Buffer

	Depth  int
	Depths []int

	ExpectedOutput []byte

	Imports map[string]struct{}

	Scope  []Scope
	Scopes [][]Scope

	Frames []Context

	Buffers []Buffer

	Concepts map[string]Concept
}

//NewScope creates and returns a new compiler scope.
func NewScope() Scope {
	return Scope{
		Table: make(map[string]Type),
	}
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

//Buffer is where the compiler writes data to.
type Buffer struct {
	Head, Neck, Body bytes.Buffer
}

//FlipBuffer creates a new buffer that future writes will write to, returns the old buffer.
func (compiler *Compiler) FlipBuffer() *Buffer {
	var current = compiler.Buffer

	compiler.Buffers = append(compiler.Buffers, current)

	compiler.Buffer = Buffer{}

	return &compiler.Buffers[len(compiler.Buffers)-1]
}

//DumpBuffer collapses the current buffer onto the old one.
func (compiler *Compiler) DumpBuffer() {
	var last = compiler.Buffers[len(compiler.Buffers)-1]

	last.Neck.Write(compiler.Neck.Bytes())
	last.Head.Write(compiler.Head.Bytes())
	last.Body.Write(compiler.Body.Bytes())

	compiler.Buffer = last

	compiler.Buffers = compiler.Buffers[:len(compiler.Buffers)-1]
}

//DumpBufferHead collapses the current buffer onto the old one, writing the body onto the head.
func (compiler *Compiler) DumpBufferHead() {

	var last = compiler.Buffers[len(compiler.Buffers)-1]

	last.Head.Write(compiler.Head.Bytes())
	last.Neck.Write(compiler.Neck.Bytes())
	last.Head.Write(compiler.Body.Bytes())

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
		compiler.Write(s("return " + compiler.TypeName.String() + "{}\n"))

		//We need to create the Go code for this type definition.
		var T = compiler.TypeDefinition

		compiler.Head.Write(s("type " + compiler.TypeName.String() + " struct {\n"))
		compiler.Indent(&compiler.Head)
		for _, field := range T.Fields {
			compiler.Head.Write(s(field.Name + " "))
			compiler.Head.Write(GoTypeOf(field.Type))
			compiler.Head.Write(s("\n"))
		}
		compiler.Head.Write(s("}\n\n"))

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
		compiler.Neck.Write([]byte(`import "` + pkg + `"`))
		compiler.Neck.Write([]byte("\n"))
	}
}

//Write writes bytes to the body of the compiler output.
func (buffer *Buffer) Write(data []byte) {
	buffer.Body.Write(data)
}

//WriteString writes a string to the body of the compiler output.
func (buffer *Buffer) WriteString(s string) {
	buffer.Body.Write([]byte(s))
}

//WriteLine writes a newline to the body of the compiler's output.
func (buffer *Buffer) WriteLine() {
	buffer.Body.Write([]byte{'\n'})
}

//Indent writes indentation to the body of the compiler's output.
func (compiler *Compiler) Indent(writers ...io.Writer) {
	if len(writers) == 0 {
		for i := 0; i < compiler.Depth; i++ {
			compiler.Write([]byte{'\t'})
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
		compiler.Write(s(" "))
		compiler.Write(token)
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

	compiler.Neck.Write([]byte("package main\n\n"))

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
			compiler.Write([]byte("}"))
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
	compiler.SetReader(reader)

	for {
		err := compiler.CompileStatement()
		if err == io.EOF {

			//Return to the last frame.
			if len(compiler.Frames) > 0 {
				var context = compiler.Frames[len(compiler.Frames)-1]
				compiler.Context = context
				compiler.Frames = compiler.Frames[:len(compiler.Frames)-1]
				continue
			}

			return nil
		} else if err != nil {
			return err
		}

		compiler.Write([]byte("\n"))
	}
}

//WriteTo writes the compiler's output buffer to the specified buffer.
func (compiler *Compiler) WriteTo(writer io.Writer) (int64, error) {

	n, err := writer.Write(compiler.Neck.Bytes())
	if err != nil {
		return int64(n), err
	}

	n2, err := writer.Write(compiler.Head.Bytes())
	if err != nil {
		return int64(n2), err
	}

	writer.Write([]byte{'\n'})

	n3, err := writer.Write(compiler.Body.Bytes())
	if err != nil {
		return int64(n3), err
	}

	return int64(n + n2 + n3), nil
}
