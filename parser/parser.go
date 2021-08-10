// Package parser is used to parse input-programs written in cozy
// and convert them to an abstract-syntax tree.
package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/zacanger/cozy/ast"
	"github.com/zacanger/cozy/lexer"
	"github.com/zacanger/cozy/token"
	"github.com/zacanger/cozy/utils"
)

// prefix Parse function
// infix parse function
// postfix parse function
type (
	prefixParseFn  func() ast.Expression
	infixParseFn   func(ast.Expression) ast.Expression
	postfixParseFn func() ast.Expression
)

// precedence order
const (
	_ int = iota
	LOWEST
	COND        // OR or AND
	ASSIGN      // =
	EQUALS      // == or !=
	LESSGREATER // > or <
	SUM         // + or -
	PRODUCT     // * or /
	POWER       // **
	MOD         // %
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	RANGE       // ..
	INDEX       // array[index], map[key], map.key
	HIGHEST
)

// each token precedence
var precedences = map[token.Type]int{
	token.ASSIGN:    ASSIGN,
	token.RANGE:     RANGE,
	token.EQ:        EQUALS,
	token.NOT_EQ:    EQUALS,
	token.LT:        LESSGREATER,
	token.LT_EQUALS: LESSGREATER,
	token.GT:        LESSGREATER,
	token.GT_EQUALS: LESSGREATER,

	token.PLUS:            SUM,
	token.PLUS_EQUALS:     SUM,
	token.MINUS:           SUM,
	token.MINUS_EQUALS:    SUM,
	token.SLASH:           PRODUCT,
	token.SLASH_EQUALS:    PRODUCT,
	token.ASTERISK:        PRODUCT,
	token.ASTERISK_EQUALS: PRODUCT,
	token.POW:             POWER,
	token.MOD:             MOD,
	token.AND:             COND,
	token.OR:              COND,
	token.LPAREN:          CALL,
	token.PERIOD:          CALL,
	token.LBRACKET:        INDEX,
}

// Parser object
type Parser struct {
	// l is our lexer
	l *lexer.Lexer

	// prevToken holds the previous token from our lexer.
	// (used for "++" + "--")
	prevToken token.Token

	// curToken holds the current token from our lexer.
	curToken token.Token

	// peekToken holds the next token which will come from the lexer.
	peekToken token.Token

	// errors holds parsing-errors.
	errors []string

	// prefixParseFns holds a map of parsing methods for
	// prefix-based syntax.
	prefixParseFns map[token.Type]prefixParseFn

	// infixParseFns holds a map of parsing methods for
	// infix-based syntax.
	infixParseFns map[token.Type]infixParseFn

	// postfixParseFns holds a map of parsing methods for
	// postfix-based syntax.
	postfixParseFns map[token.Type]postfixParseFn
}

// New returns our new parser-object.
func New(l *lexer.Lexer) *Parser {
	// Create the parser, and prime the pump
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()

	// Register prefix-functions
	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.EOF, p.parsingBroken)
	p.registerPrefix(token.FALSE, p.ParseBoolean)
	p.registerPrefix(token.FLOAT, p.ParseFloatLiteral)
	p.registerPrefix(token.FOR, p.parseForLoopExpression)
	p.registerPrefix(token.FOREACH, p.parseForEach)
	p.registerPrefix(token.IMPORT, p.parseImportExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.ILLEGAL, p.parsingBroken)
	p.registerPrefix(token.INT, p.ParseIntegerLiteral)
	p.registerPrefix(token.LBRACE, p.ParseHashLiteral)
	p.registerPrefix(token.LBRACKET, p.ParseArrayLiteral)
	p.registerPrefix(token.CURRENT_ARGS, p.parseCurrentArgsLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.NULL, p.parseNull)
	p.registerPrefix(token.STRING, p.ParseStringLiteral)
	p.registerPrefix(token.DOCSTRING, p.parseDocStringLiteral)
	p.registerPrefix(token.TRUE, p.ParseBoolean)

	// Register infix functions
	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK_EQUALS, p.parseAssignExpression)
	p.registerInfix(token.RANGE, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GT_EQUALS, p.parseInfixExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LT_EQUALS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS_EQUALS, p.parseAssignExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.PERIOD, p.parseIndexDotExpression)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.PLUS_EQUALS, p.parseAssignExpression)
	p.registerInfix(token.POW, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.SLASH_EQUALS, p.parseAssignExpression)

	// Register postfix functions.
	p.postfixParseFns = make(map[token.Type]postfixParseFn)
	p.registerPostfix(token.MINUS_MINUS, p.parsePostfixExpression)
	p.registerPostfix(token.PLUS_PLUS, p.parsePostfixExpression)

	// All done
	return p
}

// registerPrefix registers a function for handling a prefix-based statement
func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix registers a function for handling a infix-based statement
func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// registerPostfix registers a function for handling a postfix-based statement
func (p *Parser) registerPostfix(tokenType token.Type, fn postfixParseFn) {
	p.postfixParseFns[tokenType] = fn
}

// Errors return stored errors
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError raises an error if the next token is not the expected type.
func (p *Parser) peekError(t token.Type) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead around line %d", t, p.curToken.Type, p.l.GetLine())
	p.errors = append(p.errors, msg)
}

// nextToken moves to our next token from the lexer.
func (p *Parser) nextToken() {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseProgram used to parse the whole program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)
		p.nextToken()
	}
	return program
}

// parseStatement parses a single statement.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.MUTABLE:
		return p.parseMutableStatement()
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseMutableStatement parses a mutable-statement.
func (p *Parser) parseMutableStatement() *ast.MutableStatement {
	stmt := &ast.MutableStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseLetStatement parses a let (constant) declaration.
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement parses a return-statement.
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// no prefix parse function error
func (p *Parser) noPrefixParseFnError(t token.Type) {
	msg := fmt.Sprintf("no prefix parse function for %s found around line %d", t, p.l.GetLine())
	p.errors = append(p.errors, msg)
}

// parse Expression Statement
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	postfix := p.postfixParseFns[p.curToken.Type]
	if postfix != nil {
		return postfix()
	}
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

// parsingBroken is hit if we see an EOF in our input-stream
// this means we're screwed
func (p *Parser) parsingBroken() ast.Expression {
	return nil
}

// parseIdentifier parses an identifier.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// ParseIntegerLiteral parses an integer literal.
func (p *Parser) ParseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	var value int64
	var err error

	if strings.HasPrefix(p.curToken.Literal, "0b") {
		value, err = strconv.ParseInt(p.curToken.Literal[2:], 2, 64)
	} else if strings.HasPrefix(p.curToken.Literal, "0x") {
		value, err = strconv.ParseInt(p.curToken.Literal[2:], 16, 64)
	} else {
		value, err = strconv.ParseInt(p.curToken.Literal, 10, 64)
	}

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer around line %d", p.curToken.Literal, p.l.GetLine())
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

// ParseFloatLiteral parses a float-literal
func (p *Parser) ParseFloatLiteral() ast.Expression {
	flo := &ast.FloatLiteral{Token: p.curToken}
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float around line %d", p.curToken.Literal, p.l.GetLine())
		p.errors = append(p.errors, msg)
		return nil
	}
	flo.Value = value
	return flo
}

// ParseBoolean parses a boolean token.
func (p *Parser) ParseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// parseNull parses a null keyword
func (p *Parser) parseNull() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

// parsePrefixExpression parses a prefix-based expression.
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

// parsePostfixExpression parses a postfix-based expression.
func (p *Parser) parsePostfixExpression() ast.Expression {
	expression := &ast.PostfixExpression{
		Token:    p.prevToken,
		Operator: p.curToken.Literal,
	}
	return expression
}

// parseInfixExpression parses an infix-based expression.
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

// parseGroupedExpression parses a grouped-expression.
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// parseIfCondition parses an if-expression.
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	if expression == nil {
		p.errors = append(p.errors, "unexpected nil expression")
		return nil
	}

	// Look for the condition, surrounded by "(" + ")".
	expression.Condition = p.parseBracketExpression()
	if expression.Condition == nil {
		return nil
	}

	// Now "{"
	if !p.expectPeek(token.LBRACE) {
		msg := fmt.Sprintf("expected '{' but got %s", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	// The consequence
	expression.Consequence = p.parseBlockStatement()
	if expression.Consequence == nil {
		p.errors = append(p.errors, "unexpected nil expression")
		return nil
	}

	// Else?
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		// else if
		if p.peekTokenIs(token.IF) {

			p.nextToken()

			expression.Alternative = &ast.BlockStatement{
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: p.parseIfExpression(),
					},
				},
			}

			return expression
		}

		// else { block }
		if !p.expectPeek(token.LBRACE) {
			msg := fmt.Sprintf("expected '{' but got %s", p.curToken.Literal)
			p.errors = append(p.errors, msg)
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
		if expression.Alternative == nil {
			p.errors = append(p.errors, "unexpected nil expression")
			return nil
		}
	}
	return expression
}

// parseBracketExpression looks for an expression surrounded by "(" + ")".
// Used by parseIfExpression.
func (p *Parser) parseBracketExpression() ast.Expression {
	usingParens := false

	// check for (
	if p.peekTokenIs(token.LPAREN) {
		usingParens = true
		p.nextToken()
	}

	p.nextToken()

	// Look for the expression itself
	tmp := p.parseExpression(LOWEST)
	if tmp == nil {
		return nil
	}

	// if we started with parens...
	if usingParens {
		if !p.expectPeek(token.RPAREN) {
			msg := fmt.Sprintf("expected ')' but got %s", p.curToken.Literal)
			p.errors = append(p.errors, msg)
			return nil
		}
	}

	// otherwise
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	return tmp
}

// parseForLoopExpression parses a for-loop.
func (p *Parser) parseForLoopExpression() ast.Expression {
	expression := &ast.ForLoopExpression{Token: p.curToken}
	usingParens := false

	// see if we're using parens
	if p.peekTokenIs(token.LPAREN) {
		usingParens = true
		p.nextToken()
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	// if we started with parens
	if usingParens {
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	// otherwise
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()
	return expression
}

// parseForEach parses 'foreach x X { .. block .. }`
func (p *Parser) parseForEach() ast.Expression {
	expression := &ast.ForeachStatement{Token: p.curToken}

	// get the id
	p.nextToken()
	expression.Ident = p.curToken.Literal

	// If we find a "," we then get a second identifier too.
	if p.peekTokenIs(token.COMMA) {

		// Generally we have:
		//    foreach IDENT in THING { .. }
		// If we have two arguments the first becomes
		// the index, and the second becomes the IDENT.

		// skip the comma
		p.nextToken()

		if !p.peekTokenIs(token.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("second argument to foreach must be ident, got %v", p.peekToken))
			return nil
		}
		p.nextToken()

		// Record the updated values.
		expression.Index = expression.Ident
		expression.Ident = p.curToken.Literal

	}

	// The next token, after the ident(s), should be `in`.
	if !p.expectPeek(token.IN) {
		return nil
	}
	p.nextToken()

	// get the thing we're going to iterate  over.
	expression.Value = p.parseExpression(LOWEST)
	if expression.Value == nil {
		return nil
	}

	// parse the block
	p.nextToken()
	expression.Body = p.parseBlockStatement()

	return expression
}

// Parse import statements for modules
func (p *Parser) parseImportExpression() ast.Expression {
	expression := &ast.ImportExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Name = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return expression
}

// parseBlockStatement parses a block.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) {

		// Don't loop forever
		if p.curTokenIs(token.EOF) {
			p.errors = append(p.errors,
				"unterminated block statement")
			return nil
		}

		stmt := p.parseStatement()
		block.Statements = append(block.Statements, stmt)
		p.nextToken()
	}
	return block
}

// parseFunctionLiteral parses a function-literal.
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Defaults, lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	if p.peekTokenIs(token.DOCSTRING) {
		x := p.parseDocStringLiteral()
		switch a := x.(type) {
		case *ast.DocStringLiteral:
			lit.DocString = a
		}
	}
	lit.Body = p.parseBlockStatement()
	return lit
}

// ...
func (p *Parser) parseCurrentArgsLiteral() ast.Expression {
	return &ast.CurrentArgsLiteral{Token: p.curToken}
}

// parseFunctionParameters parses the parameters used for a function.
func (p *Parser) parseFunctionParameters() (map[string]ast.Expression, []*ast.Identifier) {

	// Any default parameters.
	m := make(map[string]ast.Expression)

	// The argument-definitions.
	identifiers := make([]*ast.Identifier, 0)

	// Is the next parameter ")" ?  If so we're done. No args.
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return m, identifiers
	}
	p.nextToken()

	// Keep going until we find a ")"
	for !p.curTokenIs(token.RPAREN) {

		if p.curTokenIs(token.EOF) {
			p.errors = append(p.errors, "unterminated function parameters")
			return nil, nil
		}

		// Get the identifier.
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
		p.nextToken()

		// If there is "=xx" after the name then that's
		// the default parameter.
		if p.curTokenIs(token.ASSIGN) {
			p.nextToken()
			// Save the default value.
			m[ident.Value] = p.parseExpressionStatement().Expression
			p.nextToken()
		}

		// Skip any comma.
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	return m, identifiers
}

// ParseStringLiteral parses a string-literal.
func (p *Parser) ParseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseDocStringLiteral parses a docstring-literal.
func (p *Parser) parseDocStringLiteral() ast.Expression {
	p.nextToken()
	x := &ast.DocStringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return x
}

// ParseArrayLiteral parses an array literal.
func (p *Parser) ParseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

// parsearray elements literal
func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	list := make([]ast.Expression, 0)
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}

// parseInfixExpression parses an array index expression.
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

// parseAssignExpression parses a bare assignment, without a `mutable`.
func (p *Parser) parseAssignExpression(name ast.Expression) ast.Expression {
	stmt := &ast.AssignStatement{Token: p.curToken}
	if n, ok := name.(*ast.Identifier); ok {
		stmt.Name = n
	} else {
		msg := fmt.Sprintf("expected assign token to be IDENT, got %s instead around line %d", name.TokenLiteral(), p.l.GetLine())
		p.errors = append(p.errors, msg)
	}

	oper := p.curToken
	p.nextToken()

	// An assignment is generally:
	//    variable = value
	// But we cheat and reuse the implementation for:
	//    i += 4
	// In this case we record the "operator" as "+="
	switch oper.Type {
	case token.PLUS_EQUALS:
		stmt.Operator = "+="
	case token.MINUS_EQUALS:
		stmt.Operator = "-="
	case token.SLASH_EQUALS:
		stmt.Operator = "/="
	case token.ASTERISK_EQUALS:
		stmt.Operator = "*="
	default:
		stmt.Operator = "="
	}
	stmt.Value = p.parseExpression(LOWEST)
	return stmt
}

// parseCallExpression parses a function-call expression.
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// ParseHashLiteral parses a hash literal.
func (p *Parser) ParseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}
	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return hash
}

// parseIndexDotExpression parses an index with DOT separator.
func (p *Parser) parseIndexDotExpression(obj ast.Expression) ast.Expression {
	curToken := p.curToken
	p.nextToken()
	name := p.parseIdentifier()
	return &ast.IndexExpression{
		Token: curToken,
		Left:  obj,
		Index: &ast.StringLiteral{
			Token: token.Token{
				Type:    token.IDENT,
				Literal: name.TokenLiteral(),
			},
			Value: name.String(),
		},
	}
}

// curTokenIs tests if the current token has the given type.
func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

// peekTokenIs tests if the next token has the given type.
func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

// expectPeek validates the next token is of the given type,
// and advances if so. If it is not an error is stored.
func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}

	p.peekError(t)
	return false
}

// peekPrecedence looks up the next token precedence.
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence looks up the current token precedence.
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// ParserErrorsParams is used in main and repl
type ParserErrorsParams struct {
	Errors []string
	Out    io.Writer
}

// PrintParserErrors prints parser errors
func PrintParserErrors(arg ParserErrorsParams) {
	if arg.Out != nil {
		io.WriteString(arg.Out, "ERROR!\n")
		io.WriteString(arg.Out, " parser errors:\n")
		for _, msg := range arg.Errors {
			io.WriteString(arg.Out, "\t"+msg+"\n")
		}
		// we're in a repl, so don't do anything else
	} else {
		for _, msg := range arg.Errors {
			fmt.Printf("\t%s\n", msg)
		}
		// we're not in a repl, so exit
		utils.ExitConditionally(1)
	}
}
