package types

import "github.com/qlova/viking/compiler"

//Undefined is an 'i' undefined.
type Undefined struct {
	compiler.Nothing
}

var _ = compiler.RegisterType(Undefined{})

//Name returns the name of this type.
func (Undefined) Name() compiler.String {
	return compiler.String{
		compiler.English: `undefined`,
	}
}

func (Undefined) String(c *compiler.Compiler) string {
	return Undefined{}.Name()[c.Language]
}

//Equals returns true if the other type is equal to this type.
func (Undefined) Equals(other compiler.Type) bool {
	_, ok := other.(Undefined)
	return ok
}
