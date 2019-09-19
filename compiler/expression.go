package compiler

import (
	"fmt"
	"io"
	"os"

	"github.com/qlova/viking/compiler/target"
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
	if compiler.JS.Enabled {
		expression.JS.Enabled = true
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
	return compiler.Expression(compiler.Scan())
}

//Expression acts like ScanExpression but without the shunting.
func (compiler *Compiler) Expression(token Token) (Expression, error) {
	var expression = compiler.NewExpression()

	if token == nil {
		return Expression{}, io.EOF
	}

	//Ignore comments
	if len(token) > 2 && token[0] == '/' && token[1] == '/' {
		compiler.Go.Write(token)
		return expression, nil
	}

	//Inverse expression.
	if token.Is("-") {
		var expression, err = compiler.scanExpression()
		ok, expression, err := expression.Type.Operation(compiler, compiler.NewExpression(), expression, "-")
		if err != nil {
			return expression, err
		}
		if !ok {
			return expression, compiler.NewError("cannot invert " + expression.String(compiler))
		}
		return expression, nil
	}

	//Alias expression.
	if alias, ok := compiler.Aliases[token.String()]; ok {
		compiler.UnpackAlias(alias)
		return compiler.Expression(compiler.Scan())
	}

	//Variable expression.
	if variable := compiler.GetVariable(token); Defined(variable) {
		expression.Type = variable
		expression.Go.Write(token)

		if compiler.Peek().Is("[") {
			if collection, ok := variable.(Collection); ok {
				var args, err = compiler.Indicies()
				if err != nil {
					return expression, err
				}

				return collection.Index(compiler, expression, args...)
			}
			return expression, compiler.NewError("Unexpected [, type is not indexable")
		}

		return expression, nil
	}

	for _, T := range Types {
		if T.Expression != nil {
			ok, expression, err := T.Expression(compiler)
			if ok {
				if expression.Type == nil {
					expression.Type = T
				}
				return expression, err
			}
		}
	}

	//Sub-expression.
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

	//Length expression.
	if token.Is("#") {
		var subject, err = compiler.scanExpression()
		if err != nil {
			return Expression{}, err
		}

		if collection, ok := subject.Type.(Collection); ok {
			fmt.Println(collection.Length(compiler, subject).String(compiler))
			return collection.Length(compiler, subject), nil
		}
		return Expression{}, compiler.NewError("cannot take the length of " + subject.String(compiler))
	}

	//Function calls.
	if concept, ok := compiler.Concepts[token.String()]; ok {
		return concept.Call(compiler)
	}

	for _, builtin := range Builtins {
		if token.Is(builtin.Name()[English]) {

			if !compiler.ScanIf('(') {
				return Expression{}, compiler.NewError("expecting call to builtin")
			}

			var args, err = compiler.Arguments()
			if err != nil {
				return Expression{}, err
			}

			return builtin.Call(compiler, Expression{}, args...)
		}
	}

	if P, err := compiler.GetPackage(token); err == nil {
		return P.Expression(compiler)
	} else if !os.IsNotExist(err) {
		return Expression{}, err
	}

	return Expression{}, compiler.Undefined(token)
}
