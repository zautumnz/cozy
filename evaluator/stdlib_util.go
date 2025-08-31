package evaluator

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/zautumnz/cozy/object"
)

// convert a string to a float
func floatFn(args ...OBJ) OBJ {
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
		return NewError(
			"Converting string '%s' to float failed %s",
			input,
			err.Error(),
		)

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
func intFn(args ...OBJ) OBJ {
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
func lenFn(args ...OBJ) OBJ {
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
	case *object.Null:
		return &object.Integer{Value: 0}
	case *object.Hash:
		return &object.Integer{Value: int64(len(arg.Pairs))}
	default:
		return NewError("argument to `len` not supported, got=%s",
			args[0].Type())
	}
}

func strFn(args ...OBJ) OBJ {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}

	out := args[0].Inspect()
	return &object.String{Value: out}
}

// type of an item
func typeFn(args ...OBJ) OBJ {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	return &object.String{Value: strings.ToLower(string(args[0].Type()))}
}

func init() {
	RegisterBuiltin("util.int",
		func(env *ENV, args ...OBJ) OBJ {
			return intFn(args...)
		})
	RegisterBuiltin("util.float",
		func(env *ENV, args ...OBJ) OBJ {
			return floatFn(args...)
		})
	RegisterBuiltin("util.len",
		func(env *ENV, args ...OBJ) OBJ {
			return lenFn(args...)
		})
	RegisterBuiltin("util.string",
		func(env *ENV, args ...OBJ) OBJ {
			return strFn(args...)
		})
	RegisterBuiltin("util.type",
		func(env *ENV, args ...OBJ) OBJ {
			return typeFn(args...)
		})
}
