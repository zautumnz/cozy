package object

import (
	"fmt"
	"sort"
	"strings"
)

// Module is the module type used to represent a collection of vars
type Module struct {
	Name  string
	Attrs Object
}

func (m *Module) Bool() bool {
	return true
}

func (m *Module) Compare(other Object) int {
	return 1
}

func (m *Module) String() string {
	return m.Inspect()
}

// Type returns the type of the object
func (m *Module) Type() Type { return MODULE_OBJ }

// Inspect returns a stringified version of the object for debugging
func (m *Module) Inspect() string { return fmt.Sprintf("<MODULE '%s'>", m.Name) }

// GetMethod invokes a method against the object.
// (Built-in methods only.)
func (m *Module) GetMethod(method string) BuiltinFunction {
	if method == "methods" {
		return func(env *Environment, args ...Object) Object {
			static := []string{"methods", "methods"}
			dynamic := env.Names("module.")

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
func (m *Module) ToInterface() interface{} {
	return "<MODULE>"
}
