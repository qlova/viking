package compiler

//Flag is a psuedotype.
type Flag struct {
	Nothing
}

//Name returns the name of this flag.
func (flag Flag) Name() String {
	return String{
		English: `flag`,
	}
}

//SetFlag sets a new flag.
func (compiler *Context) SetFlag(name []byte) {
	compiler.Scope[len(compiler.Scope)-1].Table["flag_"+string(name)] = Flag{}
}

//Flag returns the flag with the given name.
func (compiler *Context) Flag(name Token) bool {
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

//FlagIsCurrent returns the flag with the given name inside the current scope.
func (compiler *Context) FlagIsCurrent(name Token) bool {
	if len(compiler.Scope) <= 0 {
		return false
	}
	if _, ok := compiler.Scope[len(compiler.Scope)-1].Table["flag_"+name.String()]; ok {
		return true
	}
	return false
}
