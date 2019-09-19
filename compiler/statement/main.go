package statement

import "github.com/qlova/viking/compiler"

//Main is the entrypoint of the application.
type Main struct{}

var _ = compiler.RegisterStatement(Main{})

//Name returns the name of this statement.
func (Main) Name() compiler.String {
	return compiler.String{
		compiler.English: `main`,
	}
}

//Compile compiles this statement.
func (Main) Compile(c *compiler.Compiler) error {
	c.SetMain()

	c.Import(compiler.Ilang)
	c.Go.WriteString("func main() {\n")
	c.JS.WriteString("function main() {\n")

	c.GainScope()
	c.Indent()

	c.Go.WriteString(`var ctx = I.NewContext()` + "\n")

	c.SetFlag(compiler.Token("main"))

	return c.CompileBlock()
}
