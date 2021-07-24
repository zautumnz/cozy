package object

import (
	"fmt"
	"os"
	"strings"
)

// Environment stores our functions, variables, constants, etc.
type Environment struct {
	// store holds variables, including functions.
	store map[string]Object

	// readonly marks names as read-only.
	readonly map[string]bool

	// outer holds any parent environment. Our env. allows
	// nesting to implement scope.
	outer *Environment

	// permit stores the names of variables we can set in this
	// environment, if any
	permit []string

	// Args used when creating this env. Used in ...
	CurrentArgs []Object
}

// NewEnvironment creates new environment
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	r := make(map[string]bool)
	return &Environment{store: s, readonly: r, outer: nil}
}

// NewEnclosedEnvironment create new environment by outer parameter
func NewEnclosedEnvironment(outer *Environment, args []Object) *Environment {
	env := NewEnvironment()
	env.outer = outer
	env.CurrentArgs = args
	return env
}

// NewTemporaryScope creates a temporary scope where some values
// are ignored.
// This is used as a sneaky hack to allow `foreach` to access all
// global values as if they were local, but prevent the index/value
// keys from persisting.
func NewTemporaryScope(outer *Environment, keys []string) *Environment {
	env := NewEnvironment()
	env.outer = outer
	env.permit = keys
	return env
}

// Names returns the names of every known-value with the
// given prefix.
// This function is used by `invokeMethod` to get the methods
// associated with a particular class-type.
func (e *Environment) Names(prefix string) []string {
	var ret []string

	for key := range e.store {
		if strings.HasPrefix(key, prefix) {
			ret = append(ret, key)
		}

		// Functions with an "object." prefix are available
		// to all object-methods.
		if strings.HasPrefix(key, "object.") {
			ret = append(ret, key)
		}
	}
	return ret
}

// TODO: this isn't quite right, but when it is,
// use it to ensure we're walking all the way up the chain rather than
// just one level up
func (e *Environment) lookupInEnvs(name string) (Object, Environment) {
	currentEnv := e

	for currentEnv != nil {
		if currentEnv.store[name] != nil {
			return currentEnv.store[name], *currentEnv
		}

		if currentEnv.outer == nil {
			break
		} else {
			currentEnv = currentEnv.outer
			continue
		}
	}

	return nil, *currentEnv
}

// Get returns the value of a given variable, by name.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set stores the value of a variable, by name.
func (e *Environment) Set(name string, val Object) Object {
	cur := e.store[name]

	if (cur != nil && e.readonly[name]) || (e.outer != nil && e.outer.store[name] != nil && e.outer.readonly[name]) {
		fmt.Printf("Attempting to modify '%s' denied; it was defined as a constant.\n", name)
		os.Exit(3)
	}

	// Store the (updated) value.

	// This chunk is used for temporary environments (regex and loops)
	if len(e.permit) > 0 {
		for _, v := range e.permit {
			// we're permitted to store this variable
			if v == name {
				e.store[name] = val
				return val
			}
		}
		// ok we're not permitted, we must store in the parent
		if e.outer != nil {
			return e.outer.Set(name, val)
		}

		// Otherwise something is very broken!
		fmt.Printf("Something is broken with scope!\n")
		os.Exit(5)
	}

	// Otherwise we're just in a regular block
	// First check to see if this is a shadowed var
	if e.outer != nil && e.outer.store[name] != nil {
		return e.outer.Set(name, val)
	}

	// ...and otherwise, just store it in the current scope
	e.store[name] = val
	return val
}

// SetLet sets the value of a constant by name.
// TODO: change this to just use Set, and store Constant as an optional
// bool on all values
func (e *Environment) SetLet(name string, val Object) Object {

	// store the value
	e.store[name] = val

	// flag as read-only.
	e.readonly[name] = true

	return val
}

// ExportedHash returns a new Hash with the names and values of every publically
// exported binding in the environment; that is, every top-level binding (not
// in a block).
// This is used by the module import system to wrap up the
// evaulated module into an object.
func (e *Environment) ExportedHash() *Hash {
	pairs := make(map[HashKey]HashPair)
	for k, v := range e.store {
		s := &String{Value: k}
		pairs[s.HashKey()] = HashPair{Key: s, Value: v}
	}
	return &Hash{Pairs: pairs}
}
