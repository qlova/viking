package compiler

import (
	"bytes"
	"errors"
	"io"
)

//CompileStatement compiles the next statement.
func (compiler *Compiler) CompileStatement() error {
	var token = compiler.Scan()
	/*switch token {
		case "if", "for", "return", "break", "go":
			return Unimplemented
	}*/

	if token == nil {
		return io.EOF
	}

	//Ignore newlines.
	if bytes.Equal(token, []byte("\n")) {
		return nil
	}

	//Comments.
	if len(token) > 2 && token[0] == '/' && token[1] == '/' {
		compiler.Write(token)

		//Special output comment for tests.
		if len(token) > len("//output: ") && bytes.Equal(token[:len("//output: ")], []byte("//output: ")) {
			compiler.ExpectedOutput = token[len("//output: "):]
			compiler.ExpectedOutput = bytes.Replace(compiler.ExpectedOutput, []byte(`\n`), []byte("\n"), -1)
		}

		//Special output comment for tests.
		if len(token) > len("//input: ") && bytes.Equal(token[:len("//input: ")], []byte("//input: ")) {
			compiler.ProvidedInput = token[len("//input: "):]
			compiler.ProvidedInput = bytes.Replace(compiler.ProvidedInput, []byte(`\n`), []byte("\n"), -1)
		}

		return nil
	}

	switch token.String() {

	//Import statement.
	case "import":
		var namespace = compiler.Scan()
		compiler.NewContext()
		compiler.CompileFile(string(namespace) + ".i")
		return nil

	//Export tag.
	case ".":
		compiler.Export = true
		defer func() {
			compiler.Export = false
		}()
		return compiler.CompileStatement()

	//Main statement.
	case "main":
		compiler.Write(s("func main() {"))

		compiler.GainScope()
		return compiler.CompileBlock()

	case "if":
		var condition, err = compiler.ScanExpression()
		if err != nil {
			return err
		}

		if !condition.Equals(Bit) {
			condition, err = compiler.Cast(condition, Bit)
			if err != nil {
				return err
			}
		}
		compiler.Indent()
		compiler.WriteString("if ")
		compiler.Write(condition.Bytes())
		compiler.WriteString(" {")

		compiler.GainScope()
		return compiler.CompileBlock()

	case "|":
		compiler.LoseScope()
		compiler.Indent()
		compiler.WriteString("} else {")
		compiler.GainScope()
		return compiler.CompileBlock()

	case "for":
		var name = compiler.Scan()

		if !compiler.Scan().Is("in") {
			return errors.New("expecting in")
		}

		var collection, err = compiler.ScanExpression()
		if err != nil {
			return err
		}

		if !collection.Is(Variadic) && !collection.Is(List) {
			return errors.New("unimplemented for loop for " + collection.Type.Name)
		}

		compiler.Indent()
		compiler.WriteString("for ")
		compiler.WriteString("_,")
		compiler.Write(name)
		compiler.WriteString(":= range ")
		compiler.Write(collection.Bytes())
		compiler.WriteString("{")

		compiler.GainScope()
		compiler.SetVariable(name, *collection.Type.Subtype)

		return compiler.CompileBlock()

	//Return statement.
	case "return":
		compiler.Indent()
		compiler.WriteString("return ")

		if compiler.Peek().Is("\n") {
			return nil
		}

		expression, err := compiler.ScanExpression()
		if err != nil {
			return err
		}
		*compiler.Returns = expression.Type

		compiler.Write(expression.Bytes())
		return nil

	//Close block.
	case "}":
		if compiler.Depth == 0 {
			return compiler.Unexpected()
		}

		compiler.Depth--
		compiler.Indent()
		compiler.Depth++
		compiler.Write(s("}"))
		compiler.LoseScope()
		return nil

	}

	//Is this a builtin call?
	if Builtin(token) {
		compiler.Indent()
		return compiler.CompileBuiltin(token)
	}

	//Array modification.
	if compiler.Peek().Is("[") {
		if Defined(compiler.GetVariable(token)) {
			compiler.Indent()
			return compiler.ModifyArray(token)
		}
	}

	//Variable modification.
	if compiler.ScanIf('$') {
		if compiler.ScanIf('=') {
			compiler.Indent()
			if Defined(compiler.GetVariable(token)) {
				return compiler.AssignVariable(token)
			}
			return compiler.DefineVariable(token)
		}
		return compiler.Unexpected()
	}

	//Function calls.
	if T := compiler.GetVariable(token); Defined(T) && T.Is(Function) && compiler.Peek().Is("(") {
		return compiler.CallFunction(token)
	}

	//Concept calls.
	if _, ok := compiler.Concepts[token.String()]; ok {
		return compiler.RunConcept(token)
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

		if compiler.Concepts == nil {
			compiler.Concepts = make(map[string]Concept)
		}

		//Function definition?
		if compiler.ScanIf('(') {

			//Concept with multiple arguments.
			if !compiler.ScanIf(')') {

				var arguments, err = compiler.ScanArguments()
				if err != nil {
					return err
				}

				var cache = compiler.CacheBlock()

				compiler.Concepts[token.String()] = Concept{
					Cache:     cache,
					Name:      token,
					Arguments: arguments,
				}

				return nil
			}

			var cache = compiler.CacheBlock()

			compiler.Concepts[token.String()] = Concept{
				Cache: cache,
				Name:  token,
			}

			return nil
		}

		//Assuming type definition.
		compiler.TypeName = token
		compiler.InsideTypeDefinition = true

		compiler.Write(s("func New" + token.String() + "() " + token.String() + " {\n"))

		compiler.GainScope()
		return compiler.CompileBlock()
	}

	return Unimplemented(s("statement" + token.String()))
}
