package compiler

import "errors"

type Argument struct {
	Token
	Type
	Variadic bool
}

func (compiler *Compiler) ScanArguments() ([]Argument, error) {
	var arguments []Argument

	for {
		var token = compiler.Scan()

		if compiler.ScanIf('(') {
			var filter = token

			var T = compiler.GetType(filter)
			if !Defined(T) {
				return nil, errors.New(filter.String() + " is not a valid argument filter!")
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
