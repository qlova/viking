package compiler

//ConcatArray joins two arrays.
/*func (compiler *Compiler) ConcatArray(a, b Expression) (Expression, error) {
	if !b.Is(Array) && a.Equals(b.Type) {
		return Expression{}, compiler.NewError("cannot add array and " + b.Type.Name)
	}

	var result = compiler.NewExpression()
	result.Type = a.Type
	result.Size = a.Type.Size + b.Type.Size

	result.Go.Write(GoTypeOf(result.Type))
	result.Go.WriteString(`{`)
	result.Go.Write(a.Go.Bytes())
	result.Go.WriteString(`[0]`)
	for i := 1; i < a.Size; i++ {
		result.Go.WriteString(`,`)
		result.Go.Write(a.Go.Bytes())
		result.Go.WriteString(`[` + strconv.Itoa(i) + `]`)
	}
	for i := 0; i < b.Size; i++ {
		result.Go.WriteString(`,`)
		result.Go.Write(b.Go.Bytes())
		result.Go.WriteString(`[` + strconv.Itoa(i) + `]`)
	}
	result.Go.WriteString(`}`)

	return result, nil
}

//IndexArray indexes the array with the specified index.
func (compiler *Compiler) IndexArray(array, index Expression) (Expression, error) {
	if !index.Type.Equals(Integer) {
		return Expression{}, compiler.NewError("Only integers can be used to index an array! Not " + index.Type.Name)
	}

	if !compiler.ScanIf(']') {
		return Expression{}, compiler.Expecting(']')
	}

	var expression = compiler.NewExpression()
	expression.Type = *array.Subtype
	expression.Go.Write(array.Go.Bytes())
	expression.Go.Write(s("["))
	if Deterministic {
		compiler.Import(Ilang)
		expression.Go.Write(s("I.IndexArray("))
		expression.Go.Write(index.Go.Bytes())
		expression.Go.Write(s(", len("))
		expression.Go.Write(array.Go.Bytes())
		expression.Go.Write(s("))]"))
	} else {
		expression.Go.Write(index.Go.Bytes())
		expression.Go.Write(s("]"))
	}

	return expression, nil
}

//ModifyCollection scans a collection modification.
func (compiler *Compiler) ModifyCollection(array Token) error {
	CollectionType := compiler.GetVariable(array)

	if CollectionType.Is(Array) {
		return compiler.ModifyArray(array)
	}

	if CollectionType.Is(List) {
		return compiler.ModifyList(array)
	}

	return compiler.NewError("cannot modify a collection of type " + CollectionType.Name)
}

//ModifyList scans an array modification.
func (compiler *Compiler) ModifyList(array []byte) error {
	ArrayType := compiler.GetVariable(array)

	if compiler.ScanIf('+') {
		if !compiler.ScanIf(']') {
			return compiler.Expecting(']')
		}

		if !compiler.ScanIf('$') {
			return compiler.Expecting('$')
		}

		if !compiler.ScanIf('=') {
			return compiler.Expecting('=')
		}

		value, err := compiler.ScanExpression()
		if err != nil {
			return err
		}
		if !value.Type.Equals(*ArrayType.Subtype) {
			return compiler.NewError("Type mismatch! " + value.Type.Name)
		}

		compiler.Go.Write(array)
		compiler.Go.Write(s(" = append("))
		compiler.Go.Write(array)
		compiler.Go.Write(s(","))
		compiler.Go.Write(value.Go.Bytes())
		compiler.Go.Write(s(")"))
		return nil
	}

	index, err := compiler.ScanExpression()
	if err != nil {
		return err
	}
	if !index.Type.Equals(Integer) {
		return compiler.NewError("Only integers can be used to index a list! Not " + index.Type.Name)
	}

	if !compiler.ScanIf(']') {
		return compiler.Expecting(']')
	}

	if !compiler.ScanIf('$') {
		return compiler.Expecting('$')
	}

	if !compiler.ScanIf('=') {
		return compiler.Expecting('=')
	}

	value, err := compiler.ScanExpression()
	if err != nil {
		return err
	}
	if !value.Type.Equals(*ArrayType.Subtype) {
		return compiler.NewError("Type mismatch! " + index.Type.Name)
	}

	compiler.Go.Write(array)
	compiler.Go.Write(s("["))

	if Deterministic {
		compiler.Go.Write(s("I.IndexArray("))
		compiler.Go.Write(index.Go.Bytes())
		compiler.Go.Write(s(", len("))
		compiler.Go.Write(array)
		compiler.Go.Write(s("))] = "))
	} else {
		compiler.Go.Write(index.Go.Bytes())
		compiler.Go.Write(s("] = "))
	}
	compiler.Go.Write(value.Go.Bytes())

	return nil
}

//ModifyArray scans an array modification.
func (compiler *Compiler) ModifyArray(array []byte) error {
	ArrayType := compiler.GetVariable(array)

	index, err := compiler.ScanExpression()
	if err != nil {
		return err
	}
	if !index.Type.Equals(Integer) {
		return compiler.NewError("Only integers can be used to index an array! Not " + index.Type.Name)
	}

	if !compiler.ScanIf(']') {
		return compiler.Expecting(']')
	}

	if !compiler.ScanIf('$') {
		return compiler.Expecting('$')
	}

	if !compiler.ScanIf('=') {
		return compiler.Expecting('=')
	}

	value, err := compiler.ScanExpression()
	if err != nil {
		return err
	}
	if !value.Type.Equals(*ArrayType.Subtype) {
		return compiler.NewError("Type mismatch! " + index.Type.Name)
	}

	compiler.Go.Write(array)
	compiler.Go.Write(s("["))

	if Deterministic {
		compiler.Go.Write(s("I.IndexArray("))
		compiler.Go.Write(index.Go.Bytes())
		compiler.Go.Write(s(", len("))
		compiler.Go.Write(array)
		compiler.Go.Write(s("))] = "))
	} else {
		compiler.Go.Write(index.Go.Bytes())
		compiler.Go.Write(s("] = "))
	}
	compiler.Go.Write(value.Go.Bytes())

	return nil
}
*/
