package types

import (
	"fmt"
	"strconv"

	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/target"
)

type Integer struct {
	compiler.Nothing
}

func init() {
	compiler.SequenceLength = func(c *compiler.Compiler, this compiler.Expression) (expression compiler.Expression) {
		expression = c.NewExpression()
		expression.Type = Integer{}
		expression.Go.WriteString(`I.NewInteger(int64(`)
		expression.Go.WriteString(strconv.Itoa(this.Type.(compiler.Sequence).Size))
		expression.Go.WriteString(`))`)
		return
	}
}

var _ = compiler.RegisterType(Integer{})

//Name returns the name of this type.
func (Integer) Name() compiler.String {
	return compiler.String{
		compiler.English: `integer`,
	}
}

func (Integer) String(c *compiler.Compiler) string {
	return Integer{}.Name()[c.Language]
}

//Expression does nothing.
func (Integer) Expression(c *compiler.Compiler) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	//Binary expression.
	if i, err := strconv.ParseInt(string(c.Token()), 2, 64); err == nil && len(c.Token()) > 0 && c.Token()[0] == '0' {
		expression.Go.WriteString("I.NewInteger(")
		expression.Go.WriteString(strconv.Itoa(int(i)))
		expression.Go.WriteString(")")
		return true, expression, nil
	}

	//Hexadecimal expression.
	if len(c.Token()) > 2 && c.Token()[0] == '0' && c.Token()[1] == 'x' {
		expression.Go.WriteString("I.NewInteger(")
		expression.Go.Write(c.Token())
		expression.Go.WriteString(")")
		return true, expression, nil
	}

	//Integer expression.
	if _, err := strconv.Atoi(string(c.Token())); err == nil {
		expression.Go.WriteString("I.NewInteger(")
		expression.Go.Write(c.Token())
		expression.Go.WriteString(")")
		return true, expression, nil
	}

	return
}

//Operation does nothing.
func (Integer) Operation(c *compiler.Compiler, a, b compiler.Expression, symbol string) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	switch symbol {
	case "+":
		if b.Type.Equals(Integer{}) {
			expression.Type = Integer{}

			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`.Add(`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`)`)

			return true, expression, nil
		}
	case "*":
		if b.Type.Equals(Integer{}) {
			expression.Type = Integer{}

			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`.Mul(`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`)`)

			return true, expression, nil
		}
	case "/":
		if b.Type.Equals(Integer{}) {
			expression.Type = Integer{}

			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`.Div(`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`)`)

			return true, expression, nil
		}
	case "-":
		if b.Type.Equals(Integer{}) {
			expression.Type = Integer{}

			if a.Type == nil {
				fmt.Fprintf(&expression.Go, `%v.Neg()`, b.Go)
			} else {
				fmt.Fprintf(&expression.Go, `%v.Sub(%v)`, a.Go, b.Go)
			}

			return true, expression, nil
		}
	case "^":
		if b.Type.Equals(Integer{}) {
			expression.Type = Integer{}

			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`.Pow(`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`)`)

			return true, expression, nil
		}
	case "%":
		if b.Type.Equals(Integer{}) {
			expression.Type = Integer{}

			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`.Mod(`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`)`)

			return true, expression, nil
		}

	case "=":
		if b.Type.Equals(Integer{}) {
			expression.Type = Logical{}

			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`.Equals(`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`)`)

			return true, expression, nil
		}

	case "!":
		if b.Type.Equals(Integer{}) {
			expression.Type = Logical{}

			expression.Go.WriteString(`!`)
			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`.Equals(`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`)`)

			return true, expression, nil
		}

	case ">":
		if b.Type.Equals(Integer{}) {
			expression.Type = Logical{}

			expression.Go.WriteString(`(`)
			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`.Compare(`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`) > 0)`)

			return true, expression, nil
		}
	case "<":
		if b.Type.Equals(Integer{}) {
			expression.Type = Logical{}

			expression.Go.WriteString(`(`)
			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`.Compare(`)
			expression.Go.WriteB(b.Go)
			expression.Go.WriteString(`) < 0)`)

			return true, expression, nil
		}
	}
	return
}

//Cast does nothing.
func (Integer) Cast(c *compiler.Compiler, from compiler.Expression, to compiler.Type) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if to.Equals(Symbol{}) {
		expression.Go.WriteString("rune(")
		expression.Go.Write(from.Go.Bytes())
		expression.Go.WriteString(".Int64())")
		return expression, nil
	}

	if to.Equals(Logical{}) {
		expression.Go.WriteString("bool(")
		expression.Go.Write(from.Go.Bytes())
		expression.Go.WriteString(".Bool())")
		return expression, nil
	}

	return c.CastingError(from, to)
}

//Equals returns true if the other type is equal to this type.
func (Integer) Equals(other compiler.Type) bool {
	_, ok := other.(Integer)
	return ok
}

//Native returns this type's native token.
func (Integer) Native(c *compiler.Compiler) (token compiler.Token) {
	if c.Target == target.Go {
		return compiler.Token("I.Integer")
	}
	return
}

//Zero returns this type's zero expression.
func (Integer) Zero(c *compiler.Compiler) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = Integer{}

	expression.Go.WriteString(`I.Integer{}`)

	return
}

//Copy returns a copy of the nothing type.
func (Integer) Copy(c *compiler.Compiler, item compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()
	expression.Type = Integer{}

	expression.Go.WriteB(item.Go)

	return
}
