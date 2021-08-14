package object

// ReturnValue wraps Object and implements Object interface.
type ReturnValue struct {
	// Value is the object that is to be returned
	Value Object
}

// Type returns the type of this object.
func (r *ReturnValue) Type() Type {
	return RETURN_VALUE_OBJ
}

// Inspect returns a string-representation of the given object.
func (r *ReturnValue) Inspect() string {
	return r.Value.Inspect()
}

// GetMethod returns a method against the object.
// (Built-in methods only.)
func (r *ReturnValue) GetMethod(string) BuiltinFunction {
	// There are no methods available upon a return-object.
	return nil
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
func (r *ReturnValue) ToInterface() interface{} {
	return "<RETURN_VALUE>"
}

// Json returns a json-friendly string
func (r *ReturnValue) Json() string {
	return r.Inspect()
}
