package compiler

//Precedence returns i's precedence for the specified symbol.
func Precedence(symbol []byte) int {
	if symbol == nil {
		return -1
	}
	switch string(symbol) {
	case ",", ")", "]", "\n", "", "}", "in", ":", ";", "to":
		return -1

	case "|":
		return 0

	case "&":
		return 1

	case "=", "<", ">", "!":
		return 2

	case "+", "-":
		return 3

	case "*", "/", `%`:
		return 4

	case "^":
		return 5

	case "(", "[":
		return 6

	case ".":
		return 7

	default:
		return -2
	}
}

//Shunt shunts an expression with the next part of the expression. Emplying operators.
func (compiler *Compiler) Shunt(expression Expression, precedence int) (result Expression, err error) {
	result = expression

	//shunting:
	for peek := compiler.Peek(); Precedence(peek) >= precedence; {

		if Precedence(compiler.Peek()) <= -1 {
			break
		}

		precedence := Precedence(peek)
		symbol := peek
		if compiler.Scan() == nil {
			break
		}

		rhs, err := compiler.scanExpression()
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

		if result.Is(Array) || result.Is(List) {
			if equal(symbol, "[") {
				result, err = compiler.IndexArray(result, rhs)
				if err != nil {
					return result, err
				}
				continue
			}
		}

		if expression.Is(Array) {
			if equal(symbol, "+") {
				result, err = compiler.ConcatArray(expression, rhs)
				if err != nil {
					return result, err
				}
				continue
			}
		}

		if result.Equals(String) {
			switch s := symbol.String(); s {
			case "+":
				result, err = compiler.BasicConcat(result, rhs)
				if err != nil {
					return result, err
				}
				continue
			}
		}

		if result.Equals(Bit) {
			switch s := symbol.String(); s {
			case "-":
				result, err = compiler.BasicXOr(result, rhs)
			case "&":
				result, err = compiler.BasicAnd(result, rhs)
			case "|":
				result, err = compiler.BasicOr(result, rhs)
			}
			if err != nil {
				return result, err
			}
			continue
		}

		switch s := symbol.String(); s {
		case "=":
			result, err = compiler.BasicEquals(result, rhs)
			if err != nil {
				return result, err
			}
			continue
		}

		if result.Equals(Integer) {
			switch s := symbol.String(); s {
			case "=":
				result, err = compiler.BasicEquals(result, rhs)
			case "<":
				result, err = compiler.BasicLessThan(result, rhs)
			case ">":
				result, err = compiler.BasicGreaterThan(result, rhs)
			case "+":
				result, err = compiler.BasicAdd(result, rhs)
			case "*":
				result, err = compiler.BasicMultiply(result, rhs)
			case "-":
				result, err = compiler.BasicSubtract(result, rhs)
			case "%":
				result, err = compiler.Mod(result, rhs)
			case "!":
				result, err = compiler.BasicNotEquals(result, rhs)
			case "^":
				result, err = compiler.Pow(result, rhs)
			case "/":
				result, err = compiler.Divide(result, rhs)
			}
			if err != nil {
				return result, err
			}
			continue
		}

		//Lets do the shunting!

		/*for i := range c.Shunts {
			if result = c.Shunts[i](c, symbol, t, rhs); result != nil {
				continue shunting
			}
		}*/

		return Expression{}, compiler.NewError("Operator " + string(symbol) + " does not apply to " + result.Name)
	}
	return result, nil
}
