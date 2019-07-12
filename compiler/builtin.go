package compiler

import "bytes"
import "errors"

var Builtins = []string{"print"}

func Builtin(check []byte) bool {
	for _, builtin := range Builtins {
		if bytes.Equal([]byte(builtin), check) {
			return true
		}
	}
	return false
}

func (compiler *Compiler) CompileBuiltin(builtin []byte) error {
	if bytes.Equal(builtin, []byte("print")) {

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

		compiler.Write([]byte("fmt.Println("))

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
