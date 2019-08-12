package compiler

import (
	"fmt"
	"strings"
)

type Error struct {
	Formatted string
	Message   string
}

func (err Error) Error() string {
	return err.Formatted
}

func (compiler *Compiler) NewError(msg string) error {
	var RestOfTheLine, _ = compiler.Reader.ReadString('\n')
	if compiler.NextToken != nil {
		RestOfTheLine = compiler.NextToken.String() + RestOfTheLine
	}
	var formatted = fmt.Sprint(compiler.LineNumber, ": ", string(compiler.Line), RestOfTheLine, "\n", strings.Repeat(" ", compiler.Column+2), "^\n", msg)
	return Error{formatted, msg}
}

//Unimplemented is an error describing that the component is unimplemented.
func (compiler *Compiler) Unimplemented(component []byte) error {
	return compiler.NewError("unimplemented " + string(component))
}

//Undefined is an error describing that the name is undefined.
func (compiler *Compiler) Undefined(name []byte) error {
	return compiler.NewError("undefined " + string(name))
}

//Expecting returns an error in the form "expecting [token]"
func (compiler *Compiler) Expecting(symbol byte) error {
	return compiler.NewError("expecting " + string(symbol))
}
