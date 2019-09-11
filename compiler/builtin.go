package compiler

import (
	"bytes"
)

//Builtins is a list of all builtin functions.
var Builtins = []string{"print", "out", "in", "copy", "throw"}

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

		return Expression{}, compiler.Unimplemented(s("copy for " + argument.Type.Name))
	case "in":
		compiler.Throws = true

		if !compiler.ScanIf('(') {
			return Expression{}, compiler.Expecting('(')
		}

		if compiler.ScanIf(')') {
			compiler.Import(Ilang)

			expression.Type = String
			expression.Go.WriteString("I.InSymbol(ctx, '\n')")
			return expression, nil
		}

		var argument, err = compiler.ScanExpression()
		if err != nil {
			return Expression{}, err
		}

		if !compiler.ScanIf(')') {
			return Expression{}, compiler.Expecting(')')
		}

		if argument.Equals(Symbol) {
			compiler.Import(Ilang)

			expression.Type = String
			expression.Go.WriteString("I.InSymbol(ctx, ")
			expression.Go.Write(argument.Go.Bytes())
			expression.Go.WriteString(")")
			return expression, nil
		}

		return Expression{}, compiler.NewError("invalid type " + expression.Type.Name + " passed to builtin")
	}
	return Expression{}, compiler.NewError("invalid builtin " + builtin.String())
}

//CompileBuiltin compiles a call to a builtin.
func (compiler *Compiler) CompileBuiltin(builtin Token) error {
	if builtin.Is("print") || builtin.Is("out") {

		if !compiler.ScanIf('(') {
			return compiler.Expecting('(')
		}

		if compiler.ScanIf(')') {
			compiler.Go.Write([]byte("fmt.Println()"))
			return nil
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

		compiler.Go.Write([]byte(")"))

		return nil
	}

	return compiler.NewError(string(builtin) + " is not a builtin")
}
