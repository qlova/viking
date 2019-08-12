package compiler

//BasicOperation returns 'a [operation] b'
func (compiler *Compiler) BasicOperation(operation string, a, b Expression) (Expression, error) {
	var expression = compiler.NewExpression()
	expression.Type = a.Type

	switch operation {
	case "==", "&&", "||":
		expression.Type = Bit
	}

	expression.Go.Write(a.Go.Bytes())
	expression.Go.Write(s(operation))
	expression.Go.Write(b.Go.Bytes())

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

	var expression = compiler.NewExpression()
	expression.Type = Integer
	expression.Go.WriteString("div_integer(")
	expression.Go.Write(a.Go.Bytes())
	expression.Go.WriteString(",")
	expression.Go.Write(b.Go.Bytes())
	expression.Go.WriteString(")")

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
		return Expression{}, compiler.NewError("type mismatch")
	}

	compiler.Import("math")

	var expression = compiler.NewExpression()
	expression.Type = a.Type
	expression.Go.WriteString("int(math.Pow(float64(")
	expression.Go.Write(a.Go.Bytes())
	expression.Go.WriteString("),float64(")
	expression.Go.Write(b.Go.Bytes())
	expression.Go.WriteString(")))")

	return expression, nil
}
