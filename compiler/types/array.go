package types

import (
	"fmt"
	"strconv"

	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/target"
)

//Array is an 'i' array.
type Array struct {
	Size    int
	subtype compiler.Type
}

//Length returns the size/length/count of this type.
func (array Array) Length(c *compiler.Compiler, this compiler.Expression) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = Integer{}
	expression.Go.WriteString(`I.NewInteger(int64(`)
	expression.Go.WriteString(strconv.Itoa(array.Size))
	expression.Go.WriteString(`))`)
	return expression
}

//Subtype returns the subtype.
func (array Array) Subtype() compiler.Type {
	return array.subtype
}

var _ = compiler.RegisterType(Array{})

//Name returns the name of this type.
func (Array) Name() compiler.String {
	return compiler.String{
		compiler.English: `array`,
	}
}

func (array Array) String(c *compiler.Compiler) string {
	var name = Array{}.Name()[c.Language]
	var subtype string
	if array.subtype != nil {
		subtype = "." + array.subtype.String(c)
	}
	return name + fmt.Sprint("[", array.Size, "]", subtype)
}

//Expression does nothing.
func (Array) Expression(c *compiler.Compiler) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	return
}

//Operation does nothing.
func (Array) Operation(c *compiler.Compiler, a, b compiler.Expression, symbol string) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	return
}

//Cast does nothing.
func (array Array) Cast(c *compiler.Compiler, from compiler.Expression, to compiler.Type) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if sequence, ok := from.Type.(compiler.Sequence); ok {
		if array.Size == 0 {
			array.Size = sequence.Size
		}

		if array.Size != sequence.Size {
			return from, c.NewError("array size must be the same as the sequence size")
		}

		array.subtype = sequence.Subtype()

		var buffer = from.Go.Bytes()
		if buffer[0] == '[' {
			buffer[0] = '.'
			expression.Go.WriteString(`[..`)
			defer func() {
				buffer[0] = '['
			}()
		} else {
			expression.Go.WriteString(`func() (array `)
			expression.Go.Write(array.Native(c))
			expression.Go.WriteString(`) { copy(array[:], `)
			expression.Go.WriteB(from.Go)
			expression.Go.WriteString(`); return; }`)
		}

		expression.Go.Write(buffer)

		expression.Type = array
		return expression, nil
	}

	return c.CastingError(from, to)
}

//Equals returns true if the other type is equal to this type.
func (array Array) Equals(other compiler.Type) bool {
	a, ok := other.(Array)

	if ok && array.Subtype() != nil {
		ok = array.Subtype().Equals(a.Subtype())
	}

	return ok
}

//Native returns this type's native token.
func (array Array) Native(c *compiler.Compiler) (token compiler.Token) {
	var subtype string
	if array.Subtype() != nil {
		subtype = array.Subtype().Native(c).String()
	}
	if c.Target == target.Go {
		return compiler.Token(fmt.Sprint("[", array.Size, "]", subtype))
	}
	return
}

//Zero returns this type's zero expression.
func (array Array) Zero(c *compiler.Compiler) (expression compiler.Expression) {
	expression = c.NewExpression()
	expression.Type = array

	expression.Go.Write(array.Native(c))
	expression.Go.WriteString("{}")

	return
}

//Copy returns a copy of the nothing type.
func (array Array) Copy(c *compiler.Compiler, item compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()
	expression.Type = array

	expression.Go.WriteB(item.Go)

	return
}

//Index a value of this type with the specified indicies.
func (array Array) Index(c *compiler.Compiler, this compiler.Expression, indices ...compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if len(indices) != 1 {
		return expression, c.NewError("array takes 1 integer offset")
	}
	var index = indices[0]

	if !index.Equals(Integer{}) {
		return expression, c.NewError("array takes 1 integer offset")
	}

	expression.Type = array.Subtype()
	expression.Go.WriteB(this.Go)
	expression.Go.WriteString(`[I.IndexArray(`)
	expression.Go.WriteB(index.Go)
	expression.Go.WriteString(`,len(`)
	expression.Go.WriteB(this.Go)
	expression.Go.WriteString(`))]`)

	return expression, nil
}

//Modify a value of this type with the specified indicies.
func (array Array) Modify(c *compiler.Compiler, this compiler.Expression, modification compiler.Expression, indices ...compiler.Expression) error {
	if len(indices) != 1 {
		return c.NewError("array takes 1 integer offset")
	}

	if !modification.Equals(array.Subtype()) {
		return c.NewError("cannot value of type ", modification.String(c), "to array of type ", array.String(c))
	}

	var index = indices[0]

	if !index.Equals(Integer{}) {
		return c.NewError("array takes 1 integer offset")
	}

	c.Go.WriteB(this.Go)
	c.Go.WriteString(`[I.IndexArray(`)
	c.Go.WriteB(index.Go)
	c.Go.WriteString(`,len(`)
	c.Go.WriteB(this.Go)
	c.Go.WriteString(`))] = `)
	c.Go.WriteB(modification.Go)

	return nil
}

//Specify this type with the provided args.
func (array Array) Specify(c *compiler.Compiler, args ...compiler.Expression) (compiler.Type, error) {
	if len(args) != 1 {
		array.Size = 0
		return array, nil
	}

	var argument = args[0]

	var integer = string(argument.Go.Bytes())

	size, err := strconv.Atoi(integer[len("I.NewInteger(") : len(integer)-1])
	if err != nil {
		return nil, c.NewError("array takes 1 'constant' integer size argument")
	}

	array.Size = size

	return array, nil
}

//With should return this type containing the provided subtype.
func (array Array) With(c *compiler.Compiler, subtype compiler.Type) compiler.Type {
	array.subtype = subtype
	return array
}
