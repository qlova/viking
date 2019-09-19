package compiler

import (
	"bytes"
	"fmt"

	"github.com/qlova/viking/compiler/target"
)

//Field is a thing field.
type Field struct {
	Type
	Exported bool
}

//Thing is a structured type.
type Thing struct {
	Fields map[string]Field
}

//Name returns the name of this type.
func (Thing) Name() String {
	return String{
		English: `thing`,
	}
}

func (Thing) String(c *Compiler) string {
	return Thing{}.Name()[c.Language]
}

//Index does nothing.
func (thing Thing) Index(c *Compiler, this Expression, name Token) (expression Expression, err error) {
	expression = c.NewExpression()

	if field, ok := thing.Fields[name.String()]; ok {
		expression.Type = field.Type

		fmt.Fprintf(&expression.Go, `%v.%v`, this.Go, name)
		return expression, nil
	}
	return expression, c.NewError("no such field: ", name.String())
}

//Expression does nothing.
func (thing Thing) Expression(c *Compiler) (ok bool, expression Expression, err error) {
	expression = c.NewExpression()

	if c.Token().Is("{") {

		var context = c.NewContext()
		//Simple case. A function with an unknown return value.
		context.GainScope()
		c.FlipBuffer()

		c.GainScope()
		c.SetFlag(Token("thing"))

		thing.Fields = make(map[string]Field)
		c.SetVariable(Token("thing"), thing)

		var scope = len(c.Scope)
		for {
			if c.Peek().Is("}") && len(c.Scope) == scope {
				c.Scan()
				c.LoseScope()
				break
			}
			err := c.CompileStatement()
			if err != nil {
				return true, expression, err
			}
			c.Go.WriteString("\n")
		}
		c.Indent()

		fmt.Fprintf(&c.Go, `return %v{`, thing.Native(c))
		for name := range thing.Fields {
			fmt.Fprintf(&c.Go, `%v: %v,`, name, name)
		}
		fmt.Fprintf(&c.Go, `}`)

		fmt.Fprintf(&expression.Go, `func() %v {`, thing.Native(c))

		expression.Go.Write(c.DumpAndReturnBuffer(nil))

		fmt.Fprintf(&expression.Go, `}()`)

		expression.Type = thing

		return true, expression, nil
	}
	return
}

//Operation does nothing.
func (Thing) Operation(c *Compiler, a, b Expression, symbol string) (ok bool, expression Expression, err error) {
	return
}

//Cast does nothing.
func (Thing) Cast(c *Compiler, from Expression, to Type) (expression Expression, err error) {
	return c.CastingError(from, to)
}

//Equals returns true if the other type is equal to this type.
func (Thing) Equals(other Type) bool {
	_, ok := other.(Thing)
	return ok
}

//Native returns this type's native token.
func (thing Thing) Native(c *Compiler) (token Token) {
	if c.Target == target.Go {
		var buffer bytes.Buffer
		buffer.WriteString("struct{")
		for name, field := range thing.Fields {
			buffer.WriteString(name)
			buffer.WriteString(" ")
			buffer.Write(field.Native(c))
			buffer.WriteString(";")
		}
		buffer.WriteString("}")
		return buffer.Bytes()
	}
	return
}

//Zero returns this type's zero expression.
func (Thing) Zero(c *Compiler) (expression Expression) {
	expression = c.NewExpression()
	expression.Type = Nothing{}

	expression.Go.WriteString(`struct{}{}`)
	return
}

//Copy returns a copy of the nothing type.
func (Thing) Copy(c *Compiler, item Expression) (expression Expression, err error) {
	expression = c.NewExpression()
	expression.Type = Nothing{}

	expression.Go.WriteString(`struct{}{}`)
	return
}
