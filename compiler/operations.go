package compiler

import "errors"

//BasicOperation returns 'a [operation] b'
func (compiler *Compiler) BasicOperation(operation string, a, b Expression) (Expression, error) {
	var expression Expression
	expression.Type = a.Type

	switch operation {
	case "==", "&&", "||":
		expression.Type = Bit
	}

	expression.Write(a.Bytes())
	expression.Write(s(operation))
	expression.Write(b.Bytes())

	return expression, nil
}

//BasicAnd returns a && b
func (compiler *Compiler) BasicAnd(a, b Expression) (Expression, error) {
	return compiler.BasicOperation("&&", a, b)
}

//BasicOr returns a || b
func (compiler *Compiler) BasicOr(a, b Expression) (Expression, error) {
	return compiler.BasicOperation("||", a, b)
}

//BasicEquals returns a == b
func (compiler *Compiler) BasicEquals(a, b Expression) (Expression, error) {
	return compiler.BasicOperation("==", a, b)
}

//BasicAdd returns a + b
func (compiler *Compiler) BasicAdd(a, b Expression) (Expression, error) {
	return compiler.BasicOperation("+", a, b)
}

//BasicMultiply returns a * b
func (compiler *Compiler) BasicMultiply(a, b Expression) (Expression, error) {
	return compiler.BasicOperation("*", a, b)
}

//Divide returns a / b
func (compiler *Compiler) Divide(a, b Expression) (Expression, error) {

	compiler.Require(`func div_integer(a, b int) (n int) {
	if b == 0 {
		if a == 0 {
			return 1
		}
		return 0
	}
	return a/b
}
`)

	var expression Expression
	expression.Type = Integer
	expression.WriteString("div_integer(")
	expression.Write(a.Bytes())
	expression.WriteString(",")
	expression.Write(b.Bytes())
	expression.WriteString(")")

	return expression, nil
}

//BasicSubtract returns a - b
func (compiler *Compiler) BasicSubtract(a, b Expression) (Expression, error) {
	return compiler.BasicOperation("-", a, b)
}

//Mod returns a % b
func (compiler *Compiler) Mod(a, b Expression) (Expression, error) {
	return compiler.BasicOperation("%", a, b)
}

//Pow returns a raised to power b
func (compiler *Compiler) Pow(a, b Expression) (Expression, error) {
	if !a.Type.Equals(b.Type) {
		return Expression{}, errors.New("type mismatch")
	}

	compiler.Import("math")

	var expression Expression
	expression.Type = a.Type
	expression.WriteString("int(math.Pow(float64(")
	expression.Write(a.Bytes())
	expression.WriteString("),float64(")
	expression.Write(b.Bytes())
	expression.WriteString(")))")

	return expression, nil
}
