package compiler

import "io"
import "bytes"
import "strconv"
import "errors"

type Expression struct {
	Type
	bytes.Buffer
}

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
	
	if token[0] == '"' {
		expression.Type = String
		expression.Write(token)
		return expression, nil
	}
	
	if token[0] == '\'' {
		expression.Type = Symbol
		expression.Write(token)
		return expression, nil
	}
	
	if _, err := strconv.Atoi(string(token)); err == nil {
		expression.Type = Integer
		expression.Write(token)
		return expression, nil
	}
	
	if variable := compiler.GetVariable(token); Defined(variable) {
		expression.Type = variable
		expression.Write(token)
		return expression, nil
	}
	
	if t := compiler.GetType(token); Defined(t) {
		var next = compiler.Scan()
		
		if equal(next, "(") {
			if compiler.ScanIf(')') {
				return compiler.Type(t)
			} else {
				return compiler.Cast(t)
			}
		}
		
		if equal(next, ".") {
			var collection = compiler.GetType(compiler.Scan())
			if Defined(collection) {
				return compiler.Collection(collection, t)
			} else {
				return Expression{}, errors.New("No such collection "+string(compiler.LastToken))
			}
		}
		
		return Expression{}, Unimplemented(append(append(token, next...), compiler.Peek()...))
	}

	return Expression{}, Unimplemented(token)
}

