package evaluator

import (
	"math/rand"
	"time"

	"github.com/zacanger/cozy/object"
)

func timeSleep(args ...object.Object) object.Object {
	var ms int64
	switch arg := args[0].(type) {
	case *object.Integer:
		ms = arg.Value
	default:
		return newError("argument to `time.sleep` not supported, got=%s", arg.Type())
	}

	time.Sleep(time.Duration(ms) * time.Millisecond)
	return &object.Integer{Value: ms}
}

func timeUnix(args ...object.Object) object.Object {
	return &object.Float{Value: float64(time.Now().UnixNano() / 1000000)}
}

func timeUtc(args ...object.Object) object.Object {
	return &object.String{Value: time.Now().Format(time.RFC3339)}
}

var intervalIDs = make(map[int64]chan bool)
var timeoutIDs = make(map[int64]bool)

func timeTimeout(env *object.Environment, args ...object.Object) object.Object {
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

	timeoutID := rand.Int63()
	timeoutIDs[timeoutID] = false
	time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
		v, ok := timeoutIDs[timeoutID]
		if ok && !v {
			ApplyFunction(env, f, make([]object.Object, 0))
		}
	})

	return &object.Integer{Value: timeoutID}
}

func timeInterval(env *object.Environment, args ...object.Object) object.Object {
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
	clear := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				go ApplyFunction(env, f, make([]object.Object, 0))
			case <-clear:
				ticker.Stop()
				return
			}
		}
	}()

	intervalID := rand.Int63()
	intervalIDs[intervalID] = clear
	return &object.Integer{Value: intervalID}
}

func timeCancel(args ...object.Object) object.Object {
	switch t := args[0].(type) {
	case *object.Integer:
		if intervalIDs[t.Value] != nil {
			intervalIDs[t.Value] <- true
		} else {
			_, ok := timeoutIDs[t.Value]
			if ok {
				timeoutIDs[t.Value] = true
			}
		}
	default:
		return newError("Expected timerid, got %s", args[0].Type())
	}

	return &object.Boolean{Value: true}
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
			return (timeInterval(env, args...))
		})
	RegisterBuiltin("time.timeout",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (timeTimeout(env, args...))
		})
	RegisterBuiltin("time.cancel",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (timeCancel(args...))
		})
}
