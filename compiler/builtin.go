package compiler

import "bytes"
import "errors"

//Builtins is a list of all builtin functions.
var Builtins = []string{"print", "write"}

//Builtin returns true if the builtin exists.
func Builtin(check []byte) bool {
	for _, builtin := range Builtins {
		if bytes.Equal([]byte(builtin), check) {
			return true
		}
	}
	return false
}

//CompileBuiltin compiles a call to a builtin.
func (compiler *Compiler) CompileBuiltin(builtin []byte) error {
	if bytes.Equal(builtin, []byte("print")) || bytes.Equal(builtin, []byte("write")) {

		if !compiler.ScanIf('(') {
			return compiler.Unexpected()
		}

		var expression, err = compiler.ScanExpression()
		if err != nil {
			return err
		}

		if !compiler.ScanIf(')') {
			return compiler.Unexpected()
		}

		compiler.Import("fmt")

		if bytes.Equal(builtin, []byte("write")) {
			compiler.Write([]byte("fmt.Print("))
		} else {
			compiler.Write([]byte("fmt.Println("))
		}

		if expression.Type.Equals(Symbol) {
			compiler.Write([]byte("string("))
			compiler.Write(expression.Bytes())
			compiler.Write([]byte(")"))
		} else {
			compiler.Write(expression.Bytes())
		}
		compiler.Write([]byte(")\n"))

		err = compiler.ScanLine()
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New(string(builtin) + " is not a builtin")
}
