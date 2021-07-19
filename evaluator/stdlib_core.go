package evaluator

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/zacanger/cozy/lexer"
	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/parser"
)

// EvalFun evaluates a string containing cozy code
// Exported for use in timers
func EvalFun(env *object.Environment, args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch args[0].(type) {
	case *object.String:
		txt := args[0].(*object.String).Value

		// Lex the input
		l := lexer.New(txt)

		// parse it.
		p := parser.New(l)

		// If there are no errors
		program := p.ParseProgram()
		if len(p.Errors()) == 0 {
			// evaluate it, and return the output.
			return (Eval(program, env))
		}

		// Otherwise abort.  We should have try { } catch
		// to allow this kind of error to be caught in the future!
		fmt.Printf("Error parsing eval-string: %s", txt)
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}
	return newError("argument to `eval` not supported, got=%s",
		args[0].Type())
}

// exit a program.
func exitFun(args ...object.Object) object.Object {
	code := 0

	// Optionally an exit-code might be supplied as an argument
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case *object.Integer:
			code = int(arg.Value)
		case *object.Float:
			code = int(arg.Value)
		}
	}

	os.Exit(code)
	return &object.Integer{Value: int64(code)}
}

// convert a string to a float
func floatFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch args[0].(type) {
	case *object.String:
		input := args[0].(*object.String).Value
		i, err := strconv.Atoi(input)
		if err == nil {
			return &object.Float{Value: float64(i)}
		}
		return newError("Converting string '%s' to float failed %s", input, err.Error())

	case *object.Boolean:
		input := args[0].(*object.Boolean).Value
		if input {
			return &object.Float{Value: float64(1)}

		}
		return &object.Float{Value: float64(0)}
	case *object.Float:
		// noop
		return args[0]
	case *object.Integer:
		input := args[0].(*object.Integer).Value
		return &object.Float{Value: float64(input)}
	default:
		return newError("argument to `float` not supported, got=%s",
			args[0].Type())
	}
}

// convert a double/string to an int
func intFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch args[0].(type) {
	case *object.String:
		input := args[0].(*object.String).Value
		i, err := strconv.Atoi(input)
		if err == nil {
			return &object.Integer{Value: int64(i)}
		}
		return newError("Converting string '%s' to int failed %s", input, err.Error())

	case *object.Boolean:
		input := args[0].(*object.Boolean).Value
		if input {
			return &object.Integer{Value: 1}

		}
		return &object.Integer{Value: 0}
	case *object.Integer:
		// noop
		return args[0]
	case *object.Float:
		input := args[0].(*object.Float).Value
		return &object.Integer{Value: int64(input)}
	default:
		return newError("argument to `int` not supported, got=%s",
			args[0].Type())
	}
}

// length of item
func lenFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(utf8.RuneCountInString(arg.Value))}
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	default:
		return newError("argument to `len` not supported, got=%s",
			args[0].Type())
	}
}

// regular expression match
func matchFun(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2",
			len(args))
	}

	if args[0].Type() != object.STRING_OBJ {
		return newError("argument to `match` must be STRING, got %s",
			args[0].Type())
	}
	if args[1].Type() != object.STRING_OBJ {
		return newError("argument to `match` must be STRING, got %s",
			args[1].Type())
	}

	// Compile and match
	reg := regexp.MustCompile(args[0].(*object.String).Value)
	res := reg.FindStringSubmatch(args[1].(*object.String).Value)

	if len(res) > 0 {

		newHash := make(map[object.HashKey]object.HashPair)

		// If we get a match then the output is an array
		// First entry is the match, any additional parts
		// are the capture-groups.
		if len(res) > 1 {
			for i := 1; i < len(res); i++ {

				// Capture groups start at index 0.
				k := &object.Integer{Value: int64(i - 1)}
				v := &object.String{Value: res[i]}

				newHashPair := object.HashPair{Key: k, Value: v}
				newHash[k.HashKey()] = newHashPair

			}
		}

		return &object.Hash{Pairs: newHash}
	}

	// No match
	return &object.Boolean{Value: false}
}

// output a string to stdout
func printFun(args ...object.Object) object.Object {
	for _, arg := range args {
		fmt.Print(arg.Inspect() + " ")
	}
	fmt.Print("\n")
	return &object.Boolean{Value: true}
}

// printfFun is the implementation of our `printf` function.
func printfFun(args ...object.Object) object.Object {
	// Convert to the formatted version, via our `sprintf`
	// function.
	out := sprintfFun(args...)

	// If that returned a string then we can print it
	if out.Type() == object.STRING_OBJ {
		fmt.Print(out.(*object.String).Value)

	}

	return &object.Boolean{Value: true}
}

// sprintfFun is the implementation of our `sprintf` function.
func sprintfFun(args ...object.Object) object.Object {
	// We expect 1+ arguments
	if len(args) < 1 {
		return &object.String{Value: ""}
	}

	// Type-check
	if args[0].Type() != object.STRING_OBJ {
		return &object.String{Value: ""}
	}

	// Get the format-string.
	fs := args[0].(*object.String).Value

	// Convert the arguments to something go's sprintf
	// code will understand.
	argLen := len(args)
	fmtArgs := make([]interface{}, argLen-1)

	// Here we convert and assign.
	for i, v := range args[1:] {
		fmtArgs[i] = v.ToInterface()
	}

	// Call the helper
	out := fmt.Sprintf(fs, fmtArgs...)

	// And now return the value.
	return &object.String{Value: out}
}

func strFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}

	out := args[0].Inspect()
	return &object.String{Value: out}
}

// docstring of a function
func docFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch a := args[0].(type) {
	case *object.Function:
		return &object.String{Value: a.DocString.Value}

	default:
		return newError("argument to `doc` not supported, got=%s",
			args[0].Type())
	}
}

// type of an item
func typeFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch args[0].(type) {
	case *object.String:
		return &object.String{Value: "string"}
	case *object.Regexp:
		return &object.String{Value: "regexp"}
	case *object.Boolean:
		return &object.String{Value: "bool"}
	case *object.Builtin:
		return &object.String{Value: "builtin"}
	case *object.File:
		return &object.String{Value: "file"}
	case *object.Array:
		return &object.String{Value: "array"}
	case *object.Function:
		return &object.String{Value: "function"}
	case *object.Integer:
		return &object.String{Value: "integer"}
	case *object.Float:
		return &object.String{Value: "float"}
	case *object.Hash:
		return &object.String{Value: "hash"}
	case *object.Module:
		return &object.String{Value: "module"}
	case *object.DocString:
		return &object.String{Value: "docstring"}
	default:
		return newError("argument to `type` not supported, got=%s",
			args[0].Type())
	}
}

// flag("my-flag")
func flagFun(args ...object.Object) object.Object {
	// flag we're trying to retrieve
	name := args[0].(*object.String)
	found := false

	// Loop through all the arguments passed to the script
	// This is O(n), but performance is not a big deal
	for _, v := range os.Args {
		// If the flag was found in the previous argument...
		if found {
			// ...and the next one is another flag
			// means we're done parsing eg. --flag1 --flag2
			if strings.HasPrefix(v, "-") {
				break
			}

			// else return the next argument eg --flag1 something --flag2
			return &object.String{Value: v}
		}

		// try to parse the flag as key=value
		parts := strings.SplitN(v, "=", 2)
		// let's just take the left-side of the flag
		left := parts[0]

		// if the left side of the current argument corresponds
		// to the flag we're looking for (both in the form of "--flag" and "-flag")...
		if (len(left) > 1 && left[1:] == name.Value) ||
			(len(left) > 2 && left[2:] == name.Value) {
			if len(parts) > 1 {
				return &object.String{Value: parts[1]}
			}
			found = true
		}
	}

	// If the flag was found but we got here it means no value
	// was assigned to it, so default to true
	if found {
		return TRUE
	}

	// else a flag that's not found is false
	return FALSE
}

func init() {
	RegisterBuiltin("eval",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (EvalFun(env, args...))
		})
	RegisterBuiltin("exit",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (exitFun(args...))
		})
	RegisterBuiltin("int",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (intFun(args...))
		})
	RegisterBuiltin("float",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (floatFun(args...))
		})
	RegisterBuiltin("len",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (lenFun(args...))
		})
	RegisterBuiltin("match",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (matchFun(args...))
		})
	RegisterBuiltin("print",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (printFun(args...))
		})
	RegisterBuiltin("printf",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (printfFun(args...))
		})
	RegisterBuiltin("sprintf",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (sprintfFun(args...))
		})
	RegisterBuiltin("string",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (strFun(args...))
		})
	RegisterBuiltin("type",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (typeFun(args...))
		})
	RegisterBuiltin("doc",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (docFun(args...))
		})
	RegisterBuiltin("flag",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (flagFun(args...))
		})
}
