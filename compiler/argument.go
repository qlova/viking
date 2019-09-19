package compiler

import "errors"

//Argument is a function/concept argument.
type Argument struct {
	Token
	Type
	Variadic bool
}

//Arguments is a wrapper around exression lists.
type Arguments []Expression

func (args Arguments) String() string {
	var result string
	for i, arg := range args {
		result += arg.Go.String()
		if i < len(args)-1 {
			result += ","
		}
	}
	return result
}

//Arguments scans and returns a slice of arguments.
func (compiler *Compiler) Arguments() (Arguments []Expression, err error) {
	if !compiler.ScanIf(')') {
		first, err := compiler.ScanExpression()
		if err != nil {
			return nil, err
		}

		Arguments = append(Arguments, first)

		for compiler.ScanIf(',') {
			expression, err := compiler.ScanExpression()
			if err != nil {
				return nil, err
			}

			Arguments = append(Arguments, expression)
		}
		if !compiler.ScanIf(')') {
			return nil, compiler.NewError("expecting )")
		}
	}
	return
}

//Indicies scans and returns a slice of indicies.
func (compiler *Compiler) Indicies() (indicies []Expression, err error) {
	if !compiler.ScanIf('[') {
		return nil, nil
	}
	if !compiler.ScanIf(']') {
		first, err := compiler.ScanExpression()
		if err != nil {
			return nil, err
		}

		indicies = append(indicies, first)

		for compiler.ScanIf(',') {
			expression, err := compiler.ScanExpression()
			if err != nil {
				return nil, err
			}

			indicies = append(indicies, expression)
		}
		if !compiler.ScanIf(']') {
			return nil, compiler.NewError("expecting ]")
		}
	}
	return
}

//ScanArguments scans a function/concept argument definition.
func (compiler *Compiler) ScanArguments() ([]Argument, error) {
	var arguments []Argument

	for {
		var token = compiler.Scan()

		if compiler.ScanIf('(') {
			var filter = token

			var T = compiler.Type(filter)
			if !Defined(T) {
				return nil, errors.New(filter.String() + " is not a valid argument filter!")
			}

			T, err := compiler.SpecifyType(T)
			if err != nil {
				return nil, err
			}

			for {
				var token = compiler.Scan()

				var variadic = false
				if compiler.ScanIf('.') {
					if !compiler.ScanIf('.') {
						return nil, compiler.Expecting('.')
					}
					if !compiler.ScanIf('.') {
						return nil, compiler.Expecting('.')
					}

					variadic = true
				}

				arguments = append(arguments, Argument{
					Token:    token,
					Type:     T,
					Variadic: variadic,
				})

				if compiler.ScanIf(')') {
					break
				}

				if !compiler.ScanIf(',') {
					return nil, compiler.Expecting(',')
				}
			}
		} else {
			arguments = append(arguments, Argument{
				Token: token,
			})
		}

		if compiler.ScanIf(')') {
			break
		}
		if !compiler.ScanIf(',') {
			return nil, compiler.Expecting(',')
		}
	}

	return arguments, nil
}
