package compiler

import "errors"

//BasicAdd returns a + b
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

//BasicMultiply returns a * b
func (compiler *Compiler) BasicMultiply(a, b Expression) (Expression, error) {
	if !a.Type.Equals(b.Type) {
		return Expression{}, errors.New("Type mismatch")
	}

	var expression Expression
	expression.Type = a.Type
	expression.Write(a.Bytes())
	expression.Write(s("*"))
	expression.Write(b.Bytes())

	return expression, nil
}

//BasicDivide returns a - b
func (compiler *Compiler) BasicDivide(a, b Expression) (Expression, error) {
	if !a.Type.Equals(b.Type) {
		return Expression{}, errors.New("Type mismatch")
	}

	var expression Expression
	expression.Type = a.Type
	expression.Write(a.Bytes())
	expression.Write(s("/"))
	expression.Write(b.Bytes())

	return expression, nil
}

//BasicSubtract returns a - b
func (compiler *Compiler) BasicSubtract(a, b Expression) (Expression, error) {
	if !a.Type.Equals(b.Type) {
		return Expression{}, errors.New("Type mismatch")
	}

	var expression Expression
	expression.Type = a.Type
	expression.Write(a.Bytes())
	expression.Write(s("-"))
	expression.Write(b.Bytes())

	return expression, nil
}
