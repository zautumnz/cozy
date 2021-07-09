// Package object contains our core-definitions for objects.
package object

// Type describes the type of an object.
type Type string

// pre-defined constant Type
const (
	ARRAY_OBJ        = "ARRAY"
	BOOLEAN_OBJ      = "BOOLEAN"
	BUILTIN_OBJ      = "BUILTIN"
	ERROR_OBJ        = "ERROR"
	FILE_OBJ         = "FILE"
	FLOAT_OBJ        = "FLOAT"
	FUNCTION_OBJ     = "FUNCTION"
	HASH_OBJ         = "HASH"
	INTEGER_OBJ      = "INTEGER"
	MACRO_OBJ        = "MACRO"
	MODULE_OBJ       = "MODULE"
	NULL_OBJ         = "NULL"
	QUOTE_OBJ        = "QUOTE"
	REGEXP_OBJ       = "REGEXP"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	STRING_OBJ       = "STRING"
)

// Object is the interface that all of our various object-types must implmenet.
type Object interface {

	// Type returns the type of this object.
	Type() Type

	// Inspect returns a string-representation of the given object.
	Inspect() string

	// InvokeMethod invokes a method against the object.
	// (Built-in methods only.)
	InvokeMethod(method string, env Environment, args ...Object) Object

	// ToInterface converts the given object to a "native" golang value,
	// which is required to ensure that we can use the object in our
	// `sprintf` or `printf` primitives.
	ToInterface() interface{}
}

// Hashable type can be hashed
type Hashable interface {

	// HashKey returns a hash key for the given object.
	HashKey() HashKey
}

// Iterable is an interface that some objects might support.
//
// If this interface is implemented then it will be possible to
// use the `foreach` function to iterate over the object.  If
// the interface is not implemented then a run-time error will
// be generated instead.
type Iterable interface {

	// Reset the state of any previous iteration.
	Reset()

	// Get the next "thing" from the object being iterated
	// over.
	//
	// The return values are the item which is to be returned
	// next, the index of that object, and finally a boolean
	// to say whether the function succeeded.
	//
	// If the boolean value returned is false then that
	// means the iteration has completed and no further
	// items are available.
	Next() (Object, Object, bool)
}
