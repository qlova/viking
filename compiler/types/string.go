package types

import (
	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/target"
)

//String is an 'i' string.
type String struct {
	compiler.Nothing
}

var _ = compiler.RegisterType(String{})

//Name returns the name of this type.
func (String) Name() compiler.String {
	return compiler.String{
		compiler.English: `string`,
	}
}

func (String) String(c *compiler.Compiler) string {
	return String{}.Name()[c.Language]
}

//Expression does nothing.
func (String) Expression(c *compiler.Compiler) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if c.Token()[0] == '"' {
		expression.Go.Write(c.Token())
		expression.JS.Write(c.Token())
		return true, expression, nil
	}

	return
}

//Operation does nothing.
func (String) Operation(c *compiler.Compiler, a, b compiler.Expression, symbol string) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	switch symbol {
	case "+":
		if b.Type.Equals(String{}) {
			expression.Type = String{}

			expression.Go.WriteB(a.Go)
			expression.Go.WriteString(`+`)
			expression.Go.WriteB(b.Go)

			expression.JS.WriteB(a.JS)
			expression.JS.WriteString(`+`)
			expression.JS.WriteB(b.JS)

			return true, expression, nil
		}
	}
	return
}

//Cast does nothing.
func (String) Cast(c *compiler.Compiler, from compiler.Expression, to compiler.Type) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if to.Equals(Integer{}) {
		c.Throws = true

		expression.Go.WriteString("I.Atoi(ctx, ")
		expression.Go.Write(from.Go.Bytes())
		expression.Go.WriteString(")")
		return expression, nil
	}

	if to.Equals(Number{}) {
		c.Throws = true

		expression.Go.WriteString("ctx.Aton(")
		expression.Go.Write(from.Go.Bytes())
		expression.Go.WriteString(")")
		return expression, nil
	}

	return c.CastingError(from, to)
}

//Equals returns true if the other type is equal to this type.
func (String) Equals(other compiler.Type) bool {
	_, ok := other.(String)
	return ok
}

//Native returns this type's native token.
func (String) Native(c *compiler.Compiler) (token compiler.Token) {
	if c.Target == target.Go {
		return compiler.Token("string")
	}
	if c.Target == target.JS {
		return compiler.Token("String")
	}
	return
}

//Zero returns this type's zero expression.
func (String) Zero(c *compiler.Compiler) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = String{}

	expression.Go.WriteString(`""`)
	expression.JS.WriteString(`""`)

	return
}

//Copy returns a copy of the nothing type.
func (String) Copy(c *compiler.Compiler, item compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()
	expression.Type = String{}

	expression.Go.WriteB(item.Go)

	expression.JS.WriteString(`(' ' + `)
	expression.JS.WriteB(item.JS)
	expression.JS.WriteString(`).slice(1)`)

	return
}

var _ = compiler.Collection(String{})

//Length returns the size/length/count of this type.
func (String) Length(c *compiler.Compiler, this compiler.Expression) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = Integer{}
	expression.Go.WriteString(`ctx.CountString(`)
	expression.Go.WriteB(this.Go)
	expression.Go.WriteString(`)`)
	return expression
}

//Subtype returns the subtype.
func (String) Subtype() compiler.Type {
	return Symbol{}
}

//Index a value of this type with the specified indicies.
func (String) Index(c *compiler.Compiler, this compiler.Expression, indices ...compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if len(indices) != 1 {
		return expression, c.NewError("array takes 1 symbol index")
	}

	var index = indices[0]

	expression.Type = Symbol{}
	expression.Go.WriteString(`ctx.Strindex(`)
	expression.Go.WriteB(this.Go)
	expression.Go.WriteString(`,`)
	expression.Go.WriteB(index.Go)
	expression.Go.WriteString(`)`)

	return expression, nil
}

//Modify a value of this type with the specified indicies.
func (String) Modify(c *compiler.Compiler, this compiler.Expression, modification compiler.Expression, indices ...compiler.Expression) error {
	return c.NewError("strings are immutable")
}

//Specify this type with the provided args.
func (String) Specify(c *compiler.Compiler, args ...compiler.Expression) (compiler.Type, error) {
	return nil, c.NewError("string doesn't take any arguments")
}

//With should return this type containing the provided subtype.
func (String) With(c *compiler.Compiler, subtype compiler.Type) compiler.Type {
	return String{}
}
