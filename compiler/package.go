package compiler

import (
	"os"
	"path/filepath"
)

//Package stores saved context for a package.
type Package struct {
	Context
}

//GetPackage returns the package with the given name.
func (compiler *Compiler) GetPackage(name Token) (Package, error) {
	if P, ok := compiler.Packages[name.String()]; ok {
		return P, nil
	}

	return compiler.CompilePackage(name.String())
}

func (compiler *Compiler) CompilePackage(name string) (Package, error) {
	var executable, err = os.Executable()
	if err != nil {
		return Package{}, err
	}

	var path = filepath.Dir(executable) + "/library/" + name + "/" + name + ".i"

	var context = compiler.NewPackageContext()
	context.Directory = path

	compiler.PushContext(context)
	err = compiler.CompileFile(path)

	return Package{context}, err
}

func (P Package) Expression(compiler *Compiler) (Expression, error) {
	if !compiler.ScanIf('.') {
		return Expression{}, compiler.Expecting('.')
	}

	var token = compiler.Scan()

	concept, ok := P.Concepts[token.String()]
	if !ok {
		return Expression{}, compiler.Undefined(token)
	}

	return concept.Call(compiler)
}
