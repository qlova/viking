package compiler

import "bytes"
import "errors"

//Builtins is a list of all builtin functions.
var Builtins = []string{"print", "write", "in"}

//Builtin returns true if the builtin exists.
func Builtin(check Token) bool {
	for _, builtin := range Builtins {
		if check.Is(builtin) {
			return true
		}
	}
	return false
}

//CallBuiltin calls a builtin function and returns the resulting expression.
func (compiler *Compiler) CallBuiltin(builtin Token) (Expression, error) {
	var expression Expression

	switch builtin.String() {
	case "in":
		if !compiler.ScanIf('(') {
			return Expression{}, compiler.Expecting('(')
		}

		var argument, err = compiler.ScanExpression()
		if err != nil {
			return Expression{}, err
		}

		if !compiler.ScanIf(')') {
			return Expression{}, compiler.Expecting(')')
		}

		if argument.Equals(Symbol) {
			expression.Type = String
			expression.WriteString("in_symbol(")
			expression.Write(argument.Bytes())
			expression.WriteString(")")

			compiler.Import("os")
			compiler.Import("bufio")
			compiler.Require("var std_in = bufio.NewReader(os.Stdin)\n")
			compiler.Require(`func in_symbol(r rune) string {
	s, err := std_in.ReadString(byte(r))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return s[:len(s)-1]
}
`)
			return expression, nil
		}

		return Expression{}, errors.New("invalid type " + expression.Type.Name + " passed to builtin")
	}
	return Expression{}, errors.New("invalid builtin " + builtin.String())
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
