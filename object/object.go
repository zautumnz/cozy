// Package object contains our core-definitions for objects.
package object

// Type describes the type of an object.
type Type string

// pre-defined constant Type
const (
	ARRAY_OBJ        = "ARRAY"
	BOOLEAN_OBJ      = "BOOLEAN"
	BUILTIN_OBJ      = "BUILTIN"
	DOCSTRING_OBJ    = "DOCSTRING"
	ERROR_OBJ        = "ERROR"
	FILE_OBJ         = "FILE"
	FLOAT_OBJ        = "FLOAT"
	FUNCTION_OBJ     = "FUNCTION"
	HASH_OBJ         = "HASH"
	INTEGER_OBJ      = "INTEGER"
	MODULE_OBJ       = "MODULE"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	STRING_OBJ       = "STRING"
)

// SystemTypesMap map system types by type name
var SystemTypesMap = map[Type]Object{
	ARRAY_OBJ:        &Array{},
	BOOLEAN_OBJ:      &Boolean{},
	BUILTIN_OBJ:      &Builtin{},
	DOCSTRING_OBJ:    &DocString{},
	ERROR_OBJ:        &Error{},
	FILE_OBJ:         &File{},
	FLOAT_OBJ:        &Float{},
	FUNCTION_OBJ:     &Function{},
	HASH_OBJ:         &Hash{},
	INTEGER_OBJ:      &Integer{},
	MODULE_OBJ:       &Module{},
	NULL_OBJ:         &Null{},
	RETURN_VALUE_OBJ: &ReturnValue{},
	STRING_OBJ:       &String{},
}

// Object is the interface that all of our various object-types must implmenet.
type Object interface {
	// Type returns the type of this object.
	Type() Type

	// Inspect returns a string-representation of the given object.
	Inspect() string

	// GetMethod invokes a method against the object.
	// (Built-in methods only.)
	GetMethod(method string) BuiltinFunction

	// ToInterface converts the given object to a "native" golang value,
	// which is required to ensure that we can use the object in our
	// `sprintf` or `printf` primitives.
	ToInterface() interface{}

	// Return a JSON-friendly string
	JSON(indent bool) string
}

// Hashable type can be hashed
type Hashable interface {
	// HashKey returns a hash key for the given object.
	HashKey() HashKey
}

// Iterable is an interface that some objects might support.
// If this interface is implemented then it will be possible to
// use the `foreach` function to iterate over the object. If
// the interface is not implemented then a run-time error will
// be generated instead.
type Iterable interface {
	// Reset the state of any previous iteration.
	Reset()

	// Get the next "thing" from the object being iterated
	// over.
	// The return values are the item which is to be returned
	// next, the index of that object, and finally a boolean
	// to say whether the function succeeded.
	// If the boolean value returned is false then that
	// means the iteration has completed and no further
	// items are available.
	Next() (Object, Object, bool)
}
