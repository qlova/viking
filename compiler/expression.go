package compiler

import (
	"io"
	"strconv"
	"viking/compiler/target"
)

//Expression is a type with content.
type Expression struct {
	Type
	target.Buffer
}

func (compiler *Compiler) NewExpression() Expression {
	var expression Expression
	if compiler.Go.Enabled {
		expression.Go.Enabled = true
	}
	return expression
}

//ScanExpression returns the next expression or an error.
func (compiler *Compiler) ScanExpression() (Expression, error) {
	var expression, err = compiler.scanExpression()
	if err != nil {
		return expression, err
	}

	return compiler.Shunt(expression, 0)
}

func (compiler *Compiler) scanExpression() (Expression, error) {
	var expression = compiler.NewExpression()

	var token = compiler.Scan()

	/*switch token {
		case "if", "for", "return", "break", "go":
			return Unimplemented
	}*/

	if token == nil {
		return Expression{}, io.EOF
	}

	//Ignore comments
	if len(token) > 2 && token[0] == '/' && token[1] == '/' {
		compiler.Go.Write(token)
		return expression, nil
	}

	//String expression.
	if token[0] == '"' {
		expression.Type = String
		expression.Go.Write(token)
		return expression, nil
	}

	//Symbol expresion.
	if token[0] == '\'' {
		expression.Type = Symbol
		expression.Go.Write(token)
		return expression, nil
	}

	//String expression.
	if token.Is("(") {
		var internal, err = compiler.ScanExpression()
		if err != nil {
			return internal, err
		}
		if !compiler.ScanIf(')') {
			return internal, compiler.Expecting(')')
		}
		expression.Type = internal.Type
		expression.Go.Write(token)
		expression.Go.Write(internal.Go.Bytes())
		expression.Go.WriteString(")")
		return expression, nil
	}

	//Integer expression.
	if _, err := strconv.Atoi(string(token)); err == nil {
		expression.Type = Integer
		expression.Go.Write(token)
		return expression, nil
	}

	//List expression.
	if token.Is("[") {
		expression.Type = List

		var first, err = compiler.ScanExpression()
		if err != nil {
			return Expression{}, err
		}

		expression.Type.Subtype = &first.Type

		expression.Go.Write(GoTypeOf(expression.Type))
		expression.Go.WriteByte('{')
		expression.Go.Write(first.Go.Bytes())

		for compiler.ScanIf(',') {
			expression.Go.WriteByte(',')

			var item, err = compiler.ScanExpression()
			if err != nil {
				return Expression{}, err
			}
			expression.Go.Write(item.Go.Bytes())
		}

		if !compiler.ScanIf(']') {
			return Expression{}, compiler.Expecting(']')
		}
		expression.Go.WriteByte('}')

		return expression, nil
	}

	//Length expression.
	if token.Is("#") {
		expression.Type = Integer

		var collection, err = compiler.scanExpression()
		if err != nil {
			return Expression{}, err
		}

		if collection.Is(List) || collection.Is(String) {
			expression.Go.WriteString("len(")
			expression.Go.Write(collection.Go.Bytes())
			expression.Go.WriteString(")")

			return expression, nil
		}

		return Expression{}, compiler.NewError("cannot take the length of " + collection.Type.Name)
	}

	//Variable expression.
	if variable := compiler.GetVariable(token); Defined(variable) {
		expression.Type = variable
		expression.Go.Write(token)
		return expression, nil
	}

	//Function calls.
	if concept, ok := compiler.Concepts[token.String()]; ok {
		if len(concept.Arguments) == 0 {
			expression.Type = Function
			expression.Go.Write(token)
			return expression, nil
		}

		return compiler.CallConcept(token)
	}

	//Is this a builtin call?
	if Builtin(token) {
		return compiler.CallBuiltin(token)
	}

	//Prototype conversion.
	if T := compiler.GetPrototype(token); T.Defined() {
		if T.ScanExpression != nil {
			return T.ScanExpression(compiler)
		}
	}

	//Collections, arrays, lists etc.
	if T := compiler.GetType(token); Defined(T) {
		var next = compiler.Scan()

		if next.Is("(") {
			if compiler.ScanIf(')') {
				return compiler.Type(T)
			}
			return compiler.ScanCast(T)
		}

		if next.Is(".") {
			var collection = T

			var subtype = compiler.GetType(compiler.Scan())

			if Defined(subtype) {
				return compiler.Collection(collection, subtype)
			}
			return Expression{}, compiler.NewError("No such collection " + string(compiler.LastToken))
		}

		return Expression{}, compiler.Unimplemented(append(append(token, next...), compiler.Peek()...))
	}

	return Expression{}, compiler.Undefined(token)
}
