package evaluator

import (
	"strings"

	"github.com/zautumnz/keai/ast"
	"github.com/zautumnz/keai/lexer"
	"github.com/zautumnz/keai/object"
	"github.com/zautumnz/keai/parser"
)

// Converts a valid JSON string to a keai value
func jsonDeserialize(args ...OBJ) OBJ {
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

	if str == "null" {
		return NULL
	}

	if ok {
		return Eval(node, env)
	}

	return NewError(
		"argument to `json` must be a valid JSON object, got '%s'",
		s.Value,
	)
}

// Converts a keai value to a JSON string
// Every keai object (type) has a JSON method, so this is easy
func jsonSerialize(args ...OBJ) OBJ {
	indent := false
	if len(args) > 1 {
		if isTruthy(args[1]) {
			indent = true
		}
	}

	return &object.String{Value: args[0].JSON(indent)}
}

func init() {
	RegisterBuiltin("json.deserialize",
		func(env *ENV, args ...OBJ) OBJ {
			return jsonDeserialize(args...)
		})
	RegisterBuiltin("json.serialize",
		func(env *ENV, args ...OBJ) OBJ {
			return jsonSerialize(args...)
		})
}
