package compiler

import "io"
import "bytes"
import "strconv"
import "errors"

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

	//Integer expression.
	if _, err := strconv.Atoi(string(token)); err == nil {
		expression.Type = Integer
		expression.Write(token)
		return expression, nil
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
