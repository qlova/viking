package compiler

import "errors"

func Defined(T Type) bool {
	return T.Name != ""
}

func (compiler *Compiler) SetVariable(name []byte, T Type) {
	compiler.Scope[len(compiler.Scope)-1][string(name)] = T
}

func (compiler *Compiler) GetVariable(name []byte) Type {
	return compiler.Scope[len(compiler.Scope)-1][string(name)]
}

func (compiler *Compiler) DefineVariable(name []byte) error {
	if !compiler.ScanIf('=') {
		compiler.Unexpected()
	}

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

func (compiler *Compiler) AssignVariable(name []byte) error {
	if !compiler.ScanIf('=') {
		compiler.Unexpected()
	}
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
