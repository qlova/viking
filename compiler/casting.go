package compiler

func (compiler *Compiler) Cast(t Type) (Expression, error) {
	var expression Expression
	expression.Type = t

	var other, err = compiler.ScanExpression()
	if err != nil {
		return Expression{}, err
	}

	if !compiler.ScanIf(')') {
		compiler.Unexpected()
	}

	//Casting to integer.
	if t.Equals(Integer) {
		if other.Type.Equals(Symbol) {
			expression.Write([]byte("int("))
			expression.Write(other.Bytes())
			expression.Write([]byte(")"))
			return expression, nil
		}
	}

	//Casting to symbol.
	if t.Equals(Symbol) {
		if other.Type.Equals(Integer) {
			expression.Write([]byte("rune("))
			expression.Write(other.Bytes())
			expression.Write([]byte(")"))
			return expression, nil
		}
	}

	return Expression{}, Unimplemented([]byte("casting " + other.Type.Name + " to " + t.Name))
}
