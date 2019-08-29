package compiler

//Concept is a generic functions.
type Concept struct {
	Name      Token
	Arguments []Argument
	Cache
}

//Run runs a concept with the specified name wihout return values.
func (concept Concept) Run(compiler *Compiler) error {
	expression, err := concept.Call(compiler)
	compiler.Indent()
	compiler.Go.Write(expression.Go.Bytes())

	if CompilerErr, ok := err.(Error); ok && CompilerErr.Message == errorConceptHasNoReturns {
		return nil
	}

	return err
}

//Call runs a concept with the specified name wihout return values.
func (concept Concept) Call(compiler *Compiler) (Expression, error) {
	if !compiler.ScanIf('(') {
		return Expression{}, compiler.Expecting('(')
	}

	var Arguments = make([]Expression, len(concept.Arguments))
	var extra = false

	for i := range Arguments {
		Arguments[i] = compiler.NewExpression()
	}

	for i, argument := range concept.Arguments {

	variadic:

		var expression, err = compiler.ScanExpression()
		if err != nil {
			return Expression{}, err
		}

		if Defined(argument.Type) && !expression.Equals(argument.Type) {
			expression, err = compiler.Cast(expression, argument.Type)
			if err != nil {
				return Expression{}, compiler.NewError("type mismatch got type " + expression.Type.Name + " expecting type " + argument.Type.Name)
			}
		}

		if argument.Variadic && extra {
			Arguments = append(Arguments, expression)
		} else {
			Arguments[i] = expression
			if argument.Variadic {
				extra = true
			}
		}

		if i < len(concept.Arguments)-1 || argument.Variadic {
			if !compiler.ScanIf(',') {
				if compiler.Peek().Is(")") {
					break
				}
				return Expression{}, compiler.Expecting(',')
			}
			if argument.Variadic {
				goto variadic
			}
		}
	}

	if !compiler.ScanIf(')') {
		return Expression{}, compiler.Expecting(')')
	}

	return compiler.generateAndCallConcept(concept, Arguments)
}

var errorConceptHasNoReturns = "function does not return any values and cannot be used in an expression"

//CallConcept calls a concept with the specified name.
func (compiler *Compiler) generateAndCallConcept(concept Concept, arguments []Expression) (Expression, error) {

	if compiler.Functions == nil {
		compiler.Functions = make(map[string]*Type)
	}

	var returns = new(Type)

	if r, ok := compiler.Functions[concept.Name.String()]; ok {
		returns = r
	} else if !ok {

		var context = compiler.NewContext()
		context.Returns = returns

		//Simple case. A function with an unknown return value.
		context.GainScope()

		compiler.FlipBuffer()

		for i, argument := range arguments {
			if concept.Arguments[i].Variadic {
				context.SetVariable(concept.Arguments[i].Token, argument.Type.Collection(Variadic))
				break
			}
			context.SetVariable(concept.Arguments[i].Token, argument.Type)
		}

		if err := compiler.CompileCacheWithContext(concept.Cache, context); err != nil {
			return Expression{}, err
		}

		var FunctionHeader = compiler.Target

		//Build function definition.
		FunctionHeader.Go.WriteString("func ")
		FunctionHeader.Go.Write(concept.Name)
		FunctionHeader.Go.WriteString("(ctx I.Context")

		for i, argument := range concept.Arguments {
			FunctionHeader.Go.WriteString(",")
			FunctionHeader.Go.Write(argument.Token)
			FunctionHeader.Go.WriteString(" ")
			if concept.Arguments[i].Variadic {
				FunctionHeader.Go.WriteString("...")
			}
			FunctionHeader.Go.Write(GoTypeOf(arguments[i].Type))
		}

		FunctionHeader.Go.WriteString(")")
		if returns != nil && Defined(*returns) {
			FunctionHeader.Go.Write(GoTypeOf(*returns))
		}
		FunctionHeader.Go.WriteString("{\n")

		compiler.DumpBufferHead(FunctionHeader.Go.Bytes())
	}
	compiler.Functions[concept.Name.String()] = returns

	var expression = compiler.NewExpression()
	if returns != nil {
		expression.Type = *returns
	}
	expression.Go.Write(concept.Name)
	expression.Go.WriteString("(ctx")
	for _, argument := range arguments {
		expression.Go.WriteString(",")
		expression.Go.Write(argument.Go.Bytes())
	}
	expression.Go.WriteString(")")

	if returns == nil || !Defined(*returns) {
		return expression, compiler.NewError(errorConceptHasNoReturns)
	}

	return expression, nil
}
