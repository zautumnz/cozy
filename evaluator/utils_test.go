package evaluator

import (
	"testing"

	"github.com/zacanger/cozy/object"
)

func TestIsNumber(t *testing.T) {
	tests := []struct {
		number   string
		expected bool
	}{
		{"12", true},
		{"12a", false},
		{"12.2", true},
	}

	for _, tt := range tests {
		if tt.expected != IsNumber(tt.number) {
			t.Fatalf("expected %v (%s)", tt.expected, tt.number)
		}
	}
}

func TestInterpolate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"string", "string"},
		{"{string}", "{string}"},
		{`\{{string}}`, "{{string}}"},
		{`xy\z`, `xy\z`},
		{"{{string}}", "test"},
		{"{{string}}_", "test_"},
		{"{{string x", "{{string x"},
		{"{{string} x", "{{string} x"},
	}

	env := object.NewEnvironment()
	env.SetLet("string", &object.String{Value: "test"})

	for _, tt := range tests {
		output := Interpolate(tt.input, env)
		if tt.expected != output {
			t.Fatalf(
				"expected '%v', got '%v' (original: %s)",
				tt.expected,
				output,
				tt.input,
			)
		}
	}
}

func TestNewHash(t *testing.T) {
	res := NewHash(StringObjectMap{
		"foo": &object.Integer{Value: 1},
	})

	f := &object.String{Value: "foo"}
	p := res.Pairs[f.HashKey()]
	v := p.Value
	if v.Type() != object.INTEGER_OBJ {
		t.Fatalf("NewHash failed, expected integer, got %s", v.Type())
	}
	switch x := v.(type) {
	case *object.Integer:
		if x.Value != 1 {
			t.Fatalf("NewHash failed, expected value of 1, got %d", x.Value)
		}
	}
}
