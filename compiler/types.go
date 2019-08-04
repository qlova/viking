package compiler

import "errors"
import "strconv"

//Field is an 'i' type field.
type Field struct {
	Name     string
	Type     Type
	Embedded bool
}

//Type is a 'i' type.
type Type struct {
	Name    string
	Size    int
	Subtype *Type

	//Can this type be modified?
	Frozen bool

	//Is this type a share?
	Share bool

	//The relative value of this type.
	Value int

	Fields []Field
}

//String is an immutable sequence of symbols.
var String = Type{Name: "string"}

//Symbol is a contextual reference point.
var Symbol = Type{Name: "symbol"}

//Integer is a positive or negative integer.
var Integer = Type{Name: "integer"}

//Byte is a precisional reference point.
var Byte = Type{Name: "byte"}

//Function is a code block that can be run with parameters.
var Function = Type{Name: "function"}

//Array is a fixed-length sequence of values.
var Array = Type{Name: "array"}

//Variadic is a dynamic-length sequence of values.
var Variadic = Type{Name: "variadic"}

//Types is a slice of all 'i' types.
var Types = []Type{String, Integer, Symbol, Array, List, Byte, Function, Variadic}

//Is returns true if Type is a collection of type 'collection'.
func (a Type) Is(collection Type) bool {
	return a.Name == collection.Name
}

//Collection returns Type 'a' in collection.
func (a Type) Collection(collection Type) Type {
	collection.Subtype = &a
	return collection
}

//Equals checks if Type a is equal to Type b.
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

//GetType returns the Type with the given name.
func (compiler *Compiler) GetType(name []byte) Type {

	for _, t := range Types {
		if equal(name, t.Name) {
			return t
		}
	}

	return Type{}
}

//Type returns the type as an expression.
func (compiler *Compiler) Type(t Type) (Expression, error) {
	var expression Expression
	expression.Type = t

	switch t.Name {
	case "integer":
		expression.Write([]byte("int(0)"))
	}

	return Expression{}, errors.New("Invalid type")
}

//GoTypeOf returns the go type of the Type.
func GoTypeOf(t Type) []byte {
	switch t.Name {
	case "array":
		return append(append([]byte("["+strconv.Itoa(t.Size)+"]"), GoTypeOf(*t.Subtype)...))
	case "list":
		return append(append([]byte("[]"), GoTypeOf(*t.Subtype)...))
	case "string":
		return s("string")
	case "integer":
		return s("int")
	case "function":
		return s("func()")
	case "symbol":
		return s("rune")
	}

	panic("unimplemented " + t.Name)
}

//Collection returns a collection of Type t with the specified subtype.
func (compiler *Compiler) Collection(t Type, subtype Type) (Expression, error) {
	var expression Expression
	expression.Type = t
	expression.Type.Subtype = &subtype

	var next = compiler.Scan()

	var index, other Expression
	var err error

	if next.Is("[") {
		index, err = compiler.ScanExpression()
		if err != nil {
			return Expression{}, err
		}

		if !compiler.ScanIf(']') {
			return Expression{}, compiler.Expecting(']')
		}
	}

	if !compiler.ScanIf('(') {
		return Expression{}, compiler.Expecting('(')
	}

	if !compiler.ScanIf(')') {
		other, err = compiler.ScanExpression()
		if err != nil {
			return Expression{}, err
		}

		if !compiler.ScanIf(')') {
			return Expression{}, compiler.Expecting(')')
		}
	}

	switch t.Name {
	case "array":
		size, err := strconv.Atoi(string(index.Bytes()))
		if err != nil {
			return Expression{}, errors.New("Invalid array size " + strconv.Itoa(size))
		}

		_ = other.String()

		expression.Size = size
		expression.Write(GoTypeOf(expression.Type))
		expression.WriteString("{}")
		return expression, nil
	}

	return Expression{}, errors.New("Invalid type")
}
