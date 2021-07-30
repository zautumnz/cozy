package evaluator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/zacanger/cozy/lexer"
	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/parser"
	"github.com/zacanger/cozy/utils"
)

// Evaluate a string containing cozy code
// This creates basically a whole new instance of cozy,
// which is inefficient, but it's the same thing we do when
// evaling modules and working with string interpolation.
func evalFun(env *object.Environment, args ...object.Object) object.Object {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
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

		// Otherwise abort. We should have try { } catch
		// to allow this kind of error to be caught in the future!
		fmt.Printf("Error parsing eval-string: %s", txt)
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		utils.ExitConditionally(1)
	}
	return NewError("argument to `eval` not supported, got=%s",
		args[0].Type())
}

// convert a string to a float
func floatFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch args[0].(type) {
	case *object.String:
		input := args[0].(*object.String).Value
		i, err := strconv.Atoi(input)
		if err == nil {
			return &object.Float{Value: float64(i)}
		}
		return NewError("Converting string '%s' to float failed %s", input, err.Error())

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
		return NewError("argument to `float` not supported, got=%s",
			args[0].Type())
	}
}

// convert a double/string to an int
func intFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch args[0].(type) {
	case *object.String:
		input := args[0].(*object.String).Value
		i, err := strconv.Atoi(input)
		if err == nil {
			return &object.Integer{Value: int64(i)}
		}
		return NewError("Converting string '%s' to int failed %s", input, err.Error())

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
		return NewError("argument to `int` not supported, got=%s",
			args[0].Type())
	}
}

// length of item
func lenFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(utf8.RuneCountInString(arg.Value))}
	case *object.DocString:
		return &object.Integer{Value: int64(utf8.RuneCountInString(arg.Value))}
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	case *object.Hash:
		return &object.Integer{Value: int64(len(arg.Pairs))}
	default:
		return NewError("argument to `len` not supported, got=%s",
			args[0].Type())
	}
}

// regular expression match
func matchFun(args ...object.Object) object.Object {
	if len(args) != 2 {
		return NewError("wrong number of arguments. got=%d, want=2",
			len(args))
	}

	if args[0].Type() != object.STRING_OBJ {
		return NewError("argument to `match` must be STRING, got %s",
			args[0].Type())
	}
	if args[1].Type() != object.STRING_OBJ {
		return NewError("argument to `match` must be STRING, got %s",
			args[1].Type())
	}

	// Compile and match
	reg := regexp.MustCompile(args[0].(*object.String).Value)
	res := reg.FindStringSubmatch(args[1].(*object.String).Value)

	if len(res) > 0 {
		newArray := make([]object.Object, len(res))

		// If we get a match then the output is an array
		// First entry is the match, any additional parts
		// are the capture-groups.
		if len(res) > 1 {
			for i, v := range res {
				newArray[i] = &object.String{Value: v}
			}
		}

		return &object.Array{Elements: newArray}
	}

	// No match
	return &object.Array{Elements: make([]object.Object, 0)}
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
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}

	out := args[0].Inspect()
	return &object.String{Value: out}
}

// type of an item
func typeFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	return &object.String{Value: strings.ToLower(string(args[0].Type()))}
}

func init() {
	RegisterBuiltin("eval",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (evalFun(env, args...))
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
}
