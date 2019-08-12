package compiler

import (
	"bytes"
	"errors"
)

//Builtins is a list of all builtin functions.
var Builtins = []string{"print", "out", "in", "copy", "throw"}

//Throw builtin.
const Throw = `func (ctx *Context) Throw(code int, msg string) {
	ctx.Error = Error{
		Code: code,
		Message: msg,
	}
}
`

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
	var expression = compiler.NewExpression()

	switch builtin.String() {
	case "copy":
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

		expression.Type = argument.Type

		if argument.Is(List) {
			expression.Go.WriteString("append(")
			expression.Go.Write(argument.Go.Bytes())
			expression.Go.WriteString("[:0:0]")
			expression.Go.WriteString(",")
			expression.Go.Write(argument.Go.Bytes())
			expression.Go.WriteString("...)")
			return expression, nil
		}

		return Expression{}, Unimplemented(s("copy for " + argument.Type.Name))
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
			expression.Go.WriteString("in_symbol(")
			expression.Go.Write(argument.Go.Bytes())
			expression.Go.WriteString(")")

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
func (compiler *Compiler) CompileBuiltin(builtin Token) error {
	if builtin.Is("print") || builtin.Is("out") {

		if !compiler.ScanIf('(') {
			return compiler.Unexpected()
		}

		var Arguments []Expression

		var first, err = compiler.ScanExpression()
		if err != nil {
			return err
		}
		Arguments = append(Arguments, first)

		for compiler.ScanIf(',') {
			var arg, err = compiler.ScanExpression()
			if err != nil {
				return err
			}
			Arguments = append(Arguments, arg)
		}

		if !compiler.ScanIf(')') {
			return compiler.Expecting(')')
		}

		compiler.Import("fmt")

		if bytes.Equal(builtin, []byte("out")) {
			compiler.Go.Write([]byte("fmt.Print("))
		} else {
			compiler.Go.Write([]byte("fmt.Println("))
		}

		for i, argument := range Arguments {
			if argument.Type.Equals(Symbol) {
				compiler.Go.Write([]byte("string("))
				compiler.Go.Write(argument.Go.Bytes())
				compiler.Go.Write([]byte(")"))
			} else {
				compiler.Go.Write(argument.Go.Bytes())
			}
			if i < len(Arguments)-1 {
				compiler.Go.WriteString(",")
			}
		}

		compiler.Go.Write([]byte(")\n"))

		err = compiler.ScanLine()
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New(string(builtin) + " is not a builtin")
}
