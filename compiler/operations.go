package compiler

import "errors"

func (compiler *Compiler) BasicAdd(a, b Expression) (Expression, error) {
	if !a.Type.Equals(b.Type) {
		return Expression{}, errors.New("Type mismatch")
	}

	var expression Expression
	expression.Type = a.Type
	expression.Write(a.Bytes())
	expression.Write(s("+"))
	expression.Write(b.Bytes())

	return expression, nil
}
