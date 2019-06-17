package compiler

import "errors"

func (compiler *Compiler) IndexArray(array, index Expression) (Expression, error) {
	if !index.Type.Equals(Integer) {
		return Expression{}, errors.New("Only integers can be used to index an array! Not "+index.Type.Name)
	}
	
	if !compiler.ScanIf(']') {
		return Expression{}, compiler.Expecting(']')
	}
	
	var expression Expression
	expression.Type = *array.Subtype
	expression.Write(array.Bytes())
	expression.Write(s("["))
	expression.Write(index.Bytes())
	expression.Write(s("%len("))
	expression.Write(array.Bytes())
	expression.Write(s(")]"))
	
	return expression, nil
}

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
		return errors.New("Only integers can be used to index an array! Not "+index.Type.Name)
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
		return errors.New("Type mismatch! "+index.Type.Name)
	}
	
	

	compiler.Write(array)
	compiler.Write(s("["))
	compiler.Write(index.Bytes())
	compiler.Write(s("%len("))
	compiler.Write(array)
	compiler.Write(s(")] = "))
	compiler.Write(value.Bytes())
	
	return nil
}
