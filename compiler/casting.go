package compiler

//CastingError returns an error when a type cannot be cast.
func (compiler *Compiler) CastingError(from Expression, to Type) (Expression, error) {
	return Expression{}, compiler.Unimplemented([]byte("casting " + from.Type.String(compiler) + " to " + to.String(compiler)))
}

//Cast from expression to Type 'to'.
func (compiler *Compiler) Cast(from Expression, to Type) (Expression, error) {
	if from.Equals(to) {
		return from, nil
	}

	expression, err := from.Cast(compiler, from, to)
	if expression.Type == nil {
		expression.Type = to
	}
	if err != nil {
		expression, err := to.Cast(compiler, from, to)
		if expression.Type == nil {
			expression.Type = to
		}
		return expression, err
	}
	return expression, nil
}
