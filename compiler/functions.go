package compiler

//CallFunction calls a function with the specified name.
func (compiler *Compiler) CallFunction(name Token) error {
	var function = compiler.GetVariable(name)
	if !Defined(function) || !function.Is(Function) {
		return compiler.NewError(name.String() + " is not a function")
	}

	if !compiler.ScanIf('(') {
		return compiler.Expecting('(')
	}

	if !compiler.ScanIf(')') {
		return compiler.Expecting(')')
	}

	compiler.Indent()
	compiler.Go.Write(name)
	compiler.Go.WriteString("(ctx)")

	return nil
}
