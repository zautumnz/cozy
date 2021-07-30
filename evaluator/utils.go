package evaluator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/zacanger/cozy/lexer"
	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/parser"
	"github.com/zacanger/cozy/utils"
)

var searchPaths []string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error getting cwd: %s", err)
	}

	if e := os.Getenv("COZYPATH"); e != "" {
		tokens := strings.Split(e, ":")
		for _, token := range tokens {
			addPath(token) // ignore errors
		}
	} else {
		searchPaths = append(searchPaths, cwd)
	}
}

func addPath(path string) error {
	path = os.ExpandEnv(filepath.Clean(path))
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	searchPaths = append(searchPaths, absPath)
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FindModule finds a module based on name, used by the evaluator
func FindModule(name string) string {
	basename := fmt.Sprintf("%s.cz", name)
	for _, p := range searchPaths {
		filename := filepath.Join(p, basename)
		if exists(filename) {
			return filename
		}
	}
	return ""
}

// IsNumber checks to see if a value is a number
func IsNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// Interpolate (str, env)
// return input string with $vars interpolated from environment
func Interpolate(str string, env *object.Environment) string {
	// Match all strings preceded by {{
	// re := regexp.MustCompile(`(\\)?\(\{\{)([a-zA-Z_0-9]{1,})(\}\})`)
	// re := regexp.MustCompile(`(\\)?\$(\{)([a-zA-Z_0-9]{1,})(\})`)
	re := regexp.MustCompile(`(\\)?(\{\{)(.*?)(\}\})`)
	str = re.ReplaceAllStringFunc(str, func(m string) string {
		// If the string starts with a backslash, that's an escape, so we should
		// replace it with the remaining portion of the match. \{{VAR}} becomes
		// {{VAR}}
		if string(m[0]) == "\\" {
			return m[1:]
		}

		varName := ""

		// If you type a variable wrong, forgetting the closing bracket, we
		// simply return it to you: eg "my {{variable"

		if m[len(m)-1] != '}' || m[len(m)-2] != '}' {
			return m
		}

		varName = m[2 : len(m)-2]

		v, ok := env.Get(varName)

		// The variable might be an index expression
		if !ok {
			// Basically just spinning up a whole new instance of cozy; very
			// inefficient, but it's the same thing we do on every module
			// require and eval() calls.
			l := lexer.New(string(varName))
			p := parser.New(l)
			program := p.ParseProgram()
			macroEnv := object.NewEnvironment()
			DefineMacros(program, macroEnv)
			expanded := ExpandMacros(program, macroEnv)
			evaluated := Eval(expanded, env)
			if evaluated != nil {
				return evaluated.Inspect()
			}

			// Still no match found, so return an empty string
			return ""
		}

		return v.Inspect()
	})

	return str
}

// NewErrorWithExitCode takes an exit code, format string, and variables for
// the string. It prints the error, optionally exits, and otherwise returns the
// error.
// TODO: this isn't used anywhere yet, but could be used in places where both an
// error is created and/or printed and also there's an ExitConditionally call
func NewErrorWithExitCode(code int, format string, a ...interface{}) *object.Error {
	message := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, message+"\n")
	utils.ExitConditionally(code)
	return &object.Error{Message: message, Code: &code}
}

// NewError prints and returns an error
func NewError(format string, a ...interface{}) *object.Error {
	message := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, message+"\n")
	return &object.Error{Message: message}
}
