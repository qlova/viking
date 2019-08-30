package compiler

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
)

var Trace = os.Getenv("TRACE") != ""

type Error struct {
	Formatted string
	Message   string
}

func (err Error) Error() string {
	return err.Formatted
}

func (compiler *Compiler) NewError(msg string) error {
	var wdir string
	if runtime.GOOS != "js" {
		wdir, _ = os.Getwd()
	}

	var rpath, _ = filepath.Rel(wdir, compiler.Directory)

	if len(rpath) > len(compiler.Directory) {
		rpath = compiler.Directory
	}

	var RestOfTheLine, _ = compiler.Reader.ReadString('\n')

	var formatted = fmt.Sprint(rpath, compiler.Filename, ":",
		compiler.LineNumber, ": ", string(compiler.Line), RestOfTheLine, "\n",
		strings.Repeat(" ", compiler.Column+2+len(rpath)+len(compiler.Filename)), "^\n", msg)

	if Trace {
		var stacktrace = debug.Stack()
		var reader = bufio.NewReader(bytes.NewBuffer(stacktrace))

		const count = 7
		for i := 0; i < count; i++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				formatted += string(stacktrace)
				break
			}
			if i == count-1 {
				formatted += "\n(" + strings.TrimSpace(line) + ")\n"
			}
		}

	}

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
	return compiler.NewError("expecting " + string(symbol) + " but found " + compiler.Scan().String())
}
