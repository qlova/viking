package main

import "os"
import "os/exec"
import "fmt"
import "path"
import "bytes"
import "github.com/qlova/viking/compiler"

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
				c.Dir = os.Args[2]
			}
			
			err := c.Compile()
			if err != nil {
				fmt.Println(err)
			}
			
			c.WriteTo(os.Stdout)
			
		case "test":
		
			if len(os.Args) > 2 {
				c.Dir = os.Args[2]
			}
			
			err := c.Compile()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			
			err = os.Mkdir(path.Join(c.Dir, ".viking"), 0755)
			
			var location = path.Join(c.Dir, ".viking", "main.go")
			
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
			os.Exit(0)
			
		case "run":
			if len(os.Args) > 2 {
				c.Dir = os.Args[2]
			}
			
			err := c.Compile()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			
			err = os.Mkdir(path.Join(c.Dir, ".viking"), 0755)
			
			var location = path.Join(c.Dir, ".viking", "main.go")
			
			file, err := os.Create(location)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			
			c.WriteTo(file)
			
			file.Close()
			
			var Go = exec.Command("go", "run", location)
				Go.Stdout = os.Stdout
				Go.Stderr = os.Stderr
				
			err = Go.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			
		default:
			fmt.Println("[usage] viking [build] path/to/package")
			return
	}
}
