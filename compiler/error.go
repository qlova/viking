package compiler

import "fmt"
import "strings"

type Error struct {
	compiler *Compiler
	err      error
}

func (err Error) Error() string {
	var compiler = err.compiler
	return fmt.Sprint(compiler.LineNumber, ": ", string(compiler.Line), "\n", strings.Repeat(" ", compiler.Column+3), "^\n", err.err.Error())
}
