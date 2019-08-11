package compiler

import (
	"bytes"
	"errors"
	"io"
	"strconv"
)

//Expression is a type with content.
type Expression struct {
	Type
	bytes.Buffer
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
	var expression Expression

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
		compiler.Write(token)
		return expression, nil
	}

	//String expression.
	if token[0] == '"' {
		expression.Type = String
		expression.Write(token)
		return expression, nil
	}

	//Symbol expresion.
	if token[0] == '\'' {
		expression.Type = Symbol
		expression.Write(token)
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
		expression.Write(token)
		expression.Write(internal.Bytes())
		expression.WriteString(")")
		return expression, nil
	}

	//Integer expression.
	if _, err := strconv.Atoi(string(token)); err == nil {
		expression.Type = Integer
		expression.Write(token)
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

		expression.Write(GoTypeOf(expression.Type))
		expression.WriteByte('{')
		expression.Write(first.Bytes())

		for compiler.ScanIf(',') {
			expression.WriteByte(',')

			var item, err = compiler.ScanExpression()
			if err != nil {
				return Expression{}, err
			}
			expression.Write(item.Bytes())
		}

		if !compiler.ScanIf(']') {
			return Expression{}, compiler.Expecting(']')
		}
		expression.WriteByte('}')

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
			expression.WriteString("len(")
			expression.Write(collection.Bytes())
			expression.WriteString(")")

			return expression, nil
		}

		return Expression{}, errors.New("cannot take the length of " + collection.Type.Name)
	}

	//Variable expression.
	if variable := compiler.GetVariable(token); Defined(variable) {
		expression.Type = variable
		expression.Write(token)
		return expression, nil
	}

	//Function calls.
	if concept, ok := compiler.Concepts[token.String()]; ok {
		if len(concept.Arguments) == 0 {
			expression.Type = Function
			expression.Write(token)
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
			return Expression{}, errors.New("No such collection " + string(compiler.LastToken))
		}

		return Expression{}, Unimplemented(append(append(token, next...), compiler.Peek()...))
	}

	return Expression{}, Unimplemented(token)
}
