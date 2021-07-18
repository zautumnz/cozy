package object

import (
	"bytes"
	"sort"
	"strings"

	"github.com/zacanger/cozy/ast"
)

// Function wraps ast.Identifier array, ast.BlockStatement and Environment and implements Object interface.
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Defaults   map[string]ast.Expression
	Env        *Environment
	DocString  *ast.DocStringLiteral
}

// Type returns the type of this object.
func (f *Function) Type() Type {
	return FUNCTION_OBJ
}

// Inspect returns a string-representation of the given object.
func (f *Function) Inspect() string {
	var out bytes.Buffer
	parameters := make([]string, 0)
	for _, p := range f.Parameters {
		parameters = append(parameters, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(parameters, ", "))
	out.WriteString(") {\n")
	for _, s := range f.Body.Statements {
		out.WriteString(s.String() + "\n")
	}
	out.WriteString("}")
	return out.String()
}

// GetMethod returns a method against the object.
// (Built-in methods only.)
func (f *Function) GetMethod(method string) BuiltinFunction {
	if method == "methods" {
		return func(env *Environment, args ...Object) Object {
			static := []string{"methods"}
			dynamic := env.Names("function.")

			var names []string
			names = append(names, static...)
			for _, e := range dynamic {
				bits := strings.Split(e, ".")
				names = append(names, bits[1])
			}
			sort.Strings(names)

			result := make([]Object, len(names))
			for i, txt := range names {
				result[i] = &String{Value: txt}
			}
			return &Array{Elements: result}
		}
	}
	return nil
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
func (f *Function) ToInterface() interface{} {
	return "<FUNCTION>"
}

// Json returns a json-friendly string
func (f *Function) Json() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("\"")
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(
		strings.ReplaceAll(
			strings.Join(params, ", "), `"`, `\"`))
	out.WriteString(") {")
	out.WriteString(
		strings.ReplaceAll(f.Body.String(), `"`, `\"`))
	out.WriteString("}")
	out.WriteString("\"")

	return out.String()
}
