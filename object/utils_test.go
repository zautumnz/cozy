package object

import "testing"

func TestIndentJSON(t *testing.T) {
	inp := `{"foo": [1,2,3], "bar": { "baz": 1 }}`
	exp := `{
    "foo": [
        1,
        2,
        3
    ],
    "bar": {
        "baz": 1
    }
}`
	res := indentJSON(inp)
	if exp != res {
		t.Fatalf("indentJSON failed: wanted %s, got %s", exp, res)
	}
}

func TestEscapeQuotes(t *testing.T) {
	inp := `"foo"`
	exp := `\"foo\"`
	res := escapeQuotes(inp)
	if exp != res {
		t.Fatalf("escapeQuotes failed: wanted %s, got %s", exp, res)
	}
}
