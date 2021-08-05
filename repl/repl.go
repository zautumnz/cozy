package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/zacanger/cozy/evaluator"
	"github.com/zacanger/cozy/lexer"
	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/parser"
	"github.com/zacanger/cozy/repl/readline"
	"github.com/zacanger/cozy/utils"
)

func useFancyRepl() bool {
	o := runtime.GOOS
	return o == "darwin" || o == "linux" || strings.HasSuffix(o, "bsd")
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// Start runs the REPL
func Start(in io.Reader, out io.Writer, stdlib string) {
	utils.SetReplOrRun(true)
	env := object.NewEnvironment()
	macroEnv := object.NewEnvironment()

	if useFancyRepl() {
		l, err := readline.NewEx(&readline.Config{
			Prompt:              "> ",
			HistoryFile:         os.Getenv("HOME") + "/.cozy_history",
			InterruptPrompt:     "^C",
			EOFPrompt:           "exit",
			HistorySearchFold:   true,
			FuncFilterInputRune: filterInput,
			HistoryLimit:        1000,
		})

		if err != nil {
			panic(err)
		}
		defer l.Close()

		for {
			line, err := l.Readline()
			if err == readline.ErrInterrupt {
				if len(line) == 0 {
					break
				} else {
					continue
				}
			} else if err == io.EOF {
				break
			}

			line = strings.TrimSpace(line)
			lex := lexer.New(stdlib + "\n\n" + line)
			p := parser.New(lex)
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
	} else {
		scanner := bufio.NewScanner(in)
		for {
			fmt.Print("> ")
			scanned := scanner.Scan()
			if !scanned {
				return
			}
			line := scanner.Text()
			l := lexer.New(stdlib + "\n\n" + line)
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
}
