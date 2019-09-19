package compiler

import (
	"bytes"
	"io"
	"os"

	"github.com/qlova/viking/compiler/target"
)

//Statements is a list of registered statements.
var Statements []Statement

//RegisterStatement registers a global statement and returns it.
func RegisterStatement(statement Statement) Statement {
	Statements = append(Statements, statement)
	return statement
}

//Statement is an 'i' language statement.
type Statement interface {
	Name() String
	Compile(*Compiler) error
}

//CompileStatement compiles the next statement.
func (compiler *Compiler) CompileStatement() (returning error) {
	defer func(returning *error) {
		if compiler.Throws && *returning == nil {
			if !compiler.ScanIf(';') {
				*returning = compiler.NewError("you need to handle the error")
				return
			}
			compiler.Throws = false

			switch compiler.Scan().String() {
			case "break":
				compiler.Go.WriteString("; if (len(ctx.Errors()) > 0) { break }")
			case "ignore":
				compiler.Go.WriteString("; ctx.Errors()")
			case "for":
				if !compiler.Scan().Is("errors") {
					*returning = compiler.NewError("do you mean for errors?")
					return
				}
				compiler.Go.WriteString("; for i, error := range ctx.Errors() {")
				compiler.GainScope()
				//compiler.SetVariable(s("i"), Integer)
				compiler.SetVariable(s("error"), Nothing{})
				*returning = compiler.CompileBlock()
				return
			default:
				*returning = compiler.NewError("unsupported tag " + compiler.LastToken.String())
				return
			}
		}
	}(&returning)

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

	case ";":
		return compiler.NewError("statement doesn't throw error")

	case "|":
		if !compiler.Flag(s("if")) {
			return compiler.NewError("| requires a preceding if statement")
		}
		compiler.LoseScope()
		compiler.Indent()
		compiler.Go.WriteString("} else {")
		compiler.GainScope()
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

		compiler.LoseScope()
		compiler.Go.Write(s("}"))
		compiler.JS.Write(s("}"))

		if compiler.Main && compiler.Depth == 0 {
			compiler.JS.Write(s("()"))
		}

		return nil
	}

	for _, statement := range Statements {
		if statement.Name()[English] == token.String() {
			return statement.Compile(compiler)
		}
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

	for _, builtin := range Builtins {
		if token.Is(builtin.Name()[English]) {

			if !compiler.ScanIf('(') {
				return compiler.NewError("expecting call to builtin")
			}

			var args, err = compiler.Arguments()
			if err != nil {
				return err
			}

			return builtin.Run(compiler, Expression{}, args...)
		}
	}

	if T := compiler.GetVariable(token); Defined(T) {

		var expression = compiler.NewExpression()
		expression.Type = T
		expression.Go.Write(token)

		if runnable, ok := T.(Runnable); ok && compiler.Peek().Is("(") {

			compiler.Scan()

			var args, err = compiler.Arguments()
			if err != nil {
				return err
			}

			return runnable.Run(compiler, expression, args...)
		}
	}

	//Aliases.
	if compiler.ScanIf('=') {
		compiler.DefineAlias(token)
		return nil
	}

	//Collections.
	if compiler.Peek().Is("[") {

		variable := compiler.GetVariable(token)
		if !Defined(variable) {
			return compiler.Undefined(token)
		}

		collection, ok := variable.(Collection)
		if !ok {
			return compiler.NewError("cannot index " + token.String() + ", not a collection type")
		}

		var indicies, err = compiler.Indicies()
		if err != nil {
			return err
		}

		var expression = compiler.NewExpression()
		expression.Type = variable
		expression.Go.Write(token)

		if !compiler.ScanIf('$') {
			return compiler.Expecting('$')
		}

		if !compiler.ScanIf('=') {
			return compiler.Expecting('=')
		}

		modification, err := compiler.ScanExpression()
		if err != nil {
			return err
		}

		return collection.Modify(compiler, expression, modification, indicies...)
	}

	//Variables.
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
	/*if T := compiler.GetVariable(token); Defined(T) && T.Is(Function) && compiler.Peek().Is("(") {
		return compiler.CallFunction(token)
	}*/

	//Concept calls.
	if concept, ok := compiler.Concepts[token.String()]; ok {
		return concept.Run(compiler)
	}

	//Embedded types.
	/*if T := compiler.GetType(token); Defined(T) {
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
	}*/

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
