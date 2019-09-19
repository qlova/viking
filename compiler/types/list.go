package types

import (
	"fmt"

	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/target"
)

//List is an 'i' list.
type List struct {
	size    compiler.Expression //preintialised size.
	subtype compiler.Type
}

//Length returns the size/length/count of this type.
func (list List) Length(c *compiler.Compiler, this compiler.Expression) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = Integer{}
	expression.Go.WriteString(`I.NewInteger(int64(len(`)
	expression.Go.WriteB(this.Go)
	expression.Go.WriteString(`)))`)
	return expression
}

//Subtype returns the subtype.
func (list List) Subtype() compiler.Type {
	return list.subtype
}

var _ = compiler.RegisterType(List{})

//Name returns the name of this type.
func (List) Name() compiler.String {
	return compiler.String{
		compiler.English: `list`,
	}
}

func (list List) String(c *compiler.Compiler) string {
	var name = Array{}.Name()[c.Language]
	var subtype string
	if list.subtype != nil {
		subtype = "." + list.subtype.String(c)
	}
	return name + subtype
}

//Expression does nothing.
func (List) Expression(c *compiler.Compiler) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	return
}

//Operation does nothing.
func (List) Operation(c *compiler.Compiler, a, b compiler.Expression, symbol string) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	return
}

//Cast does nothing.
func (list List) Cast(c *compiler.Compiler, from compiler.Expression, to compiler.Type) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if sequence, ok := from.Type.(compiler.Sequence); ok {
		list.subtype = sequence.Subtype()

		expression.Go.WriteB(from.Go)

		expression.Type = list
		return expression, nil
	}

	return c.CastingError(from, to)
}

//Equals returns true if the other type is equal to this type.
func (list List) Equals(other compiler.Type) bool {
	a, ok := other.(List)

	if ok && list.Subtype() != nil {
		ok = list.Subtype().Equals(a.Subtype())
	}

	return ok
}

//Native returns this type's native token.
func (list List) Native(c *compiler.Compiler) (token compiler.Token) {
	var subtype string
	if list.Subtype != nil {
		subtype = list.Subtype().Native(c).String()
	}
	if c.Target == target.Go {
		return compiler.Token(fmt.Sprint("[]", subtype))
	}
	return
}

//Zero returns this type's zero expression.
func (list List) Zero(c *compiler.Compiler) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = list

	if list.size.Type != nil {
		expression.Go.WriteString("make(")
		expression.Go.Write(list.Native(c))
		expression.Go.WriteString(",int(")
		expression.Go.WriteB(list.size.Go)
		expression.Go.WriteString(".Int64())+1)")
		return
	}

	fmt.Fprintf(&expression.Go, "make(%v, 1)", list.Native(c))

	return
}

//Copy returns a copy of the nothing type.
func (list List) Copy(c *compiler.Compiler, item compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()
	expression.Type = list

	expression.Go.WriteString(`func(list `)
	expression.Go.Write(list.Native(c))
	expression.Go.WriteString(`) `)
	expression.Go.Write(list.Native(c))
	expression.Go.WriteString(` { var clone = make(`)
	expression.Go.Write(list.Native(c))
	expression.Go.WriteString(`, len(list)); copy(clone, list); return clone }(`)
	expression.Go.WriteB(item.Go)
	expression.Go.WriteString(`)`)

	return
}

//Index a value of this type with the specified indicies.
func (list List) Index(c *compiler.Compiler, this compiler.Expression, indices ...compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if len(indices) != 1 {
		return expression, c.NewError("array takes 1 integer index")
	}
	var index = indices[0]

	if !index.Equals(Integer{}) {
		return expression, c.NewError("list takes 1 integer offset")
	}

	expression.Type = list.Subtype()
	expression.Go.WriteB(this.Go)
	expression.Go.WriteString(`[I.IndexList(`)
	expression.Go.WriteB(index.Go)
	expression.Go.WriteString(`,len(`)
	expression.Go.WriteB(this.Go)
	expression.Go.WriteString(`))]`)

	return expression, nil
}

//Modify a value of this type with the specified indicies.
func (list List) Modify(c *compiler.Compiler, this compiler.Expression, modification compiler.Expression, indices ...compiler.Expression) error {
	if len(indices) != 1 {
		return c.NewError("array takes 1 integer offset")
	}

	if !modification.Equals(list.Subtype()) {
		return c.NewError("cannot value of type ", modification.String(c), "to array of type ", list.String(c))
	}

	var index = indices[0]

	if !index.Equals(Integer{}) {

		if index.Equals(Sequencer{}) {
			if index.Type.(Sequencer).Plus {
				c.Indent()
				fmt.Fprintf(&c.Go, "func(list *%v) {*list = append(*list, %v)}(&%v)",
					list.Native(c),
					modification.Go,
					this.Go,
				)
				return nil
			}
		}

		return c.NewError("list takes 1 integer offset")
	}

	c.Indent()
	fmt.Fprintf(&c.Go, "func(list %v) { list[I.IndexList(%v, len(list))] = %v }(%v)",
		list.Native(c),
		index.Go,
		modification.Go,
		this.Go,
	)
	return nil
}

//Specify this type with the provided args.
func (list List) Specify(c *compiler.Compiler, args ...compiler.Expression) (compiler.Type, error) {
	if len(args) != 1 {
		return list, nil
	}

	var argument = args[0]

	if !argument.Equals(Integer{}) {
		return nil, c.NewError("array takes 1 integer, size argument")
	}

	list.size = argument

	return list, nil
}

//With should return this type containing the provided subtype.
func (list List) With(c *compiler.Compiler, subtype compiler.Type) compiler.Type {
	list.subtype = subtype
	return list
}
