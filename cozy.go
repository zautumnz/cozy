// Simple general-purpose interpreted programming language.
// See the docs at github.com/zautumnz/cozy.

package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"

	"github.com/zautumnz/cozy/evaluator"
	"github.com/zautumnz/cozy/lexer"
	"github.com/zautumnz/cozy/object"
	"github.com/zautumnz/cozy/parser"
	"github.com/zautumnz/cozy/repl"
	"github.com/zautumnz/cozy/utils"
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
func versionFn(args ...object.Object) object.Object {
	return &object.String{Value: COZY_VERSION}
}

// Execute the supplied string as a program.
func Execute(input string) int {
	env := object.NewEnvironment()
	l := lexer.New(input)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		parser.PrintParserErrors(parser.ParserErrorsParams{Errors: p.Errors()})
	}

	// Register a function called version()
	// that the script can call.
	evaluator.RegisterBuiltin("version",
		func(env *object.Environment, args ...object.Object) object.Object {
			return versionFn(args...)
		})

	//  Parse and evaluate our standard-library.
	initL := lexer.New(getStdlibString())
	initP := parser.New(initL)
	initProg := initP.ParseProgram()
	evaluator.Eval(initProg, env)

	//  Now evaluate the code the user wanted to load.
	//  Note that here our environment will still contain
	// the code we just loaded from our data-resource
	//  (i.e. Our cozy-based standard library.)
	evaluator.Eval(program, env)
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
		utils.ExitConditionally(0)
	}

	// Executing code?
	if *eval != "" {
		Execute(*eval)
		utils.ExitConditionally(0)
	}

	// Otherwise we're either reading from STDIN, or the
	// named file containing source-code.
	var input []byte
	var err error

	if len(flag.Args()) > 0 {
		input, err = ioutil.ReadFile(os.Args[1])
	} else {
		fmt.Printf("cozy version %s\n", COZY_VERSION)
		fmt.Println("Use ctrl+d to quit")
		repl.Start(os.Stdin, os.Stdout, getStdlibString())
	}

	if err != nil {
		fmt.Printf("Error reading: %s\n", err.Error())
	}

	Execute(string(input))
}
