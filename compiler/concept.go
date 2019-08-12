package compiler

import (
	"errors"
)

//Concept is a generic functions.
type Concept struct {
	Name      Token
	Arguments []Argument
	Cache
}

//RunConcept runs a concept with the specified name wihout return values.
func (compiler *Compiler) RunConcept(name Token) error {
	expression, err := compiler.CallConcept(name)
	compiler.Indent()
	compiler.Go.Write(expression.Go.Bytes())

	if err == errorConceptHasNoReturns {
		return nil
	}

	return err
}

//CallConcept runs a concept with the specified name wihout return values.
func (compiler *Compiler) CallConcept(name Token) (Expression, error) {

	var concept, ok = compiler.Concepts[name.String()]
	if !ok {
		return Expression{}, errors.New(name.String() + " is not a concept")
	}

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
				return Expression{}, errors.New("type mismatch got type " + expression.Type.Name + " expecting type " + argument.Type.Name)
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

var errorConceptHasNoReturns = errors.New("function does not return any values and cannot be used in an expression")

//CallConcept calls a concept with the specified name.
func (compiler *Compiler) generateAndCallConcept(concept Concept, arguments []Expression) (Expression, error) {

	if compiler.Functions == nil {
		compiler.Functions = make(map[string]struct{})
	}

	var returns *Type

	if _, ok := compiler.Functions[concept.Name.String()]; !ok {

		//Simple case. A function with an unknown return value.
		compiler.PushScope()
		compiler.GainScope()

		compiler.FlipBuffer()

		for i, argument := range arguments {
			if concept.Arguments[i].Variadic {
				compiler.SetVariable(concept.Arguments[i].Token, argument.Type.Collection(Variadic))
				break
			}
			compiler.SetVariable(concept.Arguments[i].Token, argument.Type)
		}

		var context Context
		context.Returns = &Type{}
		returns = context.Returns

		if err := compiler.CompileCacheWithContext(concept.Cache, context); err != nil {
			return Expression{}, err
		}

		compiler.PopScope()

		var FunctionHeader = compiler.Target

		//Build function definition.
		FunctionHeader.Go.WriteString("func ")
		FunctionHeader.Go.Write(concept.Name)
		FunctionHeader.Go.WriteString("(ctx *Context")

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
	compiler.Functions[concept.Name.String()] = struct{}{}

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
		return expression, errorConceptHasNoReturns
	}

	return expression, nil
}
