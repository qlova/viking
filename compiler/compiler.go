package compiler

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/qlova/viking/compiler/target"
)

//Ilang is the Ilang package for Go.
const Ilang = "github.com/qlova/i"

//ReservedWords are not available for use as names.
var ReservedWords = []string{"if", "for", "return", "break", "go", "in"}

//Set is a string set.
type Set map[string]struct{}

func (set Set) Get(key string) bool {
	_, ok := set[key]
	return ok
}

func (set Set) Set(key string) {
	set[key] = struct{}{}
}

//Compiler is an 'i' compiler.
type Compiler struct {
	Context
	target.Buffer

	Target target.Buffer

	ExpectedOutput []byte
	ProvidedInput  []byte

	Imports      Set
	Dependencies Set

	Requirements struct {
		Head Set
		Neck Set
		Tail Set
	}

	Packages map[string]Package

	Frames []Context

	Buffers []target.Buffer

	//signals for file processing.
	yield, callback chan bool

	Main bool
}

//New returns a new initialised compiler.
func New() Compiler {
	var c Compiler
	c.Init()
	return c
}

//Init initialises the compiler.
func (compiler *Compiler) Init() {
	compiler.Functions = make(map[string]*Type)
	compiler.Concepts = make(map[string]Concept)
	compiler.Aliases = make(map[string]Alias)
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

		if pkg == Ilang {
			compiler.Go.Head.Write([]byte(`import I "` + pkg + `"`))
			compiler.Go.Head.Write([]byte("\n"))
			return
		}

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
	if token.Is("\n") || token.Is("") {
		return nil
	}
	if len(token) >= 2 && (token[0] == '/' && token[1] == '/') {
		compiler.Go.Write(s(" "))
		compiler.Go.Write(token)
		return nil
	}
	return compiler.NewError("newline expected but found: " + string(token))
}

//Compile package located at Compiler.Dir or current working directory if empty.
func (compiler *Compiler) Compile() error {
	pkg, err := os.Open(compiler.Directory)
	if err != nil {
		return err
	}

	if info, err := pkg.Stat(); err != nil {
		return err
	} else if !info.IsDir() {
		return compiler.CompileFile(compiler.Directory)
	}

	files, err := ioutil.ReadDir(compiler.Directory)
	if err != nil {
		return err
	}

	compiler.Go.Head.Write([]byte("package main\n\n"))

	var next = make(chan bool)
	var done = make(chan error, 1)

	compiler.yield = make(chan bool)
	compiler.callback = make(chan bool)

	var yielded = false

	for _, file := range files {
		if path.Ext(file.Name()) == ".i" {
			go func() {
				if yielded {
					compiler.PushContext(compiler.NewContext())
				}
				compilerError := compiler.CompileFile(compiler.Directory + "/" + file.Name())
				if compiler.Main {
					done <- compilerError
					next <- true
					return
				}
				if compilerError != nil {
					err = compilerError
				}
				next <- true
			}()

		loop:
			for {
				select {
				case <-compiler.yield:
					yielded = true
					break loop
				case <-next:
					break loop
				}
			}
		}
	}

	if yielded {
		compiler.callback <- true
	} else {
		done <- nil
	}

	if err != nil {
		<-done
		return err
	}

	return <-done
}

//CompileBlock compiles an 'i' code block.
func (compiler *Compiler) CompileBlock() error {
	if compiler.ScanIf(':') {
		defer func() {
			compiler.LoseScope()
			compiler.Go.Write([]byte("}"))
		}()
		if err := compiler.CompileStatement(); err != nil {
			return err
		}
		return nil
	}

	if !compiler.ScanIf('\n') {
		return compiler.NewError("block must start with a newline")
	}

	return nil
}

//CompileFile compiles a file.
func (compiler *Compiler) CompileFile(location string) error {
	compiler.Filename = filepath.Base(location)

	file, err := os.Open(location)
	if err != nil {
		//Return to the last frame.
		if len(compiler.Frames) > 0 {
			compiler.PopContext()
		}
		return err
	}
	defer file.Close()

	return compiler.CompileReader(file)
}

//CompileReader compiles a reader.
func (compiler *Compiler) CompileReader(reader io.Reader) error {
	if reader == nil {
		return compiler.NewError("null reader")
	}

	compiler.SetReader(reader)

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
