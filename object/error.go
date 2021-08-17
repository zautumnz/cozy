package object

import (
	"fmt"
)

// Error wraps string and implements Object interface.
type Error struct {
	// Message contains the error-message we're wrapping
	Message string

	// Optional exit code
	Code *int

	// If we're calling the error() builtin
	BuiltinCall bool

	// Any extra data
	Data string
}

// Type returns the type of this object.
func (e *Error) Type() Type {
	return ERROR_OBJ
}

// Inspect returns a string-representation of the given object.
func (e *Error) Inspect() string {
	msg := "ERROR: " + e.Message
	if e.Code != nil {
		msg += "; CODE: " + fmt.Sprint(*e.Code)
	}
	if e.Data != "" {
		msg += "; DATA: " + fmt.Sprint(e.Data)
	}
	return msg
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

// Json returns a json-friendly string
func (e *Error) Json(indent bool) string {
	s := `{"error":"` + escapeQuotes(e.Message) + `"`
	if e.Code != nil {
		s += `,"code":` + fmt.Sprint(*e.Code)
	}
	if e.Data != "" {
		s += `,"data":` + fmt.Sprint(*e.Code)
	}

	s += "}"

	if indent {
		return indentJSON(s)
	}
	return s
}
