package evaluator

import (
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
	return &object.Null{}
}

// TODO: time.timeout, time.interval
func init() {
	RegisterBuiltin("time.sleep",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (timeSleep(args...))
		})
}
