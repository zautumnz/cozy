package evaluator

import (
	"fmt"

	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/utils"
)

// These stdlib functions aren't scoped/namespaced

// panic
func panicFn(args ...object.Object) object.Object {
	switch e := args[0].(type) {
	case *object.Error:
		c := 1
		fmt.Println(e.Message)
		if e.Code != nil {
			c = int(*e.Code)
		}
		utils.ExitConditionally(c)
	default:
		return NewError("panic expected an error!")
	}
	return NULL
}

// error
func errorFn(args ...object.Object) object.Object {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch t := args[0].(type) {
	case *object.String:
		return &object.Error{Message: t.Value, BuiltinCall: true}
	case *object.Hash:
		msgStr := &object.String{Value: "message"}
		codeStr := &object.String{Value: "code"}
		dataStr := &object.String{Value: "data"}
		msg := t.Pairs[msgStr.HashKey()].Value
		code := t.Pairs[codeStr.HashKey()].Value
		data := t.Pairs[dataStr.HashKey()].Value
		e := &object.Error{BuiltinCall: true}
		if msg != nil {
			switch m := msg.(type) {
			case *object.String:
				e.Message = m.Value
			default:
				return NewError("error.message should be string!")
			}
		}
		if code != nil {
			switch c := code.(type) {
			case *object.Integer:
				cc := int(c.Value)
				e.Code = &cc
			default:
				return NewError("error.code should be integer!")
			}
		}
		if data != nil {
			e.Data = data.Json()
		}
		return e
	default:
		return NewError("error() expected a string or hash!")
	}
}

// output a string to stdout
func printFn(args ...object.Object) object.Object {
	for _, arg := range args {
		fmt.Print(arg.Inspect() + " ")
	}
	fmt.Print("\n")
	return NULL
}

func init() {
	RegisterBuiltin("print",
		func(env *object.Environment, args ...object.Object) object.Object {
			return printFn(args...)
		})
	RegisterBuiltin("error",
		func(env *object.Environment, args ...object.Object) object.Object {
			return errorFn(args...)
		})
	RegisterBuiltin("panic",
		func(env *object.Environment, args ...object.Object) object.Object {
			return panicFn(args...)
		})
}
