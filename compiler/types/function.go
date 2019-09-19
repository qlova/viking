package types

import (
	"fmt"

	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/target"
)

//Function is an 'i' function.
type Function struct {
	subtype compiler.Type

	Concept compiler.Concept
	compiler.Nothing
}

var _ = compiler.RegisterType(Function{})

var _ = compiler.Collection(Function{})

//Name returns the name of this type.
func (Function) Name() compiler.String {
	return compiler.String{
		compiler.English: `function`,
	}
}

//Subtype returns the subtype.
func (function Function) Subtype() compiler.Type {
	return function.subtype
}

func (Function) String(c *compiler.Compiler) string {
	return Function{}.Name()[c.Language]
}

//Expression checks and returns a function.
func (function Function) Expression(c *compiler.Compiler) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if concept, ok := c.Concepts[c.Token().String()]; ok {
		if !c.Peek().Is("(") && len(concept.Arguments) == 0 {
			var _, returns, err = concept.Generate(c)
			if err != nil {
				return true, expression, err
			}
			function.subtype = returns
			expression.Type = function
			expression.Go.Write(c.Token())
			return true, expression, nil
		}
	}

	return
}

//Operation does nothing.
func (Function) Operation(c *compiler.Compiler, a, b compiler.Expression, symbol string) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	return
}

//Cast does nothing.
func (Function) Cast(c *compiler.Compiler, from compiler.Expression, to compiler.Type) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	return c.CastingError(from, to)
}

//Equals returns true if the other type is equal to this type.
func (Function) Equals(other compiler.Type) bool {
	_, ok := other.(Function)
	return ok
}

//Native returns this type's native token.
func (Function) Native(c *compiler.Compiler) (token compiler.Token) {
	if c.Target == target.Go {
		return compiler.Token("func(ctx I.Context)")
	}
	return
}

//Zero returns this type's zero expression.
func (Function) Zero(c *compiler.Compiler) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = Function{}

	expression.Go.WriteString(`func(ctx I.Context) {}`)

	return
}

//Copy returns a copy of the nothing type.
func (Function) Copy(c *compiler.Compiler, item compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()
	expression.Type = Function{}

	expression.Go.WriteB(item.Go)

	return
}

//Call calls this function.
func (function Function) Call(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if len(args) != len(function.Concept.Arguments) {
		return expression, c.NewError("Wrong number of argumetns, expecting ", len(function.Concept.Arguments), "but got ", len(args))
	}

	if !compiler.Defined(function.subtype) {
		return expression, c.NewError("Cannot call ", this.Type.String(c), " in expression context.")
	}

	expression.Type = function.subtype

	fmt.Fprintf(&expression.Go, `%v(ctx, %v)`, this.Go, compiler.Arguments(args))

	return
}

//Run runs this function.
func (Function) Run(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) error {
	c.Indent()
	c.Go.WriteB(this.Go)
	c.Go.WriteString(`(ctx)`)
	return nil
}

//Length returns the size/length/count of this type.
func (Function) Length(c *compiler.Compiler, this compiler.Expression) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = Integer{}
	expression.Go.WriteString(`I.NewInteger(0)`)
	return expression
}

//Index a value of this type with the specified indicies.
func (function Function) Index(c *compiler.Compiler, this compiler.Expression, indices ...compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()
	return expression, c.NewError("function cannot be indexed")
}

//Modify a value of this type with the specified indicies.
func (function Function) Modify(c *compiler.Compiler, this compiler.Expression, modification compiler.Expression, indices ...compiler.Expression) error {
	return c.NewError("function cannot be modified")
}

//Specify this type with the provided args.
func (function Function) Specify(c *compiler.Compiler, args ...compiler.Expression) (compiler.Type, error) {
	return nil, c.NewError("function cannot be specified")
}

//With should return this type containing the provided subtype.
func (function Function) With(c *compiler.Compiler, subtype compiler.Type) compiler.Type {
	function.subtype = subtype
	return function
}
