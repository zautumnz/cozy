package utils

import (
	"os"
)

var isRepl = false

// SetReplOrRun sets if the program is running in a repl or
// running code (either in a file or evaling a string);
// used below in ExitConditionally
func SetReplOrRun(rep bool) {
	isRepl = rep
	if isRepl {
		os.Setenv("COZY_RUNNING_IN_REPL", "true")
	}
}

// ExitConditionally exits only if we're not currently in a REPL
func ExitConditionally(code int) {
	if !isRepl {
		os.Exit(code)
	}
}
