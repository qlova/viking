package compiler

import (
	"bytes"
	"github.com/qlova/viking/compiler/target"
)

func (compiler *Compiler) getTargetBuffer(target *target.Mode, token Token) (*bytes.Buffer, error) {
	switch s := token.String(); s {
	case "neck":
		return &target.Neck, nil
	case "head":
		return &target.Head, nil
	case "tail":
		return &target.Tail, nil
	default:
		return nil, compiler.NewError("invalid target directive: " + s)
	}
}

func (compiler *Compiler) processRequirement(T target.Target, buffer *bytes.Buffer, data []byte) {
	var code = string(data)

	if buffer == &(compiler.Get(T).Head) {
		if compiler.Requirements.Head == nil {
			compiler.Requirements.Head = make(Set)
		}
		if !compiler.Requirements.Head.Get(code) {
			compiler.Requirements.Head.Set(code)
			compiler.FlipBuffer()
			compiler.Get(T).Head.Write(data)
			compiler.DumpBuffer()
		}
	} else if buffer == &(compiler.Get(T).Neck) {
		if compiler.Requirements.Neck == nil {
			compiler.Requirements.Neck = make(Set)
		}
		if !compiler.Requirements.Neck.Get(code) {
			compiler.Requirements.Neck.Set(code)
			compiler.FlipBuffer()
			compiler.Get(T).Neck.Write(data)
			compiler.DumpBuffer()
		}
	} else if buffer == &(compiler.Get(T).Tail) {
		if compiler.Requirements.Tail == nil {
			compiler.Requirements.Tail = make(Set)
		}
		if !compiler.Requirements.Tail.Get(code) {
			compiler.Requirements.Tail.Set(code)
			compiler.FlipBuffer()
			compiler.Get(T).Tail.Write(data)
			compiler.DumpBuffer()
		}
	} else {
		panic("no equalities")
	}

	return
}

func (compiler *Compiler) ScanIfStatement() error {

	//native code.
	if compiler.ScanIf('.') {

		var T = target.FromString(compiler.Scan().String())
		if !T.Valid() {
			return compiler.NewError("invalid target")
		}

		var TargetMode = compiler.Get(T)
		var writeTo = &TargetMode.Body
		var require bool

		if compiler.ScanIf('.') {
			var err error
			var mode Token
			mode = compiler.Scan()
			if mode.Is("require") {
				if !compiler.ScanIf('.') {
					return compiler.Expecting('.')
				}
				mode = compiler.Scan()

				require = true
			}

			writeTo, err = compiler.getTargetBuffer(TargetMode, mode)
			if err != nil {
				return err
			}
		}

		if compiler.ScanIf(':') {
			var Line, err = compiler.Reader.ReadBytes('\n')
			if err != nil {
				return err
			}

			if require {
				if TargetMode.Enabled {
					compiler.processRequirement(T, writeTo, Line)
				}
			} else {
				writeTo.Write(Line)
			}

			return nil
		}

		return compiler.Unimplemented(s("if.target {block}"))
	}

	//Standard if statement.
	var condition, err = compiler.ScanExpression()
	if err != nil {
		return err
	}

	if !condition.Equals(Bit) {
		condition, err = compiler.Cast(condition, Bit)
		if err != nil {
			return err
		}
	}
	compiler.Indent()
	compiler.Go.WriteString("if ")
	compiler.Go.Write(condition.Go.Bytes())
	compiler.Go.WriteString(" {")

	compiler.GainScope()
	return compiler.CompileBlock()
}
