package compiler

//Concept is a generic functions.
type Concept struct {
	Name      Token
	Arguments []Argument
	Cache
}

//Generate generates and returns the name and return type of this function.
func (concept Concept) Generate(compiler *Compiler, args ...Expression) (name Token, returns Type, err error) {
	if compiler.Functions == nil {
		compiler.Functions = make(map[string]Type)
	}

	var id = concept.Name.String()

	if r, ok := compiler.Functions[id]; ok {
		returns = r
	} else if !ok {

		var context = compiler.NewContext()
		context.Returns = &returns

		//Simple case. A function with an unknown return value.
		context.GainScope()

		compiler.FlipBuffer()

		for i, argument := range args {
			if concept.Arguments[i].Variadic {
				context.SetVariable(concept.Arguments[i].Token, Sequence{}.With(compiler, argument.Type))
				break
			}
			context.SetVariable(concept.Arguments[i].Token, argument.Type)
		}

		if err := compiler.CompileCacheWithContext(concept.Cache, context); err != nil {
			return Token(id), returns, err
		}

		var FunctionHeader = compiler.target

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
			FunctionHeader.Go.Write(compiler.GoTypeOf(args[i].Type))
		}

		FunctionHeader.Go.WriteString(")")
		if returns != nil && Defined(returns) {
			FunctionHeader.Go.Write(compiler.GoTypeOf(returns))
		}
		FunctionHeader.Go.WriteString("{\n")

		compiler.DumpBufferHead(FunctionHeader.Go.Bytes())
	}
	compiler.Functions[concept.Name.String()] = returns

	return Token(id), returns, nil
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
				return Expression{}, compiler.NewError("type mismatch got type " + expression.Type.String(compiler) + " expecting type " + argument.Type.String(compiler))
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

	name, returns, err := concept.Generate(compiler, arguments...)
	if err != nil {
		return Expression{}, err
	}

	var expression = compiler.NewExpression()
	expression.Type = returns
	expression.Go.Write(name)
	expression.Go.WriteString("(ctx")
	for _, argument := range arguments {
		expression.Go.WriteString(",")
		expression.Go.Write(argument.Go.Bytes())
	}
	expression.Go.WriteString(")")

	if !Defined(returns) {
		return expression, compiler.NewError(errorConceptHasNoReturns)
	}

	return expression, nil
}
