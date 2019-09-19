package statement

import (
	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/types"
)

//For is an for statement.
type For struct{}

var _ = compiler.RegisterStatement(For{})

//Name returns the name of this statement.
func (For) Name() compiler.String {
	return compiler.String{
		compiler.English: `for`,
	}
}

//Compile compiles this statement.
func (For) Compile(c *compiler.Compiler) error {
	if c.Peek().Is(":") {
		c.Indent()
		c.Go.WriteString("for i := I.NewInteger(1); true; i = i.Add(I.NewInteger(1)) {")

		c.GainScope()
		c.SetVariable(compiler.Token("i"), types.Integer{})

		return c.CompileBlock()
	}

	var name = c.Scan()

	var numeric bool

	if v := c.GetVariable(name); compiler.Defined(v) && v.Equals(types.Integer{}) {
		numeric = true
	}

	if ok, _, err := (types.Integer{}).Expression(c); ok {
		if err != nil {
			return err
		}
		numeric = true
	}

	var collection compiler.Collection
	if !numeric && c.Peek().Is(":") {
		var ok bool

		expression, err := c.Expression(name)
		if err != nil {
			return err
		}

		expression, err = c.Shunt(expression, 0)
		if err != nil {
			return err
		}

		collection, ok = expression.Type.(compiler.Collection)
		if !ok {
			return c.NewError("short for loops must be of collection type ,not", expression.String(c))
		}

	}

	if numeric || collection != nil {
		var err error

		var expression compiler.Expression

		if collection != nil {

			expression = collection.Length(c, expression)

		} else {

			expression, err = c.Expression(name)
			if err != nil {
				return err
			}

			expression, err = c.Shunt(expression, 0)
			if err != nil {
				return err
			}

			if c.Peek().Is("in") {
				c.Scan()

				var target, err = c.ScanExpression()
				if err != nil {
					return err
				}

				if target.Equals(types.Integer{}) {

					c.Go.WriteString("for i, in, to := I.SetupStep(")
					c.Go.Write(target.Go.Bytes())
					c.Go.WriteString(",")
					c.Go.Write(expression.Go.Bytes())
					c.Go.WriteString("); i.CompareStep(to, in); i = i.Add(in) {")

					c.GainScope()
					c.SetVariable(compiler.Token("i"), types.Integer{})
					return c.CompileBlock()
				}

			}

			if c.Peek().Is("to") {
				c.Scan()
				var to, err = c.ScanExpression()
				if err != nil {
					return err
				}
				c.Go.WriteString("for i, to := I.SetupTo(")
				c.Go.Write(expression.Go.Bytes())
				c.Go.WriteString(",")
				c.Go.Write(to.Go.Bytes())
				c.Go.WriteString("); i.Compare(to) != 0; i = i.To(to) {")

				c.GainScope()
				c.SetVariable(compiler.Token("i"), types.Integer{})
				return c.CompileBlock()
			}
		}

		c.Go.WriteString("for ")
		c.Go.WriteString("i := I.NewInteger(1); i.Compare(")
		c.Go.Write(expression.Go.Bytes())
		c.Go.WriteString(") <= 0; i = i.Add(I.NewInteger(1)) {")
		c.GainScope()
		c.SetVariable(compiler.Token("i"), types.Integer{})

		return c.CompileBlock()
	}

	if !c.Scan().Is("in") {
		return c.NewError("expecting 'in'")
	}

	expression, err := c.ScanExpression()
	if err != nil {
		return err
	}

	if _, ok := expression.Type.(compiler.Collection); !ok {
		return c.NewError("unimplemented for loop for " + expression.String(c))
	}

	c.Indent()
	c.Go.WriteString("for ")
	c.Go.WriteString("i,")
	c.Go.Write(name)
	c.Go.WriteString(":= range ")
	c.Go.Write(expression.Go.Bytes())
	c.Go.WriteString("{")

	c.GainScope()
	c.SetVariable(name, expression.Type.(compiler.Collection).Subtype())
	c.SetVariable(compiler.Token("i"), types.Integer{})

	return c.CompileBlock()
}
