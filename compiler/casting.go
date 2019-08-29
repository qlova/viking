package compiler

//Cast from expression to Type 'to'.
func (compiler *Compiler) Cast(from Expression, to Type) (Expression, error) {
	if from.Equals(to) {
		return from, nil
	}

	var expression = compiler.NewExpression()
	expression.Type = to

	//Casting to integer.
	if to.Equals(Integer) {
		if from.Type.Equals(Symbol) {
			expression.Go.Write([]byte("I.NewInteger(int("))
			expression.Go.Write(from.Go.Bytes())
			expression.Go.Write([]byte("))"))
			return expression, nil
		}
		if from.Type.Equals(String) {

			compiler.Import(Ilang)
			compiler.Throws = true

			if Deterministic {
				expression.Go.WriteString("I.Atoi(ctx, ")
				expression.Go.Write(from.Go.Bytes())
				expression.Go.WriteString(")")
				return expression, nil
			}

			expression.Go.WriteString("strconv_atoi(ctx,")
			expression.Go.Write(from.Go.Bytes())
			expression.Go.WriteString(")")

			compiler.Import("strconv")
			compiler.Require(`func strconv_atoi(ctx I.Context, s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		ctx.Throw(1, "invalid integer")
	}
	return i
}
`)

			return expression, nil
		}
	}

	//Casting to bit.
	if to.Equals(Bit) {
		if from.Type.Equals(Integer) {
			if Deterministic {
				compiler.Import(Ilang)
				expression.Go.Write([]byte("(!"))
				expression.Go.Write(from.Go.Bytes())
				expression.Go.Write([]byte(".Equals(I.Integer{}))"))
			} else {
				expression.Go.Write([]byte("("))
				expression.Go.Write(from.Go.Bytes())
				expression.Go.Write([]byte("!= 0)"))
			}
			return expression, nil
		}
	}

	//Casting to symbol.
	if to.Equals(Symbol) {
		if from.Type.Equals(Integer) {
			expression.Go.Write([]byte("rune("))
			expression.Go.Write(from.Go.Bytes())
			expression.Go.Write([]byte(".Int64())"))
			return expression, nil
		}
	}

	//Casting to String.
	if to.Equals(String) {
		if from.Type.Equals(Symbol) {
			expression.Go.Write([]byte("string("))
			expression.Go.Write(from.Go.Bytes())
			expression.Go.Write([]byte(")"))
			return expression, nil
		}
	}

	return Expression{}, compiler.Unimplemented([]byte("casting " + from.Type.Name + " to " + to.Name))
}

//ScanCast scans and compiles a cast.
func (compiler *Compiler) ScanCast(t Type) (Expression, error) {
	var other, err = compiler.ScanExpression()
	if err != nil {
		return Expression{}, err
	}

	expression, err := compiler.Cast(other, t)
	if err != nil {
		return expression, err
	}

	if !compiler.ScanIf(')') {
		compiler.Expecting(')')
	}

	return expression, nil
}
