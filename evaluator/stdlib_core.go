package evaluator

import (
	"context"
	"math/rand"
	"regexp"

	"github.com/zacanger/cozy/object"
)

// Async and await are on code from github.com/Gurpartap/async, Apache 2.0
// licensed.

type ValueFuture interface {
	Await() interface{}
}

type valueFuture struct {
	await func(ctx context.Context) interface{}
}

func (f valueFuture) Await() interface{} {
	return f.await(context.Background())
}

func Async(f func() interface{}) ValueFuture {
	var result interface{}
	c := make(chan struct{}, 1)
	go func() {
		defer close(c)
		result = f()
	}()
	return valueFuture{
		await: func(ctx context.Context) interface{} {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c:
				return result
			}
		},
	}
}

var asyncFunctions = make(map[int64]ValueFuture)

func awaitFn(env *ENV, args ...OBJ) OBJ {
	var res interface{}
	var err error
	switch t := args[0].(type) {
	case *object.Integer:
		f := asyncFunctions[t.Value]
		res = f.Await()
	default:
		return NewError("Expected async function id, got %s", args[0].Type())
	}

	if err != nil {
		return &object.Error{Message: err.Error()}
	}

	switch x := res.(type) {
	case OBJ:
		return x
	default:
		return NewError("Something went wrong in await!")

	}
}

func asyncFn(env *ENV, args ...OBJ) OBJ {
	x := Async(func() interface{} {
		return ApplyFunction(env, args[0], make([]OBJ, 0))
	})

	fnID := rand.Int63()
	asyncFunctions[fnID] = x
	return &object.Integer{Value: fnID}
}

func backgroundFn(env *ENV, args ...OBJ) OBJ {
	switch a := args[0].(type) {
	case *object.Function:
		go func() {
			ApplyFunction(env, a, make([]OBJ, 0))
		}()
		return NULL
	default:
		return NewError("background expected function arg!")
	}
}

// regular expression match
func matchFn(args ...OBJ) OBJ {
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
		newArray := make([]OBJ, len(res))

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
	return &object.Array{Elements: make([]OBJ, 0)}
}

func init() {
	RegisterBuiltin("core.match",
		func(env *ENV, args ...OBJ) OBJ {
			return matchFn(args...)
		})
	RegisterBuiltin("core.async",
		func(env *ENV, args ...OBJ) OBJ {
			return asyncFn(env, args...)
		})
	RegisterBuiltin("core.await",
		func(env *ENV, args ...OBJ) OBJ {
			return awaitFn(env, args...)
		})
	RegisterBuiltin("core.background",
		func(env *ENV, args ...OBJ) OBJ {
			return backgroundFn(env, args...)
		})
}
