package compiler

import "errors"
import "strconv"

type Type struct {
	Name string
	Size int
	Subtype *Type
}

var String = Type{Name: "string"}
var Symbol = Type{Name: "symbol"}
var Integer = Type{Name: "integer"}

var Array = Type{Name: "array"}

var Types = []Type{String, Integer, Symbol, Array}

func (a Type) Equals(b Type) bool {
	
	if a.Subtype != nil && b.Subtype != nil {
		return a.Name == b.Name && a.Subtype.Equals(*b.Subtype)
	}
	
	if a.Subtype == nil && b.Subtype == nil {
		return a.Name == b.Name
	}
	
	if a.Subtype == nil || b.Subtype == nil {
		return false
	}
	
	return a.Name == b.Name
}

func (compiler *Compiler) GetType(name []byte) Type {
	
	for _, t := range Types {
		if equal(name, t.Name) {
			return t
		}
	}
	
	return Type{}
}

func (compiler *Compiler) Type(t Type) (Expression, error) {
	var expression Expression
		expression.Type = t
	
	switch t.Name {
		case "integer":
			expression.Write([]byte("int(0)"))
	}
	
	return Expression{}, errors.New("Invalid type")
}

func GoTypeOf(t Type) []byte {
	switch t.Name {
		case "array":
			return append(append([]byte("["+strconv.Itoa(t.Size)+"]"), GoTypeOf(*t.Subtype)...), s("{}")...)
		case "integer":
			return s("int")
	}
	
	panic("unimplemented "+ t.Name)
	return nil
}

func (compiler *Compiler) Collection(t Type, subtype Type) (Expression, error) {
	var expression Expression
		expression.Type = t
		expression.Type.Subtype = &subtype
	
	if !compiler.ScanIf('(') {
		return Expression{}, compiler.Expecting('(')
	}
	
	var other, err = compiler.ScanExpression()
	if err != nil {
		return Expression{}, err
	}
	
	if !compiler.ScanIf(')') {
		return Expression{}, compiler.Expecting(')')
	}
	
	switch t.Name {
		case "array":
			size, err := strconv.Atoi(string(other.Bytes()))
			if err != nil {
				return Expression{}, errors.New("Invalid array size "+strconv.Itoa(size))
			}

			expression.Size = size
			expression.Write(GoTypeOf(expression.Type))
			return expression, nil
	}
	
	return Expression{}, errors.New("Invalid type")
}
