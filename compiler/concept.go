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
	compiler.Write(expression.Bytes())

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

	//Simple case. A function with an unknown return value.
	compiler.PushScope()
	compiler.GainScope()

	var buffer = compiler.FlipBuffer()

	for i, argument := range arguments {
		if concept.Arguments[i].Variadic {
			compiler.SetVariable(concept.Arguments[i].Token, argument.Type.Collection(Variadic))
			break
		}
		compiler.SetVariable(concept.Arguments[i].Token, argument.Type)
	}

	var context Context
	context.Returns = &Type{}
	var returns = context.Returns

	if err := compiler.CompileCacheWithContext(concept.Cache, context); err != nil {
		return Expression{}, err
	}

	compiler.PopScope()

	//Build function definition.
	buffer.Head.WriteString("func ")
	buffer.Head.Write(concept.Name)
	buffer.Head.WriteString("(")

	for i, argument := range concept.Arguments {
		buffer.Head.Write(argument.Token)
		buffer.Head.WriteString(" ")
		if concept.Arguments[i].Variadic {
			buffer.Head.WriteString("...")
		}
		buffer.Head.Write(GoTypeOf(arguments[i].Type))
		if i < len(arguments)-1 {
			buffer.Head.WriteString(",")
		}
	}

	buffer.Head.WriteString(")")
	if returns != nil && Defined(*returns) {
		buffer.Head.Write(GoTypeOf(*returns))
	}
	buffer.Head.WriteString("{\n")

	compiler.DumpBufferHead()

	var expression Expression
	if returns != nil {
		expression.Type = *returns
	}
	expression.Write(concept.Name)
	expression.WriteString("(")
	for i, argument := range arguments {
		expression.Write(argument.Bytes())
		if i != len(arguments)-1 {
			expression.WriteString(",")
		}
	}
	expression.WriteString(")")

	if returns == nil || !Defined(*returns) {
		return expression, errorConceptHasNoReturns
	}

	return expression, nil
}
