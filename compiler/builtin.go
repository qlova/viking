package compiler

//Builtins is a slice of all builtin functions.
var Builtins []Builtin

//RegisterBuiltin registers and returns a new builtin.
func RegisterBuiltin(builtin Builtin) Builtin {
	Builtins = append(Builtins, builtin)
	return builtin
}

//Builtin is a builtin concept.
type Builtin interface {
	Runnable

	Name() String
}

//IsBuiltin returns true if the builtin exists.
func IsBuiltin(check Token) bool {
	for _, builtin := range Builtins {
		if check.Is(builtin.Name()[English]) {
			return true
		}
	}
	return false
}
