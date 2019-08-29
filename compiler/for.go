package compiler

//ScanForStatement scans a for loop.
func (compiler *Compiler) ScanForStatement() error {
	if compiler.Peek().Is(":") {
		compiler.Indent()
		if Deterministic {
			compiler.Go.WriteString("for i := I.NewInteger(1); true; i = i.Add(I.NewInteger(1)) {")
		} else {
			compiler.Go.WriteString("for i := 1; true; i++ {")
		}

		compiler.GainScope()
		compiler.SetVariable(s("i"), Integer)

		return compiler.CompileBlock()
	}

	var name = compiler.Scan()
	if !compiler.Peek().Is("in") {
		expression, err := compiler.expression(name)
		if err != nil {
			return err
		}

		expression, err = compiler.Shunt(expression, 0)
		if err != nil {
			return err
		}

		if expression.Type.Is(Array) {
			if !Deterministic {
				return compiler.Unimplemented(s("non-deterministic for array"))
			}
			compiler.Import(Ilang)
			newExpression := compiler.NewExpression()
			newExpression.Go.WriteString("I.NewInteger(len(")
			newExpression.Go.Write(expression.Go.Bytes())
			newExpression.Go.WriteString("))")
			newExpression.Type = Integer
			expression = newExpression
		}

		if !expression.Type.Equals(Integer) {
			return compiler.NewError("unimplemented for loop for " + expression.Type.Name)
		}

		compiler.Indent()

		if compiler.Peek().Is("to") {
			compiler.Scan()
			var to, err = compiler.ScanExpression()
			if err != nil {
				return err
			}
			if Deterministic {
				compiler.Import(Ilang)
				compiler.Go.WriteString("for i, to := I.SetupTo(")
				compiler.Go.Write(expression.Go.Bytes())
				compiler.Go.WriteString(",")
				compiler.Go.Write(to.Go.Bytes())
				compiler.Go.WriteString("); i.Compare(to) != 0; i = i.To(to) {")
			} else {
				return compiler.Unimplemented(s("non-deterministic to"))
			}
		} else {

			compiler.Go.WriteString("for ")
			if Deterministic {
				compiler.Import(Ilang)
				compiler.Go.WriteString("i := I.NewInteger(1); i.Compare(")
				compiler.Go.Write(expression.Go.Bytes())
				compiler.Go.WriteString(") <= 0; i = i.Add(I.NewInteger(1)) {")
			} else {
				compiler.Go.WriteString("i := 1; i <= ")
				compiler.Go.Write(expression.Go.Bytes())
				compiler.Go.WriteString("; i++ {")
			}
		}

		compiler.GainScope()
		compiler.SetVariable(s("i"), Integer)

		return compiler.CompileBlock()
	}
	compiler.Scan()

	var collection, err = compiler.ScanExpression()
	if err != nil {
		return err
	}

	if collection.Is(Integer) {

		expression, err := compiler.expression(name)
		if err != nil {
			return err
		}

		expression, err = compiler.Shunt(expression, 0)
		if err != nil {
			return err
		}

		if Deterministic {
			compiler.Import(Ilang)
			compiler.Go.WriteString("for i, in, to := I.SetupStep(")
			compiler.Go.Write(collection.Go.Bytes())
			compiler.Go.WriteString(",")
			compiler.Go.Write(expression.Go.Bytes())
			compiler.Go.WriteString("); i.CompareStep(to, in); i = i.Add(in) {")
		} else {
			return compiler.Unimplemented(s("non-deterministic integer in integer for loop"))
		}

		compiler.GainScope()
		compiler.SetVariable(s("i"), Integer)
		return compiler.CompileBlock()
	}

	if !collection.Is(Variadic) && !collection.Is(List) {
		return compiler.NewError("unimplemented for loop for " + collection.Type.Name)
	}

	compiler.Indent()
	compiler.Go.WriteString("for ")
	compiler.Go.WriteString("i,")
	compiler.Go.Write(name)
	compiler.Go.WriteString(":= range ")
	compiler.Go.Write(collection.Go.Bytes())
	compiler.Go.WriteString("{")

	compiler.GainScope()
	compiler.SetVariable(name, *collection.Type.Subtype)
	compiler.SetVariable(s("i"), Integer)

	return compiler.CompileBlock()
}
