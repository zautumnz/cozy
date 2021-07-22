package evaluator

import (
	"fmt"
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

func timeTimeout(args ...object.Object) object.Object {
	var ms int64
	var f *object.Function
	switch t := args[0].(type) {
	case *object.Integer:
		ms = t.Value
	default:
		return newError("First argument to `time.timeout` should be integer!")
	}

	switch tt := args[1].(type) {
	case *object.Function:
		f = tt
	default:
		return newError("Second argument to `time.timeout should be function!`")
	}

	time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
		// TODO: eval the function here
		fmt.Println("here", f)
	})

	// TODO: use this return value to clear the timeout if need be
	return &object.Integer{Value: rand.Int63()}
}

func timeInterval(args ...object.Object) object.Object {
	var ms int64
	var f *object.Function
	switch t := args[0].(type) {
	case *object.Integer:
		ms = t.Value
	default:
		return newError("First argument to `time.interval` should be integer!")
	}

	switch tt := args[1].(type) {
	case *object.Function:
		f = tt
	default:
		return newError("Second argument to `time.interval should be function!`")
	}

	ticker := time.NewTicker(time.Duration(ms) * time.Millisecond)
	// TODO: this clear is what should get returned, not the random int, but how
	// to wrap it?
	clear := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				// TODO: eval the function here in another goroutine (go evalFn)
				fmt.Println("here", f)
			case <-clear:
				ticker.Stop()
				return
			}
		}
	}()
	time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
		fmt.Println("here", f)
	})

	// TODO: use this return value to clear the timeout if need be
	return &object.Integer{Value: rand.Int63()}
}

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
	RegisterBuiltin("time.interval",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (timeInterval(args...))
		})
	RegisterBuiltin("time.timeout",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (timeTimeout(args...))
		})
}
