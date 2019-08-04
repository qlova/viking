package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/containous/yaegi/interp"
	"github.com/containous/yaegi/stdlib"
	"github.com/qlova/viking/compiler"
)

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
	os.Stdin = stdinreader
	log.SetOutput(writer)
	out := make(chan []byte)
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		wg.Done()
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
	var c compiler.Compiler

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

		var buffer bytes.Buffer
		c.WriteTo(&buffer)

		out := SandBox(c.ProvidedInput, func() {
			var runtime = interp.New(interp.Options{})
			runtime.Use(stdlib.Symbols)
			_, err = runtime.Eval(buffer.String())
			if err != nil {
				fmt.Fprint(os.Stderr, err)
			}
		})

		if !bytes.Equal(out, c.ExpectedOutput) {
			fmt.Print("Expecting '", strings.Replace(string(c.ExpectedOutput), "\n", `\n`, -1),
				"' but got '", strings.Replace(string(out), "\n", `\n`, -1), "'\n")
			os.Exit(1)
		}

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

		var buffer bytes.Buffer
		c.WriteTo(&buffer)

		var runtime = interp.New(interp.Options{})
		runtime.Use(stdlib.Symbols)
		_, err = runtime.Eval(buffer.String())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	default:
		fmt.Println("[usage] viking [build] path/to/package")
		return
	}
}
