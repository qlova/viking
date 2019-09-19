package types

import (
	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/target"
)

//Logical is an 'i' logical.
type Logical struct {
	compiler.Nothing
}

var _ = compiler.RegisterType(Logical{})

//Name returns the name of this type.
func (Logical) Name() compiler.String {
	return compiler.String{
		compiler.English: `logical`,
	}
}

func (Logical) String(c *compiler.Compiler) string {
	return Logical{}.Name()[c.Language]
}

//Expression does nothing.
func (Logical) Expression(c *compiler.Compiler) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if c.Token().Is("true") || c.Token().Is("false") {
		expression.Go.Write(c.Token())
		return true, expression, nil
	}

	if c.Token().Is("!") {

		var boolean, err = c.Expression(c.Scan())
		if err != nil {
			return true, compiler.Expression{}, err
		}

		if !boolean.Equals(Logical{}) {
			return true, compiler.Expression{}, c.NewError("cannot apply not operator to value of type " + boolean.String(c))
		}

		expression.Go.WriteString("(!")
		expression.Go.Write(boolean.Go.Bytes())
		expression.Go.WriteString(")")
		return true, expression, nil
	}

	return
}

//Operation does nothing.
func (Logical) Operation(c *compiler.Compiler, a, b compiler.Expression, symbol string) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	switch symbol {
	case "&":
		if b.Type.Equals(Logical{}) {
			expression.Type = Logical{}

			expression.Go.WriteString(`(`)
			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`&&`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`)`)

			return true, expression, nil
		}
	case "|":
		if b.Type.Equals(Logical{}) {
			expression.Type = Logical{}

			expression.Go.WriteString(`(`)
			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`||`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`)`)

			return true, expression, nil
		}
	case "-":
		if b.Type.Equals(Logical{}) {
			expression.Type = Logical{}

			expression.Go.WriteString(`(`)
			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`!=`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`)`)

			return true, expression, nil
		}
	}

	return
}

//Cast does nothing.
func (Logical) Cast(c *compiler.Compiler, from compiler.Expression, to compiler.Type) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	return c.CastingError(from, to)
}

//Equals returns true if the other type is equal to this type.
func (Logical) Equals(other compiler.Type) bool {
	_, ok := other.(Logical)
	return ok
}

//Native returns this type's native token.
func (Logical) Native(c *compiler.Compiler) (token compiler.Token) {
	if c.Target == target.Go {
		return compiler.Token("bool")
	}
	return
}

//Zero returns this type's zero expression.
func (Logical) Zero(c *compiler.Compiler) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = Logical{}

	expression.Go.WriteString(`false`)

	return
}

//Copy returns a copy of the nothing type.
func (Logical) Copy(c *compiler.Compiler, item compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()
	expression.Type = Logical{}

	expression.Go.WriteB(item.Go)

	return
}
