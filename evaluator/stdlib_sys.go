package evaluator

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/utils"
)

// Split a line of text into tokens, but keep anything "quoted"
// together..
// So this input:
//
//	/bin/sh -c "ls /etc"
//
// Would give output of the form:
//
//	/bin/sh
//	-c
//	ls /etc
func splitCommand(input string) []string {
	// This does the split into an array
	r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
	res := r.FindAllString(input, -1)

	// However the resulting pieces might be quoted.
	// So we have to remove them, if present.
	var result []string
	for _, e := range res {
		result = append(result, trimQuotes(e, '"'))
	}
	return result
}

// environment() -> (Hash)
func envFn(args ...OBJ) OBJ {
	env := os.Environ()
	vals := make(StringObjectMap)

	// If we get a match then the output is an array
	// First entry is the match, any additional parts
	// are the capture-groups.
	for i := 1; i < len(env); i++ {
		vals[env[i]] = &object.String{Value: os.Getenv(env[i])}
	}
	return NewHash(vals)
}

// getenv("PATH") -> string
func getEnvFn(args ...OBJ) OBJ {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	if args[0].Type() != object.STRING_OBJ {
		return NewError("argument must be a string, got=%s",
			args[0].Type())
	}
	input := args[0].(*object.String).Value
	return &object.String{Value: os.Getenv(input)}
}

// setenv("PATH", "/home/z/bin:/usr/bin");
func setEnvFn(args ...OBJ) OBJ {
	if len(args) != 2 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	if args[0].Type() != object.STRING_OBJ {
		return NewError("argument must be a string, got=%s",
			args[0].Type())
	}
	if args[1].Type() != object.STRING_OBJ {
		return NewError("argument must be a string, got=%s",
			args[1].Type())
	}
	name := args[0].(*object.String).Value
	value := args[1].(*object.String).Value
	os.Setenv(name, value)
	return NULL
}

func sysExit(args ...OBJ) OBJ {
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

	utils.ExitConditionally(code)
	return &object.Integer{Value: int64(code)}
}

// Run a command and return a hash containing the result.
// `stderr`, `stdout`, and `error` will be the fields
func sysExec(args ...OBJ) OBJ {
	if len(args) < 1 {
		return NewError("`sys.exec` wanted string, got invalid argument")
	}

	var command string
	switch c := args[0].(type) {
	case *object.String:
		command = c.Value
	default:
		return NewError("`sys.exec` wanted string, got invalid argument")
	}

	if len(command) < 1 {
		return NewError("`sys.exec` expected string, got invalid argument")
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

	return NewHash(StringObjectMap{
		"stdout": stdout,
		"stderr": stderr,
	})
}

// Implemention of "args()" function.
func argsFn(args ...OBJ) OBJ {
	l := len(os.Args[1:])
	result := make([]OBJ, l)
	for i, txt := range os.Args[1:] {
		result[i] = &object.String{Value: txt}
	}
	return &object.Array{Elements: result}
}

// flag("my-flag")
func flagFn(args ...OBJ) OBJ {
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

func cdFn(args ...OBJ) OBJ {
	switch a := args[0].(type) {
	case *object.String:
		os.Chdir(a.Value)
	default:
		return NewError("cd expected string argument!")
	}
	return NULL
}

func infoFn(args ...OBJ) OBJ {
	return NewHash(StringObjectMap{
		"os":   &object.String{Value: runtime.GOOS},
		"arch": &object.String{Value: runtime.GOARCH},
		"cpus": &object.Integer{Value: int64(runtime.NumCPU())},
	})
}

func init() {
	RegisterBuiltin("sys.getenv",
		func(env *ENV, args ...OBJ) OBJ {
			return getEnvFn(args...)
		})
	RegisterBuiltin("sys.setenv",
		func(env *ENV, args ...OBJ) OBJ {
			return setEnvFn(args...)
		})
	RegisterBuiltin("sys.environment",
		func(env *ENV, args ...OBJ) OBJ {
			return envFn(args...)
		})
	RegisterBuiltin("sys.exit",
		func(env *ENV, args ...OBJ) OBJ {
			return sysExit(args...)
		})
	RegisterBuiltin("sys.exec",
		func(env *ENV, args ...OBJ) OBJ {
			return sysExec(args...)
		})
	RegisterBuiltin("sys.flag",
		func(env *ENV, args ...OBJ) OBJ {
			return flagFn(args...)
		})
	RegisterBuiltin("sys.args",
		func(env *ENV, args ...OBJ) OBJ {
			return argsFn(args...)
		})
	RegisterBuiltin("sys.cd",
		func(env *ENV, args ...OBJ) OBJ {
			return cdFn(args...)
		})
	RegisterBuiltin("sys.info",
		func(env *ENV, args ...OBJ) OBJ {
			return infoFn(args...)
		})
}
