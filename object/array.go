package object

import (
	"bytes"
	"sort"
	"strings"

	"github.com/zautumnz/cozy/token"
)

// Array wraps Object array and implements Object interface.
type Array struct {
	Token token.Token

	// Elements holds the individual members of the array we're wrapping.
	Elements []Object

	// offset holds our iteration-offset.
	offset int

	// special arr when used for ... args
	IsCurrentArgs bool
}

// Type returns the type of this object.
func (ao *Array) Type() Type {
	return ARRAY_OBJ
}

// Inspect returns a string-representation of the given object.
func (ao *Array) Inspect() string {
	var out bytes.Buffer
	elements := make([]string, 0)
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// GetMethod returns a method against the object.
// (Built-in methods only.)
func (ao *Array) GetMethod(method string) BuiltinFunction {
	switch method {
	case "append":
		return func(env *Environment, args ...Object) Object {
			arr := ao
			length := len(arr.Elements)
			newElements := make([]Object, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[0]
			return &Array{Elements: newElements}
		}
	case "methods":
		return func(env *Environment, args ...Object) Object {
			static := []string{"methods", "append"}
			dynamic := env.Names("array.")

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

// Reset implements the Iterable interface, and allows the contents
// of the array to be reset to allow re-iteration.
func (ao *Array) Reset() {
	ao.offset = 0
}

// Next implements the Iterable interface, and allows the contents
// of our array to be iterated over.
func (ao *Array) Next() (Object, Object, bool) {
	if ao.offset < len(ao.Elements) {
		ao.offset++

		element := ao.Elements[ao.offset-1]
		return element, &Integer{Value: int64(ao.offset - 1)}, true
	}

	return nil, &Integer{Value: 0}, false
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
func (ao *Array) ToInterface() interface{} {
	return "<ARRAY>"
}

// JSON returns a json-friendly string
func (ao *Array) JSON(indent bool) string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.JSON(indent))
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	if indent {
		return indentJSON(out.String())
	}

	return out.String()
}
