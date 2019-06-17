package compiler

import "io"
import "bytes"

func (compiler *Compiler) CompileStatement() error {
	var token = compiler.Scan()
	/*switch token {
		case "if", "for", "return", "break", "go":
			return Unimplemented
	}*/
	
	if token == nil {
		return io.EOF
	}
	
	if bytes.Equal(token, []byte("\n")) {
		return nil
	}
	
	if len(token) > 2 && token[0] == '/' && token[1] == '/' {
		compiler.Write(token)
		
		if len(token) > len("//output: ") && bytes.Equal(token[:len("//output: ")], []byte("//output: ")) {
			compiler.ExpectedOutput = token[len("//output: "):]
		}
		
		return nil
	}
	
	if bytes.Equal(token, []byte("main")) {
		
		compiler.Write([]byte("func main() {"))
		compiler.GainScope()
		
		if compiler.ScanIf(':') {
			defer func() {
				compiler.Write([]byte("}"))
				compiler.LoseScope()
			}()
			return compiler.CompileStatement()
		} else {
			
			if !compiler.ScanIf('\n') {
				return compiler.Unexpected()
			}
			
			compiler.Depth++
		}
		
		
		return nil
	}
	
	if bytes.Equal(token, []byte("}")) {
		compiler.Depth--
		compiler.Write([]byte("}"))
		compiler.LoseScope()
		return nil
	}
	
	compiler.Indent()

	if Builtin(token) {
		return compiler.CompileBuiltin(token)
	}
	
	if bytes.Equal(compiler.Peek(), []byte{'['}) {
		if Defined(compiler.GetVariable(token)) {
			return compiler.ModifyArray(token)
		}
	}

	if bytes.Equal(compiler.Peek(), []byte{'='}) {
		if Defined(compiler.GetVariable(token)) {
			return compiler.AssignVariable(token)
		} else {
			return compiler.DefineVariable(token)
		}
	}
	
	return Unimplemented(token)
}
