package compiler

//Alias is a constant level alias.
type Alias Cache

//UnpackAlias uses an alias.
func (compiler *Compiler) UnpackAlias(alias Alias) {
	var cache = Cache(alias)
	compiler.PushReader(&cache.Buffer)
}

//DefineAlias defines a new alias.
func (compiler *Compiler) DefineAlias(name Token) {
	compiler.Aliases[name.String()] = Alias(compiler.CacheLine())
}
