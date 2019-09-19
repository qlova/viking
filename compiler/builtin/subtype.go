package builtin

import (
	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/types"
)

//Subtype is an interface around getting the subtype of metatypes.
type Subtype struct {
	compiler.Nothing
}

var _ = compiler.RegisterBuiltin(Subtype{})

//Name returns in's name.
func (Subtype) Name() compiler.String {
	return compiler.String{
		compiler.English: `subtype`,
	}
}

//Run does nothing.
func (Subtype) Run(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) (err error) {
	return c.NewError("subtype cannot be called as statement")
}

//Call calls subtype.
func (Subtype) Call(c *compiler.Compiler, this compiler.Expression, args ...compiler.Expression) (expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if len(args) != 1 {
		return expression, c.NewError("subtype takes one dynamic argument")
	}

	var collection, ok = args[0].Type.(compiler.Collection)

	var subtype compiler.Type = types.Undefined{}
	if ok && collection.Subtype() != nil {
		subtype = collection.Subtype()
	}

	expression.Type = types.Metatype{
		Type: subtype,
	}

	return expression, nil
}
