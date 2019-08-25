package compiler

import (
	"bytes"
	"io"
	"os"
	"viking/compiler/target"
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
		compiler.Go.Write(token)

		//Special output comment for tests.
		if len(token) > len("//output: ") && bytes.Equal(token[:len("//output: ")], []byte("//output: ")) {
			compiler.ExpectedOutput = token[len("//output: "):]
			compiler.ExpectedOutput = bytes.Replace(compiler.ExpectedOutput, []byte(`\n`), []byte("\n"), -1)
			compiler.ExpectedOutput = bytes.Replace(compiler.ExpectedOutput, []byte(`$HOME`), []byte(os.Getenv("HOME")), -1)
			compiler.ExpectedOutput = bytes.Replace(compiler.ExpectedOutput, []byte(`$USER`), []byte(os.Getenv("USER")), -1)
			compiler.ExpectedOutput = bytes.Replace(compiler.ExpectedOutput, []byte(`$PATH`), []byte(os.Getenv("PATH")), -1)
		}

		//Special output comment for tests.
		if len(token) > len("//input: ") && bytes.Equal(token[:len("//input: ")], []byte("//input: ")) {
			compiler.ProvidedInput = token[len("//input: "):]
			compiler.ProvidedInput = bytes.Replace(compiler.ProvidedInput, []byte(`\n`), []byte("\n"), -1)
		}

		return nil
	}

	switch token.String() {

	//Export tag.
	case ".":
		compiler.Export = true
		defer func() {
			compiler.Export = false
		}()
		return compiler.CompileStatement()

	//Main statement.
	case "main":
		compiler.Require(`
type Error struct {
	Code int
	Message string
}
		
type Context struct {
	Error
}
`)
		compiler.Go.WriteString("func main() {\n")
		compiler.GainScope()
		compiler.Indent()
		compiler.Go.WriteString(`var ctx = new(Context)` + "\n")
		return compiler.CompileBlock()

	case "if":
		return compiler.ScanIfStatement()

	case "|":
		compiler.LoseScope()
		compiler.Indent()
		compiler.Go.WriteString("} else {")
		compiler.GainScope()
		return compiler.CompileBlock()

	case "for":
		var name = compiler.Scan()

		if !compiler.Scan().Is("in") {
			return compiler.NewError("expecting in")
		}

		var collection, err = compiler.ScanExpression()
		if err != nil {
			return err
		}

		if !collection.Is(Variadic) && !collection.Is(List) {
			return compiler.NewError("unimplemented for loop for " + collection.Type.Name)
		}

		compiler.Indent()
		compiler.Go.WriteString("for ")
		compiler.Go.WriteString("_,")
		compiler.Go.Write(name)
		compiler.Go.WriteString(":= range ")
		compiler.Go.Write(collection.Go.Bytes())
		compiler.Go.WriteString("{")

		compiler.GainScope()
		compiler.SetVariable(name, *collection.Type.Subtype)

		return compiler.CompileBlock()

	//Return statement.
	case "return":
		compiler.Indent()
		compiler.Go.WriteString("return ")

		if compiler.Peek().Is("\n") {
			return nil
		}

		expression, err := compiler.ScanExpression()
		if err != nil {
			return err
		}

		*compiler.Returns = expression.Type

		compiler.Go.Write(expression.Go.Bytes())
		return nil

	//Close block.
	case "}":
		if compiler.Depth == 0 {
			return compiler.NewError("closing block but there are no blocks")
		}

		compiler.Depth--
		compiler.Indent()
		compiler.Depth++
		compiler.Go.Write(s("}"))
		compiler.LoseScope()
		return nil

	case "catch":
		compiler.Require(`
func (ctx *Context) Catch() Error {
	defer func() {
		ctx.Error.Code = 0
		ctx.Error.Message = ""
	}()
	return ctx.Error
}

`)

		if err := compiler.CompileStatement(); err != nil {
			return err
		}
		compiler.Go.WriteString("\n")
		compiler.Indent()
		compiler.Go.WriteString("if err := ctx.Catch(); err.Code > 0 {")
		compiler.GainScope()
		return compiler.CompileBlock()
	}

	//Inline target code.
	if target := target.FromString(token.String()); target.Valid() {
		if inline := compiler.Peek(); inline[0] == '`' {
			compiler.Scan()
			if compiler.ScanIf(';') {
				var mode = compiler.Scan()
				switch s := mode.String(); s {
				case "head":
					compiler.Get(target).Head.Write(inline[1 : len(inline)-1])
				default:
					return compiler.NewError("unsupported tag " + s)
				}
				compiler.Get(target).Head.WriteByte('\n')
			} else {
				compiler.Get(target).Write(inline[1 : len(inline)-1])
				compiler.Get(target).WriteByte('\n')
			}
			return nil
		}
		return compiler.NewError("expecting `[inline code]`")
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
		return compiler.NewError("$ must be followed by =")
	}

	//Function calls.
	if T := compiler.GetVariable(token); Defined(T) && T.Is(Function) && compiler.Peek().Is("(") {
		return compiler.CallFunction(token)
	}

	//Concept calls.
	if concept, ok := compiler.Concepts[token.String()]; ok {
		return concept.Run(compiler)
	}

	//Embedded types.
	if T := compiler.GetType(token); Defined(T) {
		if !compiler.InsideTypeDefinition {
			return compiler.NewError("Cannnot embed type here, are you in a type definition?")
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
				return compiler.NewError("No such collection " + string(compiler.LastToken))
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

		compiler.Go.Write(s("func New" + token.String() + "() " + token.String() + " {\n"))

		compiler.GainScope()
		return compiler.CompileBlock()
	}

	return compiler.Undefined(s("statement: " + token.String()))
}
