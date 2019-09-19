package builtin

import (
	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/types"
)

//Print is a simple stdout interface that adds spaces between arguments.
type Print struct {
	compiler.Nothing
}

var _ = compiler.RegisterBuiltin(Print{})

//Name returns print's name.
func (Print) Name() compiler.String {
	return compiler.String{
		compiler.English: `print`,
	}
}

//Call does nothing.
func (Print) Call(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) (expression compiler.Expression, err error) {
	return expression, c.NewError("out cannot be called as an expression")
}

//Run runs out.
func (Print) Run(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) (err error) {
	if len(args) == 0 {
		c.Go.Write([]byte("fmt.Println()"))
		c.JS.Write([]byte("console.log()"))
		return
	}

	c.Import("fmt")

	c.JS.WriteString(`console.log(`)
	c.Go.Write([]byte("fmt.Println("))

	for i, argument := range args {
		if argument.Type.Equals(types.Symbol{}) {
			c.Go.Write([]byte("string("))
			c.Go.Write(argument.Go.Bytes())
			c.Go.Write([]byte(")"))
		} else {
			c.Go.Write(argument.Go.Bytes())
			c.JS.Write(argument.JS.Bytes())
		}
		if i < len(args)-1 {
			c.Go.WriteString(",")
			c.JS.WriteString(",")
		}
	}

	c.Go.Write([]byte(")"))
	c.JS.Write([]byte(")"))

	return
}
