package statement

import (
	"bytes"

	"github.com/qlova/viking/compiler"
	"github.com/qlova/viking/compiler/target"
	"github.com/qlova/viking/compiler/types"
)

//If is an if statement.
type If struct{}

var _ = compiler.RegisterStatement(If{})

//Name returns the name of this statement.
func (If) Name() compiler.String {
	return compiler.String{
		compiler.English: `if`,
	}
}

//Compile compiles this statement.
func (If) Compile(c *compiler.Compiler) error {

	GetTargetBuffer := func(target *target.Mode, token compiler.Token) (*bytes.Buffer, error) {
		switch s := token.String(); s {
		case "neck":
			return &target.Neck, nil
		case "head":
			return &target.Head, nil
		case "tail":
			return &target.Tail, nil
		default:
			return nil, c.NewError("invalid target directive: " + s)
		}
	}

	ProcessRequirement := func(T target.Target, buffer *bytes.Buffer, data []byte) {
		var code = string(data)

		if buffer == &(c.Get(T).Head) {
			if c.Requirements.Head == nil {
				c.Requirements.Head = make(compiler.Set)
			}
			if !c.Requirements.Head.Get(code) {
				c.Requirements.Head.Set(code)
				c.FlipBuffer()
				c.Get(T).Head.Write(data)
				c.DumpBuffer(nil)
			}
		} else if buffer == &(c.Get(T).Neck) {
			if c.Requirements.Neck == nil {
				c.Requirements.Neck = make(compiler.Set)
			}
			if !c.Requirements.Neck.Get(code) {
				c.Requirements.Neck.Set(code)
				c.FlipBuffer()
				c.Get(T).Neck.Write(data)
				c.DumpBuffer(nil)
			}
		} else if buffer == &(c.Get(T).Tail) {
			if c.Requirements.Tail == nil {
				c.Requirements.Tail = make(compiler.Set)
			}
			if !c.Requirements.Tail.Get(code) {
				c.Requirements.Tail.Set(code)
				c.FlipBuffer()
				c.Get(T).Tail.Write(data)
				c.DumpBuffer(nil)
			}
		} else {
			panic("no equalities")
		}

		return
	}

	//native code.
	if c.ScanIf('.') {

		var T = target.FromString(c.Scan().String())
		if !T.Valid() {
			return c.NewError("invalid target")
		}

		var TargetMode = c.Get(T)
		var writeTo = &TargetMode.Body
		var require bool

		if c.ScanIf('.') {
			var err error
			var mode compiler.Token
			mode = c.Scan()
			if mode.Is("require") {
				if !c.ScanIf('.') {
					return c.Expecting('.')
				}
				mode = c.Scan()

				require = true
			}

			writeTo, err = GetTargetBuffer(TargetMode, mode)
			if err != nil {
				return err
			}
		}

		if c.ScanIf(':') {
			var Line, err = c.Reader.ReadBytes('\n')
			if err != nil {
				return err
			}

			if require {
				if TargetMode.Enabled {
					ProcessRequirement(T, writeTo, Line)
				}
			} else {
				writeTo.Write(Line)
			}

			return nil
		}

		return c.Unimplemented(compiler.Token("if.target {block}"))
	}

	//Standard if statement.
	var condition, err = c.ScanExpression()
	if err != nil {
		return err
	}

	if !condition.Equals(types.Logical{}) {
		condition, err = c.Cast(condition, types.Logical{})
		if err != nil {
			return err
		}
	}
	c.Indent()
	c.Go.WriteString("if ")
	c.Go.Write(condition.Go.Bytes())
	c.Go.WriteString(" {")

	c.GainScope()
	c.SetFlag(compiler.Token("if"))

	var singleLine = c.Peek().Is(":")
	if err := c.CompileBlock(); err != nil {
		return err
	}

	if singleLine {
		c.ScanLine()
	}

	//Continuation elseif or else
	for {
		if c.ScanIf('|') {
			if c.ScanIf('|') {

				//Standard elseif statement.
				var condition, err = c.ScanExpression()
				if err != nil {
					return err
				}

				if !condition.Equals(types.Logical{}) {
					condition, err = c.Cast(condition, types.Logical{})
					if err != nil {
						return err
					}
				}

				c.Indent()
				c.Go.WriteString("else if ")
				c.Go.Write(condition.Go.Bytes())
				c.Go.WriteString(" {")
				c.GainScope()
				c.SetFlag(compiler.Token("if"))
				var singleLine = c.Peek().Is(":")
				if err := c.CompileBlock(); err != nil {
					return err
				}
				if singleLine {
					c.ScanLine()
				}
				continue
			}
			c.Indent()
			c.Go.WriteString(" else {")
			c.GainScope()
			if err := c.CompileBlock(); err != nil {
				return err
			}
			return nil
		}
		break
	}

	return nil
}
