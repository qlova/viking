package compiler

import "errors"

//IndexArray indexes the array with the specified index.
func (compiler *Compiler) IndexArray(array, index Expression) (Expression, error) {
	if !index.Type.Equals(Integer) {
		return Expression{}, errors.New("Only integers can be used to index an array! Not " + index.Type.Name)
	}

	if !compiler.ScanIf(']') {
		return Expression{}, compiler.Expecting(']')
	}

	var expression = compiler.NewExpression()
	expression.Type = *array.Subtype
	expression.Go.Write(array.Go.Bytes())
	expression.Go.Write(s("["))
	expression.Go.Write(index.Go.Bytes())
	expression.Go.Write(s("%len("))
	expression.Go.Write(array.Go.Bytes())
	expression.Go.Write(s(")]"))

	return expression, nil
}

//ModifyArray scans an array modification.
func (compiler *Compiler) ModifyArray(array []byte) error {
	ArrayType := compiler.GetVariable(array)

	if !compiler.ScanIf('[') {
		return compiler.Expecting('[')
	}

	index, err := compiler.ScanExpression()
	if err != nil {
		return err
	}
	if !index.Type.Equals(Integer) {
		return errors.New("Only integers can be used to index an array! Not " + index.Type.Name)
	}

	if !compiler.ScanIf(']') {
		return compiler.Expecting(']')
	}

	if !compiler.ScanIf('=') {
		return compiler.Expecting('=')
	}

	value, err := compiler.ScanExpression()
	if err != nil {
		return err
	}
	if !value.Type.Equals(*ArrayType.Subtype) {
		return errors.New("Type mismatch! " + index.Type.Name)
	}

	compiler.Go.Write(array)
	compiler.Go.Write(s("["))
	compiler.Go.Write(index.Go.Bytes())
	compiler.Go.Write(s("%len("))
	compiler.Go.Write(array)
	compiler.Go.Write(s(")] = "))
	compiler.Go.Write(value.Go.Bytes())

	return nil
}
