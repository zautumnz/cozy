package evaluator

import (
	"strings"

	"github.com/zacanger/cozy/ast"
	"github.com/zacanger/cozy/lexer"
	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/parser"
)

// Converts a valid JSON string to a cozy value
func jsonDeserialize(args ...object.Object) object.Object {
	s := args[0].(*object.String)
	str := strings.TrimSpace(s.Value)
	env := object.NewEnvironment()
	l := lexer.New(str)
	p := parser.New(l)
	var node ast.Node
	ok := false

	if len(str) != 0 {
		switch str[0] {
		case '{':
			node, ok = p.ParseHashLiteral().(*ast.HashLiteral)
		case '[':
			node, ok = p.ParseArrayLiteral().(*ast.ArrayLiteral)
		}
	}

	// if str is empty, the length will be 0
	// we can parse it the same way as string literal
	if len(str) == 0 || (str[0] == '"' && str[len(str)-1] == '"') {
		node, ok = p.ParseStringLiteral().(*ast.StringLiteral)
	}

	if IsNumber(str) {
		if strings.Contains(str, ".") {
			node, ok = p.ParseFloatLiteral().(*ast.FloatLiteral)
		} else {
			node, ok = p.ParseIntegerLiteral().(*ast.IntegerLiteral)
		}
	}

	if str == "false" || str == "true" {
		node, ok = p.ParseBoolean().(*ast.Boolean)
	}

	if ok {
		return Eval(node, env)
	}

	return NewError("argument to `json` must be a valid JSON object, got '%s'", s.Value)
}

// Converts a cozy value to a JSON string
// Every cozy object (type) has a Json method, so this is easy
func jsonSerialize(args ...object.Object) object.Object {
	return &object.String{Value: args[0].Json()}
}

func init() {
	RegisterBuiltin("json.deserialize",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (jsonDeserialize(args...))
		})
	RegisterBuiltin("json.serialize",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (jsonSerialize(args...))
		})
}
