package compiler

import (
	"errors"
	"fmt"
	"strings"
)

//Error is a compiler error.
type Error struct {
	compiler *Compiler
	err      error
}

func (err Error) Error() string {
	var compiler = err.compiler
	var RestOfTheLine, _ = compiler.Reader.ReadString('\n')
	if compiler.NextToken != nil {
		RestOfTheLine = compiler.NextToken.String() + RestOfTheLine
	}
	return fmt.Sprint(compiler.LineNumber, ": ", string(compiler.Line), RestOfTheLine, "\n", strings.Repeat(" ", compiler.Column+3), "^\n", err.err.Error())
}

//Unimplemented is an error describing that the component is unimplemented.
func Unimplemented(component []byte) error {
	return errors.New("unimplemented " + string(component))
}

//Expecting returns an error in the form "expecting [token]"
func (compiler *Compiler) Expecting(symbol byte) error {
	return errors.New("expecting " + string(symbol))
}
