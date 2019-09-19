package compiler

import (
	"fmt"
	"strconv"

	"github.com/qlova/viking/compiler/target"
)

//Sequence is a varadic, immutable collection type.
type Sequence struct {
	Size    int
	subtype Type

	Items []Expression
}

var _ = Collection(Sequence{})
var _ = RegisterType(Sequence{})

var SequenceLength = func(c *Compiler, this Expression) (expression Expression) {
	panic("unitialised sequence length function")
}

//Subtype returns the subtype.
func (sequence Sequence) Subtype() Type {
	return sequence.subtype
}

//Length returns the size/length/count of this type.
func (sequence Sequence) Length(c *Compiler, this Expression) (expression Expression) {
	return SequenceLength(c, this)
}

//Name returns the name of this type.
func (Sequence) Name() String {
	return String{
		English: `sequence`,
	}
}

func (sequence Sequence) String(c *Compiler) string {
	var name = Sequence{}.Name()[c.Language]
	var subtype string
	if sequence.subtype != nil {
		subtype = "." + sequence.subtype.String(c)
	}
	return fmt.Sprint(name, "[", sequence.Size, "]", subtype)
}

//Expression does nothing.
func (sequence Sequence) Expression(c *Compiler) (ok bool, expression Expression, err error) {
	expression = c.NewExpression()

	if c.Token().Is("[") {

		var first, err = c.ScanExpression()
		if err != nil {
			return true, expression, err
		}

		var items = []Expression{first}

		sequence.subtype = first.Type

		expression.Go.Write(sequence.Native(c))
		expression.Go.WriteString(`{`)
		expression.Go.WriteB(first.Go)

		var count = 1

		for c.ScanIf(',') {
			count++

			var next, err = c.ScanExpression()
			if err != nil {
				fmt.Println(err)
				return true, expression, err
			}

			if !next.Equals(first.Type) {
				return true, expression, c.NewError("elements in a sequence must share the same type")
			}

			items = append(items, next)

			expression.Go.WriteString(`,`)
			expression.Go.WriteB(next.Go)
		}

		if !c.ScanIf(']') {
			return true, expression, c.NewError("expecting ]")
		}

		expression.Go.WriteString(`}`)

		sequence.Size = count
		sequence.Items = items
		expression.Type = sequence

		return true, expression, err
	}

	return
}

//Operation does nothing.
func (sequence Sequence) Operation(c *Compiler, a, b Expression, symbol string) (ok bool, expression Expression, err error) {
	expression = c.NewExpression()

	if symbol == "+" {
		if !a.Equals(b.Type) {
			return true, Expression{}, c.NewError("cannot add array and " + b.String(c))
		}

		sequence.Size = a.Type.(Sequence).Size + b.Type.(Sequence).Size
		expression.Type = sequence

		expression.Go.Write(c.GoTypeOf(expression.Type))
		expression.Go.WriteString(`{`)
		expression.Go.Write(a.Go.Bytes())
		expression.Go.WriteString(`[0]`)
		for i := 1; i < a.Type.(Sequence).Size; i++ {
			expression.Go.WriteString(`,`)
			expression.Go.Write(a.Go.Bytes())
			expression.Go.WriteString(`[` + strconv.Itoa(i) + `]`)
		}
		for i := 0; i < b.Type.(Sequence).Size; i++ {
			expression.Go.WriteString(`,`)
			expression.Go.Write(b.Go.Bytes())
			expression.Go.WriteString(`[` + strconv.Itoa(i) + `]`)
		}
		expression.Go.WriteString(`}`)

		return true, expression, nil
	}

	return
}

//Cast does nothing.
func (Sequence) Cast(c *Compiler, from Expression, to Type) (expression Expression, err error) {
	return c.CastingError(from, to)
}

//Equals returns true if the other type is equal to this type.
func (sequence Sequence) Equals(other Type) bool {
	a, ok := other.(Sequence)

	if ok && sequence.Subtype() != nil {
		ok = sequence.Subtype().Equals(a.Subtype())
	}

	return ok
}

//Native returns this type's native token.
func (sequence Sequence) Native(c *Compiler) (token Token) {
	if c.Target == target.Go {
		var subtype Token
		if sequence.Subtype() != nil {
			subtype = sequence.Subtype().Native(c)
		}
		return Token(fmt.Sprint("[]", subtype.String()))
	}
	return
}

//Zero returns this type's zero expression.
func (sequence Sequence) Zero(c *Compiler) (expression Expression) {
	expression = c.NewExpression()
	expression.Type = sequence

	expression.Go.Write(sequence.Native(c))
	expression.Go.WriteString(`{}`)
	return
}

//Copy returns a copy of the nothing type.
func (Sequence) Copy(c *Compiler, item Expression) (expression Expression, err error) {
	expression = c.NewExpression()
	expression.Type = Sequence{}

	expression.Go.WriteB(item.Go)
	return
}

//Index a value of this type with the specified indicies.
func (arguments Sequence) Index(c *Compiler, this Expression, indices ...Expression) (expression Expression, err error) {
	expression = c.NewExpression()

	if len(indices) != 1 {
		return expression, c.NewError("arguments takes 1 integer index")
	}
	var index = indices[0]

	expression.Type = arguments.Subtype()
	expression.Go.WriteB(this.Go)
	expression.Go.WriteString(`[I.IndexList(`)
	expression.Go.WriteB(index.Go)
	expression.Go.WriteString(`,len(`)
	expression.Go.WriteB(this.Go)
	expression.Go.WriteString(`))]`)

	return expression, nil
}

//Modify a value of this type with the specified indicies.
func (arguments Sequence) Modify(c *Compiler, this Expression, modification Expression, indices ...Expression) error {
	return c.NewError("arguments cannot be modified")
}

//Specify this type with the provided args.
func (arguments Sequence) Specify(c *Compiler, args ...Expression) (Type, error) {
	if len(args) != 0 {
		return nil, c.NewError("arguments doesn not take any arguments")
	}

	return arguments, nil
}

//With should return this type containing the provided subtype.
func (arguments Sequence) With(c *Compiler, subtype Type) Type {
	arguments.subtype = subtype
	return arguments
}
