package compiler

//List is a dynamic-length sequence of values.
var List = Type{Name: "list"}

//IndexList indexes the list with the specified index.
func (compiler *Compiler) IndexList(list, index Expression) (Expression, error) {
	return compiler.IndexArray(list, index)
}

//ModifyList indexes the list with the specified index.
func (compiler *Compiler) ModifyList(list Token) error {
	return compiler.ModifyArray(list)
}
