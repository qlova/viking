package types

import (
	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/target"
)

//Number can contain any numeric value.
type Number struct {
	compiler.Nothing
}

var _ = compiler.RegisterType(Number{})

//Name returns the name of this type.
func (Number) Name() compiler.String {
	return compiler.String{
		compiler.English: `number`,
	}
}

func (Number) String(c *compiler.Compiler) string {
	return Number{}.Name()[c.Language]
}

//Equals returns true if the other type is equal to this type.
func (Number) Equals(other compiler.Type) bool {
	_, ok := other.(Number)
	return ok
}

//Expression does nothing.
func (Number) Expression(c *compiler.Compiler) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	return
}

//Operation does nothing.
func (Number) Operation(c *compiler.Compiler, a, b compiler.Expression, symbol string) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	return
}

//Cast does nothing.
func (Number) Cast(c *compiler.Compiler, from compiler.Expression, to compiler.Type) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	return c.CastingError(from, to)
}

//Native returns this type's native token.
func (Number) Native(c *compiler.Compiler) (token compiler.Token) {
	if c.Target == target.Go {
		return compiler.Token("I.Number")
	}
	return
}

//Zero returns this type's zero expression.
func (Number) Zero(c *compiler.Compiler) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = Integer{}

	expression.Go.WriteString(`I.Number{}`)

	return
}

//Copy returns a copy of the nothing type.
func (Number) Copy(c *compiler.Compiler, item compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()
	expression.Type = Number{}

	expression.Go.WriteB(item.Go)
	expression.Go.WriteString(`.Copy()`)

	return
}
