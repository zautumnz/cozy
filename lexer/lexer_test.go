package lexer

import (
	"testing"

	"github.com/zautumnz/cozy/token"
)

func TestNull(t *testing.T) {
	input := "a = null;"
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.NULL, "null"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf(
				"tests[%d] - tokentype wrong, expected=%q, got=%q",
				i,
				tt.expectedType,
				tok.Type,
			)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - Literal wrong, expected=%q, got=%q",
				i,
				tt.expectedLiteral,
				tok.Literal,
			)
		}
	}
}

func TestNextToken1(t *testing.T) {
	input := "%=+(){},;?|| &&++--***=..>>"

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.MOD, "%"},
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.QUESTION, "?"},
		{token.OR, "||"},
		{token.AND, "&&"},
		{token.PLUS_PLUS, "++"},
		{token.MINUS_MINUS, "--"},
		{token.POW, "**"},
		{token.ASTERISK_EQUALS, "*="},
		{token.RANGE, ".."},
		{token.BIT_RIGHT_SHIFT, ">>"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf(
				"tests[%d] - tokentype wrong, expected=%q, got=%q",
				i,
				tt.expectedType,
				tok.Type,
			)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - Literal wrong, expected=%q, got=%q",
				i,
				tt.expectedLiteral,
				tok.Literal,
			)
		}
	}
}

func TestNextToken2(t *testing.T) {
	input := `a.b;
{}.a;
a[b].c;
let five = 5;
let ten =10;
let add = fn (x, y){
    x+y;
};
let result = add(five, ten);
!- *5;
5<10>5;

if(5<10){
	return true;
}else{
	return false;
}
10 == 10;
10 != 9;
"foobar"
"foo bar"
[1,2];
{"foo":"bar"}
1.2
0.5
0.3
世界
for
2 >= 1
1 <= 3
...
foo
`
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.PERIOD, "."},
		{token.IDENT, "b"},
		{token.SEMICOLON, ";"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.PERIOD, "."},
		{token.IDENT, "a"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "a"},
		{token.LBRACKET, "["},
		{token.IDENT, "b"},
		{token.RBRACKET, "]"},
		{token.PERIOD, "."},
		{token.IDENT, "c"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.FLOAT, "1.2"},
		{token.FLOAT, "0.5"},
		{token.FLOAT, "0.3"},
		{token.IDENT, "世界"},
		{token.FOR, "for"},
		{token.INT, "2"},
		{token.GT_EQUALS, ">="},
		{token.INT, "1"},
		{token.INT, "1"},
		{token.LT_EQUALS, "<="},
		{token.INT, "3"},
		{token.CURRENT_ARGS, "..."},
		{token.IDENT, "foo"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf(
				"tests[%d] - tokentype wrong, expected=%q, got=%q",
				i,
				tt.expectedType,
				tok.Type,
			)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - Literal wrong, expected=%q, got=%q",
				i,
				tt.expectedLiteral,
				tok.Literal,
			)
		}
	}
}

func TestUnicodeLexer(t *testing.T) {
	input := `世界`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.IDENT {
		t.Fatalf(
			"token type wrong, expected=%q, got=%q",
			token.IDENT,
			tok.Type,
		)
	}
	if tok.Literal != "世界" {
		t.Fatalf(
			"token literal wrong, expected=%q, got=%q",
			"世界",
			tok.Literal,
		)
	}
}

func TestSimpleComment(t *testing.T) {
	input := `=+# This is a comment
# This is a comment
let a = 1; # This is a comment
# This is a final
# comment on two-lines`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LET, "let"},
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf(
				"tests[%d] - tokentype wrong, expected=%q, got=%q",
				i,
				tt.expectedType,
				tok.Type,
			)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - Literal wrong, expected=%q, got=%q",
				i,
				tt.expectedLiteral,
				tok.Literal,
			)
		}
	}
}

func TestIntegers(t *testing.T) {
	input := `10 0x10 0xF0 0xFE 0b0101 0xFF 0b101 0xFF;`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.INT, "10"},
		{token.INT, "0x10"},
		{token.INT, "0xF0"},
		{token.INT, "0xFE"},
		{token.INT, "0b0101"},
		{token.INT, "0xFF"},
		{token.INT, "0b101"},
		{token.INT, "0xFF"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf(
				"tests[%d] - tokentype wrong, expected=%q, got=%q",
				i,
				tt.expectedType,
				tok.Type,
			)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - Literal wrong, expected=%q, got=%q",
				i,
				tt.expectedLiteral,
				tok.Literal,
			)
		}
	}
}

// Test that the shebang-line is handled specially.
func TestShebang(t *testing.T) {
	input := `#!/bin/cozy
10;`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf(
				"tests[%d] - tokentype wrong, expected=%q, got=%q",
				i,
				tt.expectedType,
				tok.Type,
			)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - Literal wrong, expected=%q, got=%q",
				i,
				tt.expectedLiteral,
				tok.Literal,
			)
		}
	}
}

// TestMoreHandling does nothing real, but it bumps our coverage!
func TestMoreHandling(t *testing.T) {
	input := `#!/bin/cozy
1 += 1;
2 -= 2;
3 /= 3;
x */ 3;

let t = true;
let f = false;

if (t && f) { print("What?"); }
if (t || f) { print("What?"); }

let a = 1;
a++;

let b = a % 1;
b--;
b -= 2;

if (a<3) { print("Blah!"); }
if (a>3) { print("Blah!"); }

let b = 3;
b**b;
b *= 3;
if (b <= 3) { print "blah\n" }
if (b >= 3) { print "blah\n" }

let a = "autumn";
let a = "autumn\n";
let a = "autumn\t";
let a = "autumn\r";
let a = "autumn\\";
let a = "autumn\"";
let c = 3.113;
.;`

	l := New(input)
	tok := l.NextToken()
	for tok.Type != token.EOF {
		tok = l.NextToken()
	}
}

// TestStdLib ensures that identifiers are parsed correctly
func TestStdLib(t *testing.T) {
	input := `
sys.getenv
sys.setenv
sys.environment
fs.glob
math.abs
math.random
math.sqrt
string.interpolate
string.toupper
string.tolower
string.trim
string.reverse
string.split
foo.bar
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "sys.getenv"},
		{token.IDENT, "sys.setenv"},
		{token.IDENT, "sys.environment"},
		{token.IDENT, "fs.glob"},
		{token.IDENT, "math.abs"},
		{token.IDENT, "math.random"},
		{token.IDENT, "math.sqrt"},
		{token.IDENT, "string.interpolate"},
		{token.IDENT, "string.toupper"},
		{token.IDENT, "string.tolower"},
		{token.IDENT, "string.trim"},
		{token.IDENT, "string.reverse"},
		{token.IDENT, "string.split"},
		{token.IDENT, "foo"},
		{token.PERIOD, "."},
		{token.IDENT, "bar"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf(
				"tests[%d] - tokentype wrong, expected=%q, got=%q",
				i,
				tt.expectedType,
				tok.Type,
			)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - Literal wrong, expected=%q, got=%q",
				i,
				tt,
				tok,
			)
		}
	}
}

// TestDotMethod ensures that identifiers are parsed correctly for the
// case where we need to split at periods.
func TestDotMethod(t *testing.T) {
	input := `
foo.bar();
a?.b?();
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "foo"},
		{token.PERIOD, "."},
		{token.IDENT, "bar"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "a?"},
		{token.PERIOD, "."},
		{token.IDENT, "b?"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf(
				"tests[%d] - tokentype wrong, expected=%q, got=%q",
				i,
				tt.expectedType,
				tok.Type,
			)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - Literal wrong, expected=%q, got=%q",
				i,
				tt,
				tok,
			)
		}
	}
}

// TestIntDotMethod ensures that identifiers are parsed correctly for the
// case where they immediately follow int/float valies.
func TestIntDotMethod(t *testing.T) {
	input := `
3.foo();
3.14.bar();
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.INT, "3"},
		{token.PERIOD, "."},
		{token.IDENT, "foo"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.FLOAT, "3.14"},
		{token.PERIOD, "."},
		{token.IDENT, "bar"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf(
				"tests[%d] - tokentype wrong, expected=%q, got=%q",
				i,
				tt.expectedType,
				tok.Type,
			)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - Literal wrong, expected=%q, got=%q",
				i,
				tt,
				tok,
			)
		}
	}
}

// TestDiv is designed to test that a division is recognized
func TestDiv(t *testing.T) {
	input := `a = b / c;
a = 3/4;
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.IDENT, "b"},
		{token.SLASH, "/"},
		{token.IDENT, "c"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.INT, "3"},
		{token.SLASH, "/"},
		{token.INT, "4"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf(
				"tests[%d] - tokentype wrong, expected=%q, got=%q",
				i,
				tt.expectedType,
				tok.Type,
			)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - Literal wrong, expected=%q, got=%q",
				i,
				tt.expectedLiteral,
				tok.Literal,
			)
		}
	}
}

// TestRangeDots is designed to ensure we get a ".." not an integer value.
func TestRangeDots(t *testing.T) {
	input := `a = 1..10;`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.INT, "1"},
		{token.RANGE, ".."},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf(
				"tests[%d] - tokentype wrong, expected=%q, got=%q",
				i,
				tt.expectedType,
				tok.Type,
			)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - Literal wrong, expected=%q, got=%q",
				i,
				tt.expectedLiteral,
				tok.Literal,
			)
		}
	}
}
