package builtin

import (
	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/types"
)

//In is an interface around stdin.
type In struct {
	compiler.Nothing
}

var _ = compiler.RegisterBuiltin(In{})

//Name returns in's name.
func (In) Name() compiler.String {
	return compiler.String{
		compiler.English: `in`,
	}
}

//Run does nothing.
func (In) Run(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) (err error) {
	return c.NewError("in cannot be called as statement")
}

//Call calls in.
func (In) Call(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()
	c.Throws = true

	if len(args) == 0 {
		expression.Type = types.String{}
		expression.Go.WriteString("I.InSymbol(ctx, '\n')")
		return expression, nil
	}

	if len(args) > 1 {
		return expression, c.NewError("too many arguments passed to in")
	}

	argument := args[0]

	if argument.Equals(types.Symbol{}) {
		c.Import(compiler.Ilang)

		expression.Type = types.String{}
		expression.Go.WriteString("I.InSymbol(ctx, ")
		expression.Go.Write(argument.Go.Bytes())
		expression.Go.WriteString(")")
		return expression, nil
	}

	return expression, c.NewError("invalid type " + expression.Type.String(c) + " passed to builtin")
}
