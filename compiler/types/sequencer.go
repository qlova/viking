package types

import "github.com/qlova/viking/compiler"

//Sequencer can be + or -
type Sequencer struct {
	compiler.Nothing

	Plus  bool
	Minus bool
}

var _ = compiler.RegisterType(Sequencer{})

//Name returns the name of this type.
func (Sequencer) Name() compiler.String {
	return compiler.String{
		compiler.English: `sequencer`,
	}
}

//Equals returns true if the other type is equal to this type.
func (Sequencer) Equals(other compiler.Type) bool {
	_, ok := other.(Sequencer)
	return ok
}

//Expression returns a sequencer expression.
func (Sequencer) Expression(c *compiler.Compiler) (ok bool, expression compiler.Expression, err error) {
	expression = c.NewExpression()

	if c.Token().Is("+") {
		expression.Type = Sequencer{
			Plus: true,
		}
		return true, expression, nil
	}

	if c.Token().Is("-") {
		expression.Type = Sequencer{
			Minus: true,
		}
		return true, expression, nil
	}

	return
}
