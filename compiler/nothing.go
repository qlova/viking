package compiler

import (
	"github.com/qlova/viking/compiler/target"
)

//Nothing is an empty type.
type Nothing struct{}

//Name returns the name of this type.
func (Nothing) Name() String {
	return String{
		English: `nothing`,
	}
}

func (Nothing) String(c *Compiler) string {
	return Nothing{}.Name()[c.Language]
}

//Expression does nothing.
func (Nothing) Expression(c *Compiler) (ok bool, expression Expression, err error) {
	return
}

//Operation does nothing.
func (Nothing) Operation(c *Compiler, a, b Expression, symbol string) (ok bool, expression Expression, err error) {
	return
}

//Cast does nothing.
func (Nothing) Cast(c *Compiler, from Expression, to Type) (expression Expression, err error) {
	return c.CastingError(from, to)
}

//Equals returns true if the other type is equal to this type.
func (Nothing) Equals(other Type) bool {
	_, ok := other.(Nothing)
	return ok
}

//Native returns this type's native token.
func (Nothing) Native(c *Compiler) (token Token) {
	if c.Target == target.Go {
		return Token("struct{}")
	}
	return
}

//Zero returns this type's zero expression.
func (Nothing) Zero(c *Compiler) (expression Expression) {
	expression = c.NewExpression()
	expression.Type = Nothing{}

	expression.Go.WriteString(`struct{}`)
	return
}

//Copy returns a copy of the nothing type.
func (Nothing) Copy(c *Compiler, item Expression) (expression Expression, err error) {
	expression = c.NewExpression()
	expression.Type = Nothing{}

	expression.Go.WriteString(`struct{}`)
	return
}
