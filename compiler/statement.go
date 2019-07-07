package compiler

import (
	"bytes"
	"errors"
	"io"
)

func (compiler *Compiler) CompileStatement() error {
	var token = compiler.Scan()
	/*switch token {
		case "if", "for", "return", "break", "go":
			return Unimplemented
	}*/

	if token == nil {
		return io.EOF
	}

	if bytes.Equal(token, []byte("\n")) {
		return nil
	}

	if len(token) > 2 && token[0] == '/' && token[1] == '/' {
		compiler.Write(token)

		if len(token) > len("//output: ") && bytes.Equal(token[:len("//output: ")], []byte("//output: ")) {
			compiler.ExpectedOutput = token[len("//output: "):]
		}

		return nil
	}

	switch token.String() {

	case "import":
		var namespace = compiler.Scan()
		compiler.NewContext()
		compiler.CompileFile(string(namespace) + ".i")
		return nil

	case ".":
		compiler.Export = true
		defer func() {
			compiler.Export = false
		}()
		return compiler.CompileStatement()

	case "main":
		compiler.Write(s("func main() {"))
		return compiler.CompileBlock()

	case "}":
		compiler.Depth--
		compiler.Write(s("}"))
		compiler.LoseScope()
		return nil

	}

	if Builtin(token) {
		compiler.Indent()
		return compiler.CompileBuiltin(token)
	}

	if compiler.Peek().Is("[") {
		if Defined(compiler.GetVariable(token)) {
			compiler.Indent()
			return compiler.ModifyArray(token)
		}
	}

	if compiler.Peek().Is("=") {
		compiler.Indent()
		if Defined(compiler.GetVariable(token)) {
			return compiler.AssignVariable(token)
		} else {
			return compiler.DefineVariable(token)
		}
	}

	//Embedded types.
	if T := compiler.GetType(token); Defined(T) {
		if !compiler.InsideTypeDefinition {
			return errors.New("Cannnot embed type here, are you in a type definition?")
		}

		//Is this is a collection then there will be a dot.
		if compiler.Peek().Is(".") {

			var subtype = compiler.GetType(compiler.Scan())
			if Defined(subtype) {
				expression, err := compiler.Collection(T, subtype)
				if err != nil {
					return err
				}
				T = expression.Type
			} else {
				return errors.New("No such collection " + string(compiler.LastToken))
			}

		}

		compiler.TypeDefinition.Fields = append(compiler.TypeDefinition.Fields, Field{
			Name:     "",
			Embedded: true,
			Type:     T,
		})

		return nil
	}

	//If the compiler depth is zero then we can assume an implicit definition.
	if compiler.Depth == 0 {

		//Assuming type definition.
		compiler.TypeName = token
		compiler.InsideTypeDefinition = true

		compiler.Write(s("func New" + token.String() + "() " + token.String() + " {\n"))

		return compiler.CompileBlock()
	}

	return Unimplemented(token)
}
