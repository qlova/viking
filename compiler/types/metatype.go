package types

import (
	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/target"
)

//Metatype is an 'i' metatype.
type Metatype struct {
	compiler.Nothing

	Type compiler.Type
}

var _ = compiler.RegisterType(Metatype{})

//Name returns the name of this type.
func (Metatype) Name() compiler.String {
	return compiler.String{
		compiler.English: `metatype`,
	}
}

func (Metatype) String(c *compiler.Compiler) string {
	return Metatype{}.Name()[c.Language]
}

//Equals returns true if the other type is equal to this type.
func (Metatype) Equals(other compiler.Type) bool {
	_, ok := other.(Metatype)
	return ok
}

//Expression does nothing.
func (Metatype) Expression(c *compiler.Compiler) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if T := c.Type(c.Token()); compiler.Defined(T) {

		//This could be inline target code.
		if c.Peek().Is("if") {
			T = c.Type(c.Token())

			for {
				var name = c.ScanAndIgnoreNewLines()
				if t := target.FromString(name.String()); t.Valid() {
					var code = c.ScanAndIgnoreNewLines()
					if code[0] != '`' {
						return true, compiler.Expression{}, c.NewError("expecting `[target code]`")
					}
					expression.Get(t).Write(code[1 : len(code)-1])
				}

				if name.Is("}") {
					break
				}
				if name == nil {
					return true, compiler.Expression{}, c.NewError("if block wasn't closed")
				}
			}
			expression.Type = T
			return true, expression, nil
		}

		T, err := c.SpecifyType(T)
		if err != nil {
			return true, expression, err
		}

		expression.Type = Metatype{
			Type: T,
		}

		return true, expression, nil
	}

	return
}

//Operation does nothing.
func (Metatype) Operation(c *compiler.Compiler, a, b compiler.Expression, symbol string) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	switch symbol {
	case "=":
		if b.Type.Equals(Metatype{}) {
			expression.Type = Logical{}

			if a.Type.(Metatype).Type.Equals(b.Type.(Metatype).Type) {
				expression.Go.WriteString("true")
			} else {
				expression.Go.WriteString("false")
			}

			return true, expression, nil
		}
	}
	return
}

var _ = compiler.Runnable(Metatype{})

//Call calls the metatype.
func (meta Metatype) Call(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if len(args) == 0 {
		return meta.Type.Zero(c), nil
	}

	if len(args) == 1 {
		return c.Cast(args[0], meta.Type)
	}

	return compiler.Expression{}, c.NewError("invalid number of arguments passed to type")
}

//Run does nothing.
func (Metatype) Run(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) (err error) {
	return c.NewError("metatype cannot be called as statement")
}
