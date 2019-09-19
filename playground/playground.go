package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall/js"

	"github.com/qlova/viking/compiler/target"

	"github.com/cosmos72/gomacro/fast"
	"github.com/qlova/viking/compiler"
	_ "github.com/qlova/viking/compiler/builtin"
	_ "github.com/qlova/viking/compiler/statement"

	_ "github.com/qlova/i/imacro"
)

//Run runs the compiler.
func Run(compiler compiler.Compiler) (err error) {
	var buffer bytes.Buffer
	compiler.WriteTo(&buffer)

	/*out := SandBox(compiler.ProvidedInput, func() {
		var runtime = interp.New(interp.Options{})
		runtime.Use(stdlib.Symbols)
		_, err = runtime.Eval(buffer.String())
	})
	if len(out) == 0 || err != nil {*/
	var runtime = fast.New()
	runtime.ChangePackage("main", "")
	runtime.Eval(strings.Replace(buffer.String(), "package main", "", 1))
	out := SandBox(nil, func() {
		runtime.Eval("main()")
	})
	//}

	_, err = os.Stdout.Write(out)

	return
}

func SandBox(input []byte, f func()) []byte {
	f()
	return nil
}

func main() {
	var Value = js.Global().Get("editor").Call("getValue").String()

	var compiler = compiler.New()
	compiler.SetTarget(target.Go)
	compiler.Filename = "editor"

	err := compiler.CompileReader(strings.NewReader(Value))
	if err != nil {
		fmt.Println(err)
		return
	}

	Run(compiler)
}
