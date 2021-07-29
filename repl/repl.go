package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/zacanger/cozy/evaluator"
	"github.com/zacanger/cozy/lexer"
	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/parser"
	"github.com/zacanger/cozy/utils"
)

// Start runs the REPL
func Start(in io.Reader, out io.Writer) {
	utils.SetReplOrRun(true)
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	macroEnv := object.NewEnvironment()
	for {
		fmt.Print("> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			parser.PrintParserErrors(parser.ParserErrorsParams{Errors: p.Errors(), Out: out})
			continue
		}
		evaluator.DefineMacros(program, macroEnv)
		expanded := evaluator.ExpandMacros(program, macroEnv)
		evaluated := evaluator.Eval(expanded, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}
