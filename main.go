package main

import (
	"bytes"
	"fmt"
	"github.com/qlova/viking/compiler/target"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/cosmos72/gomacro/fast"
	"github.com/cosmos72/gomacro/imports"
	"github.com/qlova/viking/compiler"
)

func init() {
	imports.Packages["os"].Binds["Stdin"] = reflect.ValueOf(&os.Stdin).Elem()
}

//Test tests the compiler.
func Test(compiler compiler.Compiler) (err error) {
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
	out := SandBox(compiler.ProvidedInput, func() {
		func() {
			defer func() {
				recover()
			}()
			runtime.Eval("std_in = bufio.NewReader(os.Stdin)")
		}()
		runtime.Eval("main()")
	})
	//}

	if !bytes.Equal(out, compiler.ExpectedOutput) {
		fmt.Print("Expecting '", strings.Replace(string(compiler.ExpectedOutput), "\n", `\n`, -1),
			"' but got '", strings.Replace(string(out), "\n", `\n`, -1), "'\n")
		os.Exit(1)
	}

	return nil
}

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
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	stdinreader, stdinwriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	stdout := os.Stdout
	stderr := os.Stderr
	stdin := os.Stdin
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
		os.Stdin = stdin
		log.SetOutput(os.Stderr)
	}()
	os.Stdout = writer
	os.Stderr = writer
	if input != nil {
		os.Stdin = stdinreader
	}
	log.SetOutput(writer)
	out := make(chan []byte)
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		wg.Done()
		if input == nil {
			return
		}
		stdinwriter.Write(input)
	}()
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, reader)
		out <- buf.Bytes()
	}()
	wg.Wait()
	f()
	writer.Close()
	return <-out
}

func main() {
	var c = compiler.New()
	c.SetTarget(target.Go)

	if len(os.Args) <= 1 {
		fmt.Println("[usage] viking [build/test] path/to/package")
		return
	}

	var directive = os.Args[1]
	switch directive {
	case "build":

		if len(os.Args) > 2 {
			c.Directory = os.Args[2]
		}

		err := c.Compile()
		if err != nil {
			fmt.Println(err)
		}

		c.WriteTo(os.Stdout)

	case "test":

		if len(os.Args) > 2 {
			c.Directory = os.Args[2]
		}

		err := c.Compile()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		Test(c)

		//os.Exit(0)

		/*err = os.Mkdir(path.Join(c.Directory, ".viking"), 0755)

		var location = path.Join(c.Directory, ".viking", "main.go")

		file, err := os.Create(location)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		c.WriteTo(file)

		file.Close()

		var output bytes.Buffer

		var Go = exec.Command("go", "run", location)
		Go.Stdout = &output
		Go.Stderr = os.Stderr
		Go.Stdin = bytes.NewReader(bytes.Replace(c.ProvidedInput, []byte(`\n`), []byte("\n"), -1))

		err = Go.Run()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var b = output.Bytes()

		b = bytes.Replace(b, []byte("\n"), []byte(`\n`), -1)

		if !bytes.Equal(b, c.ExpectedOutput) {
			fmt.Print("Expecting '", string(c.ExpectedOutput), "' but got '", string(b), "'\n")
			os.Exit(1)
		}
		os.Exit(0)*/

	case "run":
		if len(os.Args) > 2 {
			c.Directory = os.Args[2]
		}

		err := c.Compile()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		Run(c)

	default:
		fmt.Println("[usage] viking [build] path/to/package")
		return
	}
}
