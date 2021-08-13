// Package token contains constants which are used when lexing a program
// written in the cozy language, as done by the parser.
package token

// Type is a string
type Type string

// Token struct represent the lexer token
type Token struct {
	Type    Type
	Literal string
}

// pre-defined Type
const (
	AND             = "&&"
	ASSIGN          = "="
	ASTERISK        = "*"
	ASTERISK_EQUALS = "*="
	BANG            = "!"
	COLON           = ":"
	COMMA           = ","
	CURRENT_ARGS    = "..."
	DOCSTRING       = "DOCSTRING"
	ELSE            = "ELSE"
	EOF             = "EOF"
	EQ              = "=="
	FALSE           = "FALSE"
	FLOAT           = "FLOAT"
	FOR             = "FOR"
	FOREACH         = "FOREACH"
	FUNCTION        = "FUNCTION"
	GT              = ">"
	GT_EQUALS       = ">="
	IDENT           = "IDENT"
	IF              = "IF"
	ILLEGAL         = "ILLEGAL"
	IMPORT          = "IMPORT"
	IN              = "IN"
	INT             = "INT"
	LBRACE          = "{"
	LBRACKET        = "["
	LET             = "LET"
	LPAREN          = "("
	LT              = "<"
	LT_EQUALS       = "<="
	MINUS           = "-"
	MINUS_EQUALS    = "-="
	MINUS_MINUS     = "--"
	MOD             = "%"
	MUTABLE         = "MUTABLE"
	NOT_EQ          = "!="
	NULL            = "null"
	OR              = "||"
	PERIOD          = "."
	PLUS            = "+"
	PLUS_EQUALS     = "+="
	PLUS_PLUS       = "++"
	POW             = "**"
	QUESTION        = "?"
	RANGE           = ".."
	RBRACE          = "}"
	RBRACKET        = "]"
	RETURN          = "RETURN"
	RPAREN          = ")"
	SEMICOLON       = ";"
	SLASH           = "/"
	SLASH_EQUALS    = "/="
	SPREAD          = "...."
	STRING          = "STRING"
	TRUE            = "TRUE"
)

// reversed keywords
var keywords = map[string]Type{
	"else":    ELSE,
	"false":   FALSE,
	"fn":      FUNCTION,
	"for":     FOR,
	"foreach": FOREACH,
	"if":      IF,
	"import":  IMPORT,
	"in":      IN,
	"let":     LET,
	"mutable": MUTABLE,
	"null":    NULL,
	"return":  RETURN,
	"true":    TRUE,
}

// LookupIdentifier used to determinate whether identifier is keyword nor not
func LookupIdentifier(identifier string) Type {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENT
}
