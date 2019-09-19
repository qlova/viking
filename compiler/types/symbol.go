package types

import (
	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/target"
)

//Symbol is an 'i' symbol.
type Symbol struct {
	compiler.Nothing
}

var _ = compiler.RegisterType(Symbol{})

//Name returns the name of this type.
func (Symbol) Name() compiler.String {
	return compiler.String{
		compiler.English: `symbol`,
	}
}

func (Symbol) String(c *compiler.Compiler) string {
	return Symbol{}.Name()[c.Language]
}

//Expression does nothing.
func (Symbol) Expression(c *compiler.Compiler) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if c.Token()[0] == '\'' {
		expression.Type = Symbol{}
		expression.Go.Write(c.Token())
		return true, expression, nil
	}

	return
}

//Operation does nothing.
func (Symbol) Operation(c *compiler.Compiler, a, b compiler.Expression, symbol string) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	return
}

//Cast does nothing.
func (Symbol) Cast(c *compiler.Compiler, from compiler.Expression, to compiler.Type) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if to.Equals(String{}) {
		expression.Go.WriteString("string(")
		expression.Go.Write(from.Go.Bytes())
		expression.Go.WriteString(")")
		return expression, nil
	}

	if to.Equals(Integer{}) {
		expression.Go.WriteString("I.NewInteger(int64(")
		expression.Go.Write(from.Go.Bytes())
		expression.Go.WriteString("))")
		return expression, nil
	}

	return c.CastingError(from, to)
}

//Equals returns true if the other type is equal to this type.
func (Symbol) Equals(other compiler.Type) bool {
	_, ok := other.(Symbol)
	return ok
}

//Native returns this type's native token.
func (Symbol) Native(c *compiler.Compiler) (token compiler.Token) {
	if c.Target == target.Go {
		return compiler.Token("rune")
	}
	return
}

//Zero returns this type's zero expression.
func (Symbol) Zero(c *compiler.Compiler) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = String{}

	expression.Go.WriteString(`rune(0)`)

	return
}

//Copy returns a copy of the nothing type.
func (Symbol) Copy(c *compiler.Compiler, item compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()
	expression.Type = Symbol{}

	expression.Go.WriteB(item.Go)

	return
}
