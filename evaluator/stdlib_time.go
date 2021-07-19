package evaluator

import (
	"math/rand"
	"time"

	"github.com/zacanger/cozy/object"
)

func timeSleep(args ...object.Object) object.Object {
	switch arg := args[0].(type) {
	case *object.Integer:
		time.Sleep(time.Duration(arg.Value) * time.Millisecond)
	default:
		return newError("argument to `time.sleep` not supported, got=%s", arg.Type())
	}

	// TODO: use this return value to clear the sleep if need be
	return &object.Integer{Value: rand.Int63()}
}

func timeUnix(args ...object.Object) object.Object {
	return &object.Float{Value: float64(time.Now().UnixNano() / 1000000)}
}

func timeUtc(args ...object.Object) object.Object {
	return &object.String{Value: time.Now().Format(time.RFC3339)}
}

// TODO:
// time.timeout
// time.interval
func init() {
	RegisterBuiltin("time.sleep",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (timeSleep(args...))
		})
	RegisterBuiltin("time.unix",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (timeUnix(args...))
		})
	RegisterBuiltin("time.utc",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (timeUtc(args...))
		})
}
