package compiler

//Types is a slice of all static types that the compiler package is aware of.
var Types = []Type{Nothing{}, Thing{}, Sequence{}}

//RegisterType registers a new type to the compiler package and then returns it.
func RegisterType(T Type) Type {
	Types = append(Types, T)
	return T
}

//Runnable are runnable things.
type Runnable interface {
	Call(c *Compiler, this Expression, args ...Expression) (Expression, error)
	Run(c *Compiler, this Expression, args ...Expression) error
}

//Type is an abstract type interface that can be used to define new types.
type Type interface {
	//Name should return the name of this type.
	Name() String

	//String returns this type as a human readable string.
	String(c *Compiler) string

	//Expression should evaluate the current token and attempt to return it as an expression of this type.
	//ok if success.
	Expression(c *Compiler) (bool, Expression, error)

	//Operation operates on Type (a) with Expression (b) based on the provided symbol.
	Operation(c *Compiler, a, b Expression, symbol string) (bool, Expression, error)

	//Cast cast this type as a from Expression to the requested 'to' type.
	Cast(c *Compiler, from Expression, to Type) (Expression, error)

	//Equals should return true if the other type is exactly the same as this type.
	Equals(other Type) bool

	//Native is a helper method that returns the native type token for this type.
	Native(c *Compiler) Token

	//Zero is a helper method that returns the zero-value of this type.
	Zero(c *Compiler) Expression

	//Copy is a helper method that returns a copy of an expression of this type.
	Copy(c *Compiler, e Expression) (Expression, error)
}

//Collection is a supertype of Type, collections can contain subtypes, be indexed and called.
type Collection interface {
	Type

	//Index a value of this type with the specified indicies.
	Index(c *Compiler, this Expression, indices ...Expression) (Expression, error)

	//Modify a value of this type with the specified indicies.
	Modify(c *Compiler, this Expression, modification Expression, indices ...Expression) error

	//Specify this type with the provided args.
	Specify(c *Compiler, args ...Expression) (Type, error)

	//With should return this type containing the provided subtype.
	With(c *Compiler, subtype Type) Type

	//Subtype returns the subtype of this collection.
	Subtype() Type

	//Length returns the size of this collection as an expression.
	Length(c *Compiler, this Expression) Expression
}

//Connection is a connection with the outside world.
type Connection interface {
	In(this Expression, args ...Expression) (Expression, error)
	Out(this Expression, args ...Expression) error
}

//SpecifyType specifies a collection type.
func (compiler *Compiler) SpecifyType(T Type) (Type, error) {
	var args, err = compiler.Indicies()
	if err != nil {
		return nil, err
	}

	var subtype Type
	if compiler.ScanIf('.') {

		subtype = compiler.Type(compiler.Scan())
		if !Defined(subtype) {
			return nil, compiler.NewError(compiler.Token().String() + " is not a type!")
		}
		subtype, err = compiler.SpecifyType(subtype)
		if err != nil {
			return nil, err
		}
	}

	if collection, ok := T.(Collection); ok {
		T, err = collection.Specify(compiler, args...)
		if err != nil {
			return nil, err
		}
		if T == nil {
			return nil, compiler.NewError("cannot specify " + T.String(compiler))
		}
		T = T.(Collection).With(compiler, subtype)
	} else if len(args) > 0 || subtype != nil {
		return nil, compiler.NewError(T.Name()[English] + " is not a collection type!")
	}

	return T, nil
}

//Type returns the Type with the given name.
func (compiler *Compiler) Type(name []byte) Type {

	for _, t := range Types {
		if equal(name, t.Name()[compiler.Language]) {
			return t
		}
	}

	return nil
}

//Deterministic sets the compiler to produce Deterministic code.
var Deterministic = true

//GoTypeOf returns the go type of the Type.
func (compiler *Compiler) GoTypeOf(t Type) []byte {
	return t.Native(compiler)
}
