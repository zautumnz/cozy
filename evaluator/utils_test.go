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
		{"\\{{string}}", "{{string}}"},
		{"xy\\z", "xy\\z"},
		{"{{string}}", "test"},
		{"{{string}}_", "test_"},
		{"{{string x", "{{string x"},
		{"{{string} x", "{{string} x"},
	}

	env := object.NewEnvironment()
	env.Set("string", &object.String{Value: "test"})

	for _, tt := range tests {
		output := Interpolate(tt.input, env)
		if tt.expected != output {
			t.Fatalf("expected '%v', got '%v' (original: %s)", tt.expected, output, tt.input)
		}
	}
}
