package repl

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/zacanger/cozy/evaluator"
	"github.com/zacanger/cozy/lexer"
	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/parser"
	"github.com/zacanger/cozy/utils"
)

func getHistorySize() int {
	val := os.Getenv("COZY_HISTSIZE")
	l, e := strconv.Atoi(val)
	if e != nil || val == "" {
		return 1000
	}
	return l
}

// Start runs the REPL
func Start(in io.Reader, out io.Writer, stdlib string) {
	utils.SetReplOrRun(true)
	env := object.NewEnvironment()

	l, err := readline.NewEx(&readline.Config{
		Prompt:            "> ",
		HistoryFile:       os.Getenv("HOME") + "/.cozy_history",
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		HistoryLimit:      getHistorySize(),
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
			parser.PrintParserErrors(
				parser.ParserErrorsParams{Errors: p.Errors(), Out: out},
			)
			continue
		}
		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}
