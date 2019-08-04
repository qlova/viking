package compiler

import (
	"errors"
)

//Defined returns true if T is defined.
func Defined(T Type) bool {
	return T.Name != ""
}

//SetVariable sets a new variable.
func (compiler *Compiler) SetVariable(name []byte, T Type) {
	compiler.Scope[len(compiler.Scope)-1].Table[string(name)] = T
}

//GetVariable returns the variable with the given name.
func (compiler *Compiler) GetVariable(name []byte) Type {
	if len(compiler.Scope) <= 0 {
		return Type{}
	}
	return compiler.Scope[len(compiler.Scope)-1].Table[string(name)]
}

//DefineVariable defines the variable 'name' with the scanned value.
func (compiler *Compiler) DefineVariable(name []byte) error {
	var expression, err = compiler.ScanExpression()
	if err != nil {
		return err
	}

	compiler.SetVariable(name, expression.Type)
	compiler.Write([]byte("var "))
	compiler.Write(name)
	compiler.Write([]byte(" = "))
	compiler.Write(expression.Bytes())

	if !compiler.ScanIf('\n') {
		return compiler.Unexpected()
	}

	return nil
}

//AssignVariable modifies the variable 'name' with the scanned value.
func (compiler *Compiler) AssignVariable(name []byte) error {
	var variable = compiler.GetVariable(name)

	var expression, err = compiler.ScanExpression()
	if err != nil {
		return err
	}
	if !expression.Type.Equals(variable) {
		return errors.New("type mismatch")
	}

	compiler.SetVariable(name, expression.Type)
	compiler.Write(name)
	compiler.Write([]byte(" = "))
	compiler.Write(expression.Bytes())

	if !compiler.ScanIf('\n') {
		return compiler.Unexpected()
	}

	return nil
}
