package evaluator

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/zacanger/cozy/object"
)

// getenv() -> (Hash)
func envFun(args ...object.Object) object.Object {

	env := os.Environ()
	newHash := make(map[object.HashKey]object.HashPair)

	// If we get a match then the output is an array
	// First entry is the match, any additional parts
	// are the capture-groups.
	for i := 1; i < len(env); i++ {

		// Capture groups start at index 0.
		k := &object.String{Value: env[i]}
		v := &object.String{Value: os.Getenv(env[i])}

		newHashPair := object.HashPair{Key: k, Value: v}
		newHash[k.HashKey()] = newHashPair
	}

	return &object.Hash{Pairs: newHash}
}

// getenv("PATH") -> string
func getEnvFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	if args[0].Type() != object.STRING_OBJ {
		return newError("argument must be a string, got=%s",
			args[0].Type())
	}
	input := args[0].(*object.String).Value
	return &object.String{Value: os.Getenv(input)}
}

// setenv("PATH", "/home/z/bin:/usr/bin");
func setEnvFun(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	if args[0].Type() != object.STRING_OBJ {
		return newError("argument must be a string, got=%s",
			args[0].Type())
	}
	if args[1].Type() != object.STRING_OBJ {
		return newError("argument must be a string, got=%s",
			args[1].Type())
	}
	name := args[0].(*object.String).Value
	value := args[1].(*object.String).Value
	os.Setenv(name, value)
	return &object.Boolean{Value: true}
}

func sysExit(args ...object.Object) object.Object {
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

// Run a command and return a hash containing the result.
// `stderr`, `stdout`, and `error` will be the fields
func sysExec(args ...object.Object) object.Object {
	if len(args) < 1 {
		return newError("`sys.exec` wanted string, got invalid argument")
	}

	var command string
	switch c := args[0].(type) {
	case *object.String:
		command = c.Value
	default:
		return newError("`sys.exec` wanted string, got invalid argument")
	}

	if len(command) < 1 {
		return newError("`sys.exec` expected string, got invalid argument")
	}
	// split the command
	toExec := splitCommand(command)
	cmd := exec.Command(toExec[0], toExec[1:]...)

	// get the result
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()

	// If the command exits with a non-zero exit-code it
	// is regarded as a failure. Here we test for ExitError
	// to regard that as a non-failure.
	if err != nil && err != err.(*exec.ExitError) {
		fmt.Printf("Failed to run '%s' -> %s\n", command, err.Error())
		return &object.Error{Message: "Failed to run command!"}
	}

	// The result-objects to store in our hash.
	stdout := &object.String{Value: outb.String()}
	stderr := &object.String{Value: errb.String()}

	// Create keys
	stdoutKey := &object.String{Value: "stdout"}
	stdoutHash := object.HashPair{Key: stdoutKey, Value: stdout}

	stderrKey := &object.String{Value: "stderr"}
	stderrHash := object.HashPair{Key: stderrKey, Value: stderr}

	// Make a new hash, and populate it
	newHash := make(map[object.HashKey]object.HashPair)
	newHash[stdoutKey.HashKey()] = stdoutHash
	newHash[stderrKey.HashKey()] = stderrHash

	return &object.Hash{Pairs: newHash}
}

// Implemention of "args()" function.
func argsFun(args ...object.Object) object.Object {
	l := len(os.Args[1:])
	result := make([]object.Object, l)
	for i, txt := range os.Args[1:] {
		result[i] = &object.String{Value: txt}
	}
	return &object.Array{Elements: result}
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
	RegisterBuiltin("sys.getenv",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (getEnvFun(args...))
		})
	RegisterBuiltin("sys.setenv",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (setEnvFun(args...))
		})
	RegisterBuiltin("sys.environment",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (envFun(args...))
		})
	RegisterBuiltin("sys.exit",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (sysExit(args...))
		})
	RegisterBuiltin("sys.exec",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (sysExec(args...))
		})
	RegisterBuiltin("sys.flag",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (flagFun(args...))
		})
	RegisterBuiltin("sys.args",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (argsFun(args...))
		})
}
