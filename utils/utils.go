package utils

import (
	"os"
)

// IsRepl is used by the repl and environment
// to determine whether to exit and whether mutable variables
// are allowed at the top level
var IsRepl = false

// SetReplOrRun sets if the program is running in a repl or
// running code (either in a file or evaling a string);
// used below in ExitConditionally
func SetReplOrRun(rep bool) {
	IsRepl = rep
	if IsRepl {
		os.Setenv("KEAI_RUNNING_IN_REPL", "true")
	}
}

// ExitConditionally exits only if we're not currently in a REPL
func ExitConditionally(code int) {
	if !IsRepl {
		os.Exit(code)
	}
}
