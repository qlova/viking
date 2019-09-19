package builtin

import (
	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/types"
)

//Out is an interface around standard output.
type Out struct {
	compiler.Nothing
}

var _ = compiler.RegisterBuiltin(Out{})

//Name returns out's name.
func (Out) Name() compiler.String {
	return compiler.String{
		compiler.English: `out`,
	}
}

//Call does nothing.
func (Out) Call(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) (expression compiler.Expression, err error) {
	return expression, c.NewError("out cannot be called as an expression")
}

//Run runs out.
func (Out) Run(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) (err error) {
	if len(args) == 0 {
		c.Go.Write([]byte("fmt.Print()"))
		return
	}

	c.Import("fmt")

	c.Go.Write([]byte("fmt.Print("))

	for i, argument := range args {
		if argument.Type.Equals(types.Symbol{}) {
			c.Go.Write([]byte("string("))
			c.Go.Write(argument.Go.Bytes())
			c.Go.Write([]byte(")"))
		} else {
			c.Go.Write(argument.Go.Bytes())
		}
		if i < len(args)-1 {
			c.Go.WriteString(",")
		}
	}

	c.Go.Write([]byte(")"))

	return
}
