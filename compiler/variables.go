package compiler

//Defined returns true if T is defined.
func Defined(T Type) bool {
	return T != nil
}

//SetVariable sets a new variable.
func (compiler *Context) SetVariable(name []byte, T Type) {
	compiler.Scope[len(compiler.Scope)-1].Table[string(name)] = T
}

//GetVariable returns the variable with the given name.
func (compiler *Context) GetVariable(name Token) Type {
	if len(compiler.Scope) <= 0 {
		return nil
	}
	for i := len(compiler.Scope) - 1; i >= 0; i-- {
		if v, ok := compiler.Scope[i].Table[name.String()]; ok {
			return v
		}
	}
	return nil
}

//DefineVariable defines the variable 'name' with the scanned value.
func (compiler *Compiler) DefineVariable(name []byte) error {
	var expression, err = compiler.ScanExpression()
	if err != nil {
		return err
	}

	if compiler.Flag(Token("thing")) {
		var thing = compiler.GetVariable(Token("thing")).(Thing)
		thing.Fields[string(name)] = Field{
			Type: expression.Type,
		}
	}

	compiler.SetVariable(name, expression.Type)
	compiler.Go.Write([]byte("var "))
	compiler.Go.Write(name)
	compiler.Go.Write([]byte(" = "))
	compiler.Go.Write(expression.Go.Bytes())

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
		var old = expression
		expression, err = compiler.Cast(expression, variable)
		if err != nil {
			return compiler.NewError("cannot assign value of type " + old.Type.String(compiler) + " to variable of type " + variable.String(compiler))
		}
	}

	compiler.SetVariable(name, expression.Type)
	compiler.Go.Write(name)
	compiler.Go.Write([]byte(" = "))
	compiler.Go.Write(expression.Go.Bytes())

	return nil
}

//ShortcutAssignVariable modifies the variable 'name' with the scanned value.
func (compiler *Compiler) ShortcutAssignVariable(name []byte) error {
	var variable = compiler.GetVariable(name)

	var expression, err = compiler.ScanExpression()
	if err != nil {
		return err
	}
	if !expression.Type.Equals(variable) {
		return compiler.NewError("type mismatch")
	}

	compiler.SetVariable(name, expression.Type)
	compiler.Go.Write(name)
	compiler.Go.Write([]byte(" = "))
	compiler.Go.Write(expression.Go.Bytes())

	return nil
}
