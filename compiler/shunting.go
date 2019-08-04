package compiler

import (
	"errors"
)

//Precedence returns i's precedence for the specified symbol.
func Precedence(symbol []byte) int {
	if symbol == nil {
		return -1
	}
	switch string(symbol) {
	case ",", ")", "]", "\n", "", "}":
		return -1

	default:
		return 0
	}
}

//Shunt shunts an expression with the next part of the expression. Emplying operators.
func (compiler *Compiler) Shunt(e Expression, precedence int) (result Expression, err error) {
	result = e

	//shunting:
	for peek := compiler.Peek(); Precedence(peek) >= precedence; {

		if Precedence(compiler.Peek()) == -1 {
			break
		}

		precedence := Precedence(peek)
		symbol := peek
		if compiler.Scan() == nil {
			break
		}

		rhs, err := compiler.ScanExpression()
		if err != nil {
			return result, err
		}

		peek = compiler.Peek()
		for Precedence(peek) > precedence {
			rhs, err = compiler.Shunt(rhs, Precedence(peek))
			if err != nil {
				return result, err
			}

			peek = compiler.Peek()
		}

		if result.Name == "array" {
			if equal(symbol, "[") {
				result, err = compiler.IndexArray(result, rhs)
				if err != nil {
					return result, err
				}
				continue
			}
		}

		if result.Equals(String) || result.Equals(Integer) {
			if equal(symbol, "+") {
				result, err = compiler.BasicAdd(result, rhs)
				if err != nil {
					return result, err
				}
				continue
			}
		}

		//Lets do the shunting!

		/*for i := range c.Shunts {
			if result = c.Shunts[i](c, symbol, t, rhs); result != nil {
				continue shunting
			}
		}*/

		return Expression{}, errors.New("Operator " + string(symbol) + " does not apply to " + result.Name)
	}
	return result, nil
}
