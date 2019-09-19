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

	//Callables.
	if compiler.Peek().Is(".") {
		if thing, ok := result.Type.(Thing); ok {
			compiler.Scan()
			return thing.Index(compiler, result, compiler.Scan())
		}
		return result, compiler.NewError("Cannot index ", result.Type.String(compiler))
	}

	//Callables.
	if compiler.Peek().Is("(") {
		if runnable, ok := result.Type.(Runnable); ok {
			compiler.Scan()

			var args, err = compiler.Arguments()
			if err != nil {
				return Expression{}, err
			}

			result, err = runnable.Call(compiler, result, args...)
			if err != nil {
				return result, err
			}
			return compiler.Shunt(result, precedence)
		}
		return result, compiler.NewError("Cannot call ", result.Type.String(compiler))
	}

	//shunting:
	for peek := compiler.Peek(); Precedence(peek) >= precedence; {

		if Precedence(compiler.Peek()) <= -1 {
			break
		}

		symbol := peek
		if compiler.Scan() == nil {
			break
		}

		rhs, err := compiler.scanExpression()
		if err != nil {
			return result, err
		}

		peek = compiler.Peek()
		for Precedence(peek) > Precedence(symbol) {
			rhs, err = compiler.Shunt(rhs, Precedence(peek))
			if err != nil {
				return result, err
			}

			peek = compiler.Peek()
		}

		ok, expression, err := result.Operation(compiler, result, rhs, symbol.String())
		if ok {
			result = expression
			if err != nil {
				return result, err
			}
			continue
		}

		return Expression{}, compiler.NewError("Operator " + string(symbol) + " does not apply to " + result.String(compiler) + " and " + rhs.String(compiler))
	}
	return result, nil
}
