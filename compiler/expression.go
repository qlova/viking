package compiler

import (
	"errors"
	"io"
	"os"
	"strconv"

	"github.com/qlova/viking/compiler/target"
)

//Expression is a type with content.
type Expression struct {
	Type
	target.Buffer
}

func (compiler *Compiler) NewExpression() Expression {
	var expression Expression
	if compiler.Go.Enabled {
		expression.Go.Enabled = true
	}
	return expression
}

//ScanExpression returns the next expression or an error.
func (compiler *Compiler) ScanExpression() (Expression, error) {
	var expression, err = compiler.scanExpression()
	if err != nil {
		return expression, err
	}

	return compiler.Shunt(expression, 0)
}

func (compiler *Compiler) scanExpression() (Expression, error) {
	return compiler.expression(compiler.Scan())
}

func (compiler *Compiler) expression(token Token) (Expression, error) {
	var expression = compiler.NewExpression()

	/*switch token {
		case "if", "for", "return", "break", "go":
			return Unimplemented
	}*/

	if token == nil {
		return Expression{}, io.EOF
	}

	//Ignore comments
	if len(token) > 2 && token[0] == '/' && token[1] == '/' {
		compiler.Go.Write(token)
		return expression, nil
	}

	//Alias expression.
	if alias, ok := compiler.Aliases[token.String()]; ok {
		compiler.UnpackAlias(alias)
		return compiler.expression(compiler.Scan())
	}

	//String expression.
	if token[0] == '"' {
		expression.Type = String
		expression.Go.Write(token)
		return expression, nil
	}

	//Symbol expresion.
	if token[0] == '\'' {
		expression.Type = Symbol
		expression.Go.Write(token)
		return expression, nil
	}

	//String expression.
	if token.Is("(") {
		var internal, err = compiler.ScanExpression()
		if err != nil {
			return internal, err
		}
		if !compiler.ScanIf(')') {
			return internal, compiler.Expecting(')')
		}
		expression.Type = internal.Type
		expression.Go.Write(token)
		expression.Go.Write(internal.Go.Bytes())
		expression.Go.WriteString(")")
		return expression, nil
	}

	//Binary number expression.
	if i, err := strconv.ParseInt(string(token), 2, 64); err == nil && len(token) > 0 && token[0] == '0' {
		expression.Type = Integer
		if Deterministic {
			compiler.Import(Ilang)
			expression.Go.WriteString("I.NewInteger(")
		}
		expression.Go.Write(s(strconv.Itoa(int(i))))
		if Deterministic {
			expression.Go.WriteString(")")
		}
		return expression, nil
	}

	//Integer expression.
	if _, err := strconv.Atoi(string(token)); err == nil {
		expression.Type = Integer
		if Deterministic {
			compiler.Import(Ilang)
			expression.Go.WriteString("I.NewInteger(")
		}
		expression.Go.Write(token)
		if Deterministic {
			expression.Go.WriteString(")")
		}
		return expression, nil
	}

	//Hexadecimal expression.
	if len(token) > 2 && token[0] == '0' && token[1] == 'x' {
		expression.Type = Integer
		if Deterministic {
			compiler.Import(Ilang)
			expression.Go.WriteString("I.NewInteger(")
		}
		expression.Go.Write(token)
		if Deterministic {
			expression.Go.WriteString(")")
		}
		return expression, nil
	}

	//Bit expression.
	if token.Is("true") || token.Is("false") {
		expression.Type = Bit
		expression.Go.Write(token)
		return expression, nil
	}

	//Not expression.
	if token.Is("!") {

		var boolean, err = compiler.scanExpression()
		if err != nil {
			return Expression{}, err
		}

		if !boolean.Equals(Bit) {
			return Expression{}, compiler.NewError("cannot apply not operator to value of type " + boolean.Type.Name)
		}

		expression.Type = Bit
		expression.Go.WriteString("(!")
		expression.Go.Write(boolean.Go.Bytes())
		expression.Go.WriteString(")")
		return expression, nil
	}

	//Array expression.
	if token.Is("[") {
		expression.Type = Array

		var first, err = compiler.ScanExpression()
		if err != nil {
			return Expression{}, err
		}

		expression.Type.Subtype = &first.Type

		var count = 1
		var buffer = compiler.NewExpression()
		buffer.Go.WriteByte('{')
		buffer.Go.Write(first.Go.Bytes())

		for compiler.ScanIf(',') {
			buffer.Go.WriteByte(',')

			var item, err = compiler.ScanExpression()
			if err != nil {
				return Expression{}, err
			}
			buffer.Go.Write(item.Go.Bytes())

			count++
		}

		if !compiler.ScanIf(']') {
			return Expression{}, compiler.Expecting(']')
		}
		buffer.Go.WriteByte('}')

		expression.Type.Size = count
		expression.Go.Write(GoTypeOf(expression.Type))
		expression.Go.Write(buffer.Go.Bytes())

		return expression, nil
	}

	//Length expression.
	if token.Is("#") {
		expression.Type = Integer

		var collection, err = compiler.scanExpression()
		if err != nil {
			return Expression{}, err
		}

		if collection.Is(List) || collection.Is(String) || collection.Is(Array) {

			if Deterministic {
				compiler.Import(Ilang)
				expression.Go.WriteString("I.NewInteger(")
			}
			expression.Go.WriteString("len(")
			expression.Go.Write(collection.Go.Bytes())
			expression.Go.WriteString(")")
			if Deterministic {
				expression.Go.WriteString(")")
			}

			return expression, nil
		}

		return Expression{}, compiler.NewError("cannot take the length of " + collection.Type.Name)
	}

	//Variable expression.
	if variable := compiler.GetVariable(token); Defined(variable) {
		expression.Type = variable
		expression.Go.Write(token)
		return expression, nil
	}

	//Function calls.
	if concept, ok := compiler.Concepts[token.String()]; ok {

		if !compiler.Peek().Is("(") {
			expression.Type = Function
			expression.Go.Write(token)
			return expression, nil
		}

		return concept.Call(compiler)
	}

	//Is this a builtin call?
	if Builtin(token) {
		return compiler.CallBuiltin(token)
	}

	//Prototype conversion.
	if T := compiler.GetPrototype(token); T.Defined() {
		if T.ScanExpression != nil {
			return T.ScanExpression(compiler)
		}
	}

	//Collections, arrays, lists etc.
	if T := compiler.GetType(token); Defined(T) {
		var next = compiler.Scan()

		if next.Is("(") {
			if compiler.ScanIf(')') {
				return compiler.Type(T)
			}
			return compiler.ScanCast(T)
		}

		if next.Is(".") {

			//This could be inline target code.
			if compiler.Peek().Is("if") {
				compiler.Scan()
				for {
					var name = compiler.ScanAndIgnoreNewLines()
					if t := target.FromString(name.String()); t.Valid() {
						var code = compiler.ScanAndIgnoreNewLines()
						if code[0] != '`' {
							return Expression{}, errors.New("expecting `[target code]`")
						}
						expression.Get(t).Write(code[1 : len(code)-1])
					}

					if name.Is("}") {
						break
					}
					if name == nil {
						return Expression{}, errors.New("if block wasn't closed")
					}
				}
				expression.Type = T
				return expression, nil
			}

			var collection = T

			var subtype = compiler.GetType(compiler.Scan())

			if Defined(subtype) {
				return compiler.Collection(collection, subtype)
			}
			return Expression{}, compiler.NewError("No such collection " + string(compiler.LastToken))
		}

		return Expression{}, compiler.Unimplemented(append(append(token, next...), compiler.Peek()...))
	}

	if P, err := compiler.GetPackage(token); err == nil {
		return P.Expression(compiler)
	} else if !os.IsNotExist(err) {
		return Expression{}, err
	}

	return Expression{}, compiler.Undefined(token)
}
