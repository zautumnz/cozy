// Cozy is a scripting language implemented in golang, based upon
// the book "Write an Interpreter in Go", written by Thorsten Ball.
//
// This implementation adds a number of tweaks, improvements, and new
// features.  For example we support file-based I/O, regular expressions,
// the ternary operator, and more.
//
// For full details please consult the project homepage
// https://github.com/zacanger/cozy/
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zacanger/cozy/evaluator"
	"github.com/zacanger/cozy/lexer"
	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/parser"
	"github.com/zacanger/cozy/repl"
)

// COZY_VERSION is replaced by go build in makefile
var COZY_VERSION = "cozy-version"

// TODO: embed the dir with stdlib/* as an embed.FS, read each file,
// and create a single string; is this possible without caring about the file
// names?

//go:embed stdlib/misc.cz
var misc string

//go:embed stdlib/array.cz
var array string

//go:embed stdlib/string.cz
var strings string

//go:embed stdlib/test.cz
var tests string

//go:embed stdlib/event-emitter.cz
var eventEmitter string

//go:embed stdlib/state-management.cz
var stateManagement string

//
// Implemention of "version()" function.
//
func versionFun(args ...object.Object) object.Object {
	return &object.String{Value: COZY_VERSION}
}

//
// Implemention of "args()" function.
//
func argsFun(args ...object.Object) object.Object {
	l := len(os.Args[1:])
	result := make([]object.Object, l)
	for i, txt := range os.Args[1:] {
		result[i] = &object.String{Value: txt}
	}
	return &object.Array{Elements: result}
}

//
// Execute the supplied string as a program.
//
func Execute(input string) int {

	env := object.NewEnvironment()
	macroEnv := object.NewEnvironment()
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	// Register a function called version()
	// that the script can call.
	evaluator.RegisterBuiltin("version",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (versionFun(args...))
		})

	// Access to the command-line arguments
	evaluator.RegisterBuiltin("args",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (argsFun(args...))
		})

	//
	//  Parse and evaluate our standard-library.
	//
	initL := lexer.New(
		misc,
		array,
		strings,
		tests,
		eventEmitter,
		stateManagement,
	)
	initP := parser.New(initL)
	initProg := initP.ParseProgram()
	evaluator.DefineMacros(initProg, macroEnv)
	expanded := evaluator.ExpandMacros(initProg, macroEnv)
	evaluator.Eval(expanded, env)

	//
	//  Now evaluate the code the user wanted to load.
	//
	//  Note that here our environment will still contain
	// the code we just loaded from our data-resource
	//
	//  (i.e. Our cozy-based standard library.)
	//
	evaluator.DefineMacros(program, macroEnv)
	expandedProg := evaluator.ExpandMacros(program, macroEnv)
	evaluator.Eval(expandedProg, env)
	return 0
}

func main() {

	//
	// Setup some flags.
	//
	evalDesc := "Code to execute."
	eval := flag.String("eval", "", evalDesc)
	flag.StringVar(eval, "e", "", evalDesc)
	versDesc := "Show our version and exit."
	vers := flag.Bool("version", false, versDesc)
	flag.BoolVar(vers, "v", false, versDesc)

	//
	// Parse the flags
	//
	flag.Parse()

	//
	// Showing the version?
	//
	if *vers {
		fmt.Printf("cozy %s\n", COZY_VERSION)
		os.Exit(1)
	}

	//
	// Executing code?
	//
	if *eval != "" {
		Execute(*eval)
		os.Exit(1)
	}

	//
	// Otherwise we're either reading from STDIN, or the
	// named file containing source-code.
	//
	var input []byte
	var err error

	if len(flag.Args()) > 0 {
		input, err = ioutil.ReadFile(os.Args[1])
	} else {
		fmt.Printf("cozy version %s\n", COZY_VERSION)
		fmt.Println("Use ctrl+c or exit() to quit")
		repl.Start(os.Stdin, os.Stdout)
	}

	if err != nil {
		fmt.Printf("Error reading: %s\n", err.Error())
	}

	Execute(string(input))
}
