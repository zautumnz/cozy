package object

import (
	"github.com/zacanger/cozy/ast"
)

// Quote implements Object for a Quote (for macros)
type Quote struct {
	Node ast.Node
}

// GetMethod invokes a method against the object.
func (q *Quote) GetMethod(method string) BuiltinFunction {
	return nil
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
func (q *Quote) ToInterface() interface{} {
	return "<QUOTE>"
}

// Type returns the type of this object
func (q *Quote) Type() Type {
	return QUOTE_OBJ
}

// Inspect returns a string representing the object
func (q *Quote) Inspect() string {
	return "QUOTE(" + q.Node.String() + ")"
}

// Json returns a json-friendly string
func (q *Quote) Json() string {
	return q.Inspect()
}
