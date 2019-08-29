package compiler

//Flag is a psuedotype.
var Flag = Type{Name: "flag"}

//SetFlag sets a new flag.
func (compiler *Context) SetFlag(name []byte) {
	compiler.Scope[len(compiler.Scope)-1].Table["flag_"+string(name)] = Flag
}

//GetFlag returns the flag with the given name.
func (compiler *Context) GetFlag(name Token) bool {
	if len(compiler.Scope) <= 0 {
		return false
	}
	for i := len(compiler.Scope) - 1; i >= 0; i-- {
		if _, ok := compiler.Scope[i].Table["flag_"+name.String()]; ok {
			return true
		}
	}
	return false
}
