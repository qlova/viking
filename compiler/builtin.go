package compiler

import "bytes"
import "errors"

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
	var expression Expression

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
			expression.WriteString("append(")
			expression.Write(argument.Bytes())
			expression.WriteString("[:0:0]")
			expression.WriteString(",")
			expression.Write(argument.Bytes())
			expression.WriteString("...)")
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
			compiler.Write([]byte("fmt.Print("))
		} else {
			compiler.Write([]byte("fmt.Println("))
		}

		for i, argument := range Arguments {
			if argument.Type.Equals(Symbol) {
				compiler.Write([]byte("string("))
				compiler.Write(argument.Bytes())
				compiler.Write([]byte(")"))
			} else {
				compiler.Write(argument.Bytes())
			}
			if i < len(Arguments)-1 {
				compiler.WriteString(",")
			}
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
