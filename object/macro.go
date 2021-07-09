package object

import (
	"bytes"
	"strings"

	"github.com/zacanger/cozy/ast"
)

// Macro implements a macro (used with Quote)
type Macro struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

// InvokeMethod invokes a method against the object.
func (m *Macro) InvokeMethod(
	method string,
	env Environment, args ...Object,
) Object {
	return nil
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
func (m *Macro) ToInterface() interface{} {
	return "<MACRO>"
}

// Type returns the type of this object
func (m *Macro) Type() Type {
	return MACRO_OBJ
}

// Inspect returns a string representing the object
func (m *Macro) Inspect() string {
	var out bytes.Buffer

	params := make([]string, 0)
	for _, p := range m.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("macro")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(m.Body.String())
	out.WriteString("\n}")

	return out.String()
}
