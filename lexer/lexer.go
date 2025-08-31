// Package lexer contains the code to lex input-programs into a stream
// of tokens, such that they may be parsed.
package lexer

import (
	"strings"
	"unicode"

	"github.com/zautumnz/cozy/token"
)

// Lexer holds our object-state.
type Lexer struct {
	// The current character position
	position int

	// The next character position
	readPosition int

	// The current character
	ch rune

	// A rune slice of our input string
	characters []rune

	// Previous token.
	prevToken token.Token
}

// New a Lexer instance from string input.
func New(inputs ...string) *Lexer {
	input := ""
	for _, inp := range inputs {
		input += inp
		input += "\n\n"
	}
	l := &Lexer{characters: []rune(input)}
	l.readChar()
	return l
}

// GetLine returns the rough line-number of our current position.
func (l *Lexer) GetLine() int {
	line := 0
	chars := len(l.characters)
	i := 0

	for i < l.readPosition && i < chars {
		if l.characters[i] == rune('\n') {
			line++
		}

		i++
	}

	return line
}

// read one forward character
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.characters) {
		l.ch = rune(0)
	} else {
		l.ch = l.characters[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken to read next token, skipping the white space.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()

	// skip comments
	if l.ch == rune('#') {
		l.skipComment()
		return l.NextToken()
	}

	switch l.ch {
	case rune('&'):
		if l.peekChar() == rune('&') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.AND,
				Literal: string(ch) + string(l.ch),
			}
		} else {
			tok = token.Token{
				Type:    token.BIT_AND,
				Literal: string(l.ch),
			}
		}
	case rune('|'):
		if l.peekChar() == rune('|') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.OR,
				Literal: string(ch) + string(l.ch),
			}
		} else {
			tok = token.Token{
				Type:    token.BIT_OR,
				Literal: string(l.ch),
			}
		}
	case rune('^'):
		tok = token.Token{
			Type:    token.BIT_XOR,
			Literal: string(l.ch),
		}
	case rune('~'):
		tok = token.Token{
			Type:    token.BIT_NOT,
			Literal: string(l.ch),
		}
	case rune('='):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.EQ,
				Literal: string(ch) + string(l.ch),
			}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case rune(';'):
		tok = newToken(token.SEMICOLON, l.ch)
	case rune('?'):
		tok = newToken(token.QUESTION, l.ch)
	case rune('('):
		tok = newToken(token.LPAREN, l.ch)
	case rune(')'):
		tok = newToken(token.RPAREN, l.ch)
	case rune(','):
		tok = newToken(token.COMMA, l.ch)
	case rune('.'):
		if l.peekChar() == rune('.') {
			// range, args, or spread
			l.readChar()
			if l.peekChar() == rune('.') {
				// args or spread
				l.readChar()
				if l.peekChar() == rune('.') {
					// spread
					tok = token.Token{Type: token.SPREAD, Literal: "...."}
					l.readChar()
				} else {
					// args
					tok = token.Token{Type: token.CURRENT_ARGS, Literal: "..."}
				}
			} else {
				// range
				tok = token.Token{Type: token.RANGE, Literal: ".."}
			}
		} else {
			// just a dot
			tok = newToken(token.PERIOD, l.ch)
		}
	case rune('+'):
		if l.peekChar() == rune('+') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.PLUS_PLUS,
				Literal: string(ch) + string(l.ch),
			}
		} else if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.PLUS_EQUALS,
				Literal: string(ch) + string(l.ch),
			}
		} else {
			tok = newToken(token.PLUS, l.ch)
		}
	case rune('%'):
		tok = newToken(token.MOD, l.ch)
	case rune('{'):
		tok = newToken(token.LBRACE, l.ch)
	case rune('}'):
		tok = newToken(token.RBRACE, l.ch)
	case rune('-'):
		if l.peekChar() == rune('-') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.MINUS_MINUS,
				Literal: string(ch) + string(l.ch),
			}
		} else if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.MINUS_EQUALS,
				Literal: string(ch) + string(l.ch),
			}
		} else {
			tok = newToken(token.MINUS, l.ch)
		}
	case rune('/'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.SLASH_EQUALS,
				Literal: string(ch) + string(l.ch),
			}
		} else {
			// slash is mostly division, but could
			// be the start of a regular expression

			// We exclude:
			//   a[b] / c       -> RBRACKET
			//   (a + b) / c    -> RPAREN
			//   a / c           -> IDENT
			//   3.2 / c         -> FLOAT
			//   1 / c           -> IDENT

			if l.prevToken.Type == token.RBRACKET ||
				l.prevToken.Type == token.RPAREN ||
				l.prevToken.Type == token.IDENT ||
				l.prevToken.Type == token.INT ||
				l.prevToken.Type == token.FLOAT {
				tok = newToken(token.SLASH, l.ch)
			}
		}
	case rune('*'):
		if l.peekChar() == rune('*') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.POW,
				Literal: string(ch) + string(l.ch),
			}
		} else if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.ASTERISK_EQUALS,
				Literal: string(ch) + string(l.ch),
			}
		} else {
			tok = newToken(token.ASTERISK, l.ch)
		}
	case rune('<'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.LT_EQUALS,
				Literal: string(ch) + string(l.ch),
			}
		} else if l.peekChar() == rune('<') {
			l.readChar()
			tok = token.Token{
				Type:    token.BIT_LEFT_SHIFT,
				Literal: "<<",
			}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case rune('>'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.GT_EQUALS,
				Literal: string(ch) + string(l.ch),
			}
		} else if l.peekChar() == rune('>') {
			l.readChar()
			tok = token.Token{
				Type:    token.BIT_RIGHT_SHIFT,
				Literal: ">>",
			}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case rune('!'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{
				Type:    token.NOT_EQ,
				Literal: string(ch) + string(l.ch),
			}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case rune('"'):
		tok.Type = token.STRING
		tok.Literal = l.readString(false)
	case rune('\''):
		tok.Type = token.DOCSTRING
		tok.Literal = l.readString(true)
	case rune('['):
		tok = newToken(token.LBRACKET, l.ch)
	case rune(']'):
		tok = newToken(token.RBRACKET, l.ch)
	case rune(':'):
		tok = newToken(token.COLON, l.ch)
	case rune(0):
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isDigit(l.ch) {
			tok = l.readDecimal()
			l.prevToken = tok
			return tok

		}
		tok.Literal = l.readIdentifier()
		tok.Type = token.LookupIdentifier(tok.Literal)
		l.prevToken = tok

		return tok
	}

	l.readChar()
	l.prevToken = tok
	return tok
}

// return new token
func newToken(tokenType token.Type, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// readIdentifier is designed to read an identifier (name of variable,
// function, etc).
//
// However there is a complication due to our historical implementation
// of the standard library. We really want to stop identifiers if we hit
// a period, to allow method-calls to work on objects.
//
// So with input like this:
//
//	a.blah();
//
// Our identifier should be "a" (then we have a period, then a second
// identifier "blah", followed by opening & closing parenthesis).
//
// However we also have to cover the case of:
//
//	string.toupper("blah");
//	os.getenv("PATH");
//	etc.
func (l *Lexer) readIdentifier() string {
	// Types and objects which will have valid methods.
	types := []string{
		"array.",
		"core.",
		"float.",
		"fs.",
		"hash.",
		"http.",
		"integer.",
		"json.",
		"math.",
		"net.",
		"object.",
		"string.",
		"sys.",
		"time.",
		"util.",
	}

	id := ""

	// Save our position, in case we need to jump backwards in
	// our scanning.
	position := l.position
	rposition := l.readPosition

	// Build up our identifier, handling only valid characters.
	// NOTE: This WILL consider the period valid, allowing the
	// parsing of "foo.bar", "os.getenv", "blah.blah.blah", etc.
	for isIdentifier(l.ch) {
		id += string(l.ch)
		l.readChar()
	}

	// Now we to see if our identifier had a period inside it.
	if strings.Contains(id, ".") {
		ok := false
		for _, i := range types {
			if strings.HasPrefix(id, i) {
				ok = true
			}
		}

		// Not permitted? Then we abort.
		// We reset our lexer-state to the position just ahead
		// of the period. This will then lead to a syntax
		// error.
		// Which probably means our lexer should abort instead.
		// For the moment we'll leave as-is.
		if !ok {
			// OK first of all we truncate our identifier
			// at the position before the "."
			offset := strings.Index(id, ".")
			id = id[:offset]

			// Now we have to move backwards - as a quickie
			// We'll reset our position and move forwards
			// the length of the bits we went too-far.
			l.position = position
			l.readPosition = rposition
			for offset > 0 {
				l.readChar()
				offset--
			}
		}
	}

	// And now our pain is over.
	return id
}

// skip white space
func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

// skip comment (until the end of the line).
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != rune(0) {
		l.readChar()
	}
	l.skipWhitespace()
}

// read number - this handles 0x1234 and 0b101010101 too.
func (l *Lexer) readNumber() string {
	str := ""

	// We usually just accept digits.
	accept := "0123456789"

	// But if we have `0x` as a prefix we accept hexadecimal instead.
	if l.ch == '0' && l.peekChar() == 'x' {
		accept = "0x123456789abcdefABCDEF"
	}

	// If we have `0b` as a prefix we accept binary digits only.
	if l.ch == '0' && l.peekChar() == 'b' {
		accept = "b01"
	}

	for strings.Contains(accept, string(l.ch)) {
		str += string(l.ch)
		l.readChar()
	}
	return str
}

// read decimal
func (l *Lexer) readDecimal() token.Token {
	// Read an integer-number.
	integer := l.readNumber()

	// Now we either expect:
	//   .[digits]  -> Which converts us from an int to a float.
	//   .blah      -> Which is a method-call on a raw number.
	if l.ch == rune('.') && isDigit(l.peekChar()) {
		// OK here we think we've got a float.
		l.readChar()
		fraction := l.readNumber()
		return token.Token{Type: token.FLOAT, Literal: integer + "." + fraction}
	}
	return token.Token{Type: token.INT, Literal: integer}
}

// read strings and docstrings
func (l *Lexer) readString(isDocString bool) string {
	out := ""
	delim := '"'
	if isDocString {
		delim = '\''
	}

	for {
		l.readChar()
		if l.ch == delim {
			break
		}

		// Handle \n, \r, \t, \", etc.
		if l.ch == '\\' {
			l.readChar()

			// escaped string delimiters
			if isDocString {
				if l.ch == rune('\'') {
					l.ch = '\''
				}
			} else {
				if l.ch == rune('"') {
					l.ch = '"'
				}
			}

			// other escapes
			if l.ch == rune('n') {
				l.ch = '\n'
			}
			if l.ch == rune('r') {
				l.ch = '\r'
			}
			if l.ch == rune('t') {
				l.ch = '\t'
			}
			if l.ch == rune('\\') {
				l.ch = '\\'
			}
		}

		out = out + string(l.ch)
	}

	return out
}

// peek character
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.characters) {
		return rune(0)
	}
	return l.characters[l.readPosition]
}

// determinate ch is identifier or not
func isIdentifier(ch rune) bool {
	return unicode.IsLetter(ch) ||
		unicode.IsDigit(ch) ||
		ch == '.' ||
		ch == '?' ||
		ch == '$' ||
		ch == '_'
}

// is white space
func isWhitespace(ch rune) bool {
	return ch == rune(' ') ||
		ch == rune('\t') ||
		ch == rune('\n') ||
		ch == rune('\r')
}

// is Digit
func isDigit(ch rune) bool {
	return rune('0') <= ch && ch <= rune('9')
}
