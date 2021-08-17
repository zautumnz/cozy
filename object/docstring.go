// The implementation of our string-object.

package object

import (
	"hash/fnv"
	"sort"
	"strings"
)

// DocString wraps string and implements Object and Hashable interfaces.
type DocString struct {
	// Value holds the string value this object wraps.
	Value string
}

// Type returns the type of this object.
func (s *DocString) Type() Type {
	return DOCSTRING_OBJ
}

// Inspect returns an empty string;
// a docstring should only be printable when it's part of a block
func (s *DocString) Inspect() string {
	return ""
}

// HashKey returns a hash key for the given object.
func (s *DocString) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// GetMethod invokes a method against the object.
// (Built-in methods only.)
func (s *DocString) GetMethod(method string) BuiltinFunction {
	switch method {
	case "methods":
		return func(env *Environment, args ...Object) Object {
			static := []string{"methods"}
			dynamic := env.Names("string.")

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
func (s *DocString) ToInterface() interface{} {
	return s.Value
}

// Json returns a json-friendly string
func (s *DocString) Json(indent bool) string {
	return s.Inspect()
}
