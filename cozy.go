// Simple, high-ish-level interpreted programming language that sits somewhere
// between scripting and general-purpose programming. Dynamically and strongly
// typed, with some with semantics that work well with pseudo-functional
// programming but syntax similar to Python, Go, and Shell; no OOP constructs
// like classes; instead we have first-class functions, closures, and macros.
// For more details, see github.com/zacanger/cozy.

package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"

	"github.com/zacanger/cozy/evaluator"
	"github.com/zacanger/cozy/lexer"
	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/parser"
	"github.com/zacanger/cozy/repl"
)

// COZY_VERSION is replaced by go build in makefile
var COZY_VERSION = "cozy-version"

//go:embed stdlib
var stdlibFs embed.FS

// turn the embed fs into a string we can use
func getStdlibString() string {
	s := ""
	fs.WalkDir(
		stdlibFs,
		".",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".cz") {
				c, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				s += string(c)
				s += "\n"
			}

			return nil
		})

	return s
}

// Implemention of "version()" function.
func versionFun(args ...object.Object) object.Object {
	return &object.String{Value: COZY_VERSION}
}

// Implemention of "args()" function.
func argsFun(args ...object.Object) object.Object {
	l := len(os.Args[1:])
	result := make([]object.Object, l)
	for i, txt := range os.Args[1:] {
		result[i] = &object.String{Value: txt}
	}
	return &object.Array{Elements: result}
}

// Execute the supplied string as a program.
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

	//  Parse and evaluate our standard-library.
	initL := lexer.New(getStdlibString())
	initP := parser.New(initL)
	initProg := initP.ParseProgram()
	evaluator.DefineMacros(initProg, macroEnv)
	expanded := evaluator.ExpandMacros(initProg, macroEnv)
	evaluator.Eval(expanded, env)

	//  Now evaluate the code the user wanted to load.
	//  Note that here our environment will still contain
	// the code we just loaded from our data-resource
	//  (i.e. Our cozy-based standard library.)
	evaluator.DefineMacros(program, macroEnv)
	expandedProg := evaluator.ExpandMacros(program, macroEnv)
	evaluator.Eval(expandedProg, env)
	return 0
}

func main() {
	// Setup some flags.
	evalDesc := "Code to execute"
	eval := flag.String("eval", "", evalDesc)
	flag.StringVar(eval, "e", "", evalDesc)
	versDesc := "Show our version and exit"
	vers := flag.Bool("version", false, versDesc)
	flag.BoolVar(vers, "v", false, versDesc)

	// Parse the flags
	flag.Parse()

	// Showing the version?
	if *vers {
		fmt.Printf("cozy %s\n", COZY_VERSION)
		os.Exit(1)
	}

	// Executing code?
	if *eval != "" {
		Execute(*eval)
		os.Exit(1)
	}

	// Otherwise we're either reading from STDIN, or the
	// named file containing source-code.
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
