package compiler

//Cast from expression to Type 'to'.
func (compiler *Compiler) Cast(from Expression, to Type) (Expression, error) {
	if from.Equals(to) {
		return from, nil
	}

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
		if from.Type.Equals(String) {
			expression.WriteString("strconv_atoi(")
			expression.Write(from.Bytes())
			expression.WriteString(")")

			compiler.Import("strconv")
			compiler.Require(`func strconv_atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
`)

			return expression, nil
		}
	}

	//Casting to bit.
	if to.Equals(Bit) {
		if from.Type.Equals(Integer) {
			expression.Write([]byte("("))
			expression.Write(from.Bytes())
			expression.Write([]byte(" != 0)"))
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

	//Casting to String.
	if to.Equals(String) {
		if from.Type.Equals(Symbol) {
			expression.Write([]byte("string("))
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
