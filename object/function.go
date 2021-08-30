package object

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/zacanger/cozy/ast"
)

var stringifiedAnonymousFunctionMap map[string]int

func init() {
	stringifiedAnonymousFunctionMap = make(map[string]int)
}

// Function wraps ast.Identifier array, ast.BlockStatement and Environment
// and implements Object interface.
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Defaults   map[string]ast.Expression
	Env        *Environment
	DocString  *ast.DocStringLiteral
	Name       string
}

func (f *Function) stringify() string {
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
		out.WriteString(s.String())
	}
	out.WriteString("}")
	return out.String()
}

func (f *Function) getNameOrDefault() string {
	if f.Name != "" {
		return "FN_" + f.Name
	}

	n := 1
	if stringifiedAnonymousFunctionMap[f.stringify()] != 0 {
		n = stringifiedAnonymousFunctionMap[f.stringify()]
	} else {
		x := len(stringifiedAnonymousFunctionMap) + 1
		stringifiedAnonymousFunctionMap[f.stringify()] = x
		n = x
	}

	return "ANON_FN_" + fmt.Sprint(n)
}

// Type returns the type of this object.
func (f *Function) Type() Type {
	return FUNCTION_OBJ
}

// Inspect returns a string-representation of the given object.
func (f *Function) Inspect() string {
	return f.getNameOrDefault()
}

// GetMethod returns a method against the object.
// (Built-in methods only.)
func (f *Function) GetMethod(method string) BuiltinFunction {
	if method == "methods" {
		return func(env *Environment, args ...Object) Object {
			static := []string{"methods", "doc", "name"}
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

	if method == "doc" {
		return func(env *Environment, args ...Object) Object {
			if f.DocString != nil {
				return &String{Value: f.DocString.Value}
			}
			return &String{Value: ""}
		}
	}

	if method == "name" {
		return func(env *Environment, args ...Object) Object {
			return &String{Value: f.getNameOrDefault()}
		}

	}

	return nil
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
func (f *Function) ToInterface() interface{} {
	return "<FUNCTION>"
}

// JSON returns a json-friendly string
func (f *Function) JSON(indent bool) string {
	return f.Inspect()
}
