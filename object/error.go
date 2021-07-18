package object

// Error wraps string and implements Object interface.
type Error struct {
	// Message contains the error-message we're wrapping
	Message string
}

// Type returns the type of this object.
func (e *Error) Type() Type {
	return ERROR_OBJ
}

// Inspect returns a string-representation of the given object.
func (e *Error) Inspect() string {
	return "ERROR: " + e.Message
}

// GetMethod returns a method against the object.
// (Built-in methods only.)
func (e *Error) GetMethod(string) BuiltinFunction {

	// There are no methods available upon a return-object.
	return nil
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
func (e *Error) ToInterface() interface{} {
	return "<ERROR>"
}
