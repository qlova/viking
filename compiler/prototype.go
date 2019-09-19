package compiler

//Prototypes is a list of all Prototype types.
/*var Prototypes []Prototype

//Prototype is a type of type.
type Prototype struct {
	Name string

	Requirement    string
	ScanExpression func(compiler *Compiler) (Expression, error)
}

//Defined returns true if T is defined.
func (T Prototype) Defined() bool {
	return T.Name != ""
}

//Type returns a type version of the Prototype.
func (T Prototype) Type() Type {
	return Type{
		Name:      T.Name,
		Prototype: T,
	}
}

//Number is any numeric type.
var Number = Prototype{Name: "number"}

func init() {
	Number.Requirement = "type number interface{}\n"
	Number.ScanExpression = func(compiler *Compiler) (Expression, error) {
		if !compiler.ScanIf('(') {
			return Expression{}, compiler.Expecting('(')
		}
		var other, err = compiler.ScanExpression()
		if err != nil {
			return other, err
		}
		if !compiler.ScanIf(')') {
			return Expression{}, compiler.Expecting(')')
		}

		/*if other.Equals(String) {
					compiler.Throws = true

					compiler.Import("strconv")
					compiler.Require(Number.Requirement)
					compiler.Require(`func strconv_aton(ctx I.Context, s string) number {
			if i, err := strconv.Atoi(s); err == nil {
				return i
			}
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return f
			}
			ctx.Throw(1, "invalid number")
			return 0
		}
		`)

					var expression = compiler.NewExpression()
					expression.Type = Number.Type()
					expression.Go.WriteString("strconv_aton(ctx,")
					expression.Go.Write(other.Go.Bytes())
					expression.Go.WriteString(")")
					return expression, nil
				}*/

/*return other, compiler.Unimplemented(s("number()"))
	}
}

//Collection is any collection type.
var Collection = Prototype{Name: "collection"}

//Instruction is any instruction type.
var Instruction = Prototype{Name: "instruction"}

//Data is any data type.
var Data = Prototype{Name: "data"}

//Connection is any connection type.
var Connection = Prototype{Name: "connection"}

//GetPrototype returns the Prototype with the given name.
func (compiler *Compiler) GetPrototype(name []byte) Prototype {

	for _, t := range Prototypes {
		if equal(name, t.Name) {
			return t
		}
	}

	return Prototype{}
}

func init() {
	Prototypes = []Prototype{Number, Collection, Instruction, Data, Connection}
}*/
