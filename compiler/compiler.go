package compiler

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
)

var ReservedWords = []string{"if", "for", "return", "break", "go", "in"}

func Unimplemented(component []byte) error {
	return errors.New("Unimplemented " + string(component))
}

func (compiler *Compiler) Expecting(symbol byte) error {
	return errors.New("Expecting " + string(symbol))
}

type Compiler struct {
	Context

	Depth int

	Head bytes.Buffer
	Body bytes.Buffer

	ExpectedOutput []byte

	Imports map[string]struct{}

	Scope []map[string]Type

	Frames []Context
}

func (compiler *Compiler) GainScope() {
	compiler.Depth++
	compiler.Scope = append(compiler.Scope, make(map[string]Type))
}

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

	compiler.Scope = compiler.Scope[:len(compiler.Scope)-1]
	compiler.Depth--
}

func (compiler *Compiler) Import(pkg string) {
	if compiler.Imports == nil {
		compiler.Imports = make(map[string]struct{})
	}

	if _, ok := compiler.Imports[pkg]; !ok {
		compiler.Imports[pkg] = struct{}{}
		compiler.Head.Write([]byte(`import "` + pkg + `"`))
		compiler.Head.Write([]byte("\n"))
	}
}

func (compiler *Compiler) Write(data []byte) {
	compiler.Body.Write(data)
}

func (compiler *Compiler) WriteLine() {
	compiler.Body.Write([]byte{'\n'})
}

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
	return errors.New("Newline expected but found: " + string(token))
}

//Compile package located at Compiler.Dir or current working directory if empty.
func (compiler *Compiler) Compile() error {
	files, err := ioutil.ReadDir(compiler.Directory)
	if err != nil {
		return err
	}

	compiler.Head.Write([]byte("package main\n\n"))

	for _, file := range files {
		if path.Ext(file.Name()) == ".i" {
			err := compiler.CompileFile(file.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (compiler *Compiler) CompileBlock() error {
	compiler.GainScope()

	if compiler.ScanIf(':') {
		defer func() {
			compiler.LoseScope()
			compiler.Write([]byte("}"))
		}()
		return compiler.CompileStatement()
	} else {

		if !compiler.ScanIf('\n') {
			return compiler.Unexpected()
		}

		compiler.Depth++
	}

	return nil
}

func (compiler *Compiler) CompileFile(location string) error {
	file, err := os.Open(path.Join(compiler.Directory, location))
	if err != nil {
		return err
	}
	defer file.Close()

	return compiler.CompileReader(file)
}

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

func (compiler *Compiler) WriteTo(writer io.Writer) error {
	_, err := writer.Write(compiler.Head.Bytes())
	if err != nil {
		return err
	}

	writer.Write([]byte{'\n'})

	_, err = writer.Write(compiler.Body.Bytes())
	if err != nil {
		return err
	}

	return nil
}
