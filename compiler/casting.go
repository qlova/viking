package compiler

//Cast from expression to Type 'to'.
func (compiler *Compiler) Cast(from Expression, to Type) (Expression, error) {
	var expression Expression
	expression.Type = to

	//Casting to integer.
	if to.Equals(Integer) {
		if from.Type.Equals(Symbol) {
			expression.Write([]byte("int("))
			expression.Write(from.Bytes())
			expression.Write([]byte(")"))
			return expression, nil
		}
	}

	//Casting to symbol.
	if to.Equals(Symbol) {
		if from.Type.Equals(Integer) {
			expression.Write([]byte("rune("))
			expression.Write(from.Bytes())
			expression.Write([]byte(")"))
			return expression, nil
		}
	}

	return Expression{}, Unimplemented([]byte("casting " + from.Type.Name + " to " + to.Name))
}

//ScanCast scans and compiles a cast.
func (compiler *Compiler) ScanCast(t Type) (Expression, error) {
	var other, err = compiler.ScanExpression()
	if err != nil {
		return Expression{}, err
	}

	if !compiler.ScanIf(')') {
		compiler.Unexpected()
	}

	return compiler.Cast(other, t)
}
