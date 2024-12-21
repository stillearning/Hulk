package parser

import (
	"Hulk/ast"
	"Hulk/lexer"
	"Hulk/token"
	"fmt"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      //==
	LESSGREATER //<>
	SUM         //+
	PRODUCT     //*
	PREFIX      //!5
	CALL        //function(x)
	INDEX       //index in arrays
)

var precedences = map[token.TokenType]int{
	token.EQUALS:    EQUALS,
	token.NOTEQUALS: EQUALS,
	token.LT:        LESSGREATER,
	token.GT:        LESSGREATER,
	token.ASTERISK:  PRODUCT,
	token.SLASH:     PRODUCT,
	token.PLUS:      SUM,
	token.MINUS:     SUM,
	token.LPAREN:    CALL,
	token.LBRACKET:  INDEX,
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) currPrecedence() int {
	if prec, ok := precedences[p.currToken.Type]; ok {
		return prec
	}
	return LOWEST
}

type (
	PrefixParsefn func() ast.Expression
	InfixParsefn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer
	//just like position and peekPosition in lexer, but instead of pointing to characters they point to next token
	currToken token.Token
	peekToken token.Token

	errors []string

	infixParsefns  map[token.TokenType]InfixParsefn
	prefixParsefns map[token.TokenType]PrefixParsefn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	p.prefixParsefns = make(map[token.TokenType]PrefixParsefn)
	p.RegisterPrefix(token.IDENTIFIER, p.parseIdentifier)
	p.RegisterPrefix(token.INT, p.parseIntegerLiteral)

	p.RegisterPrefix(token.BANG, p.parsePrefixExpression)
	p.RegisterPrefix(token.MINUS, p.parsePrefixExpression)

	p.RegisterPrefix(token.TRUE, p.parseBoolean)
	p.RegisterPrefix(token.FALSE, p.parseBoolean)

	p.RegisterPrefix(token.LPAREN, p.parseGroupedExpession)

	p.RegisterPrefix(token.IF, p.parseIfExpression)

	p.RegisterPrefix(token.FUNCTION, p.parseFunctionExpression)

	p.RegisterPrefix(token.STRING, p.parseStringExpression)

	p.RegisterPrefix(token.LBRACKET, p.parseArrayExpression)

	p.RegisterPrefix(token.LBRACE, p.parseHashExpression)

	p.infixParsefns = make(map[token.TokenType]InfixParsefn)
	p.RegisterInfix(token.PLUS, p.parseInfixExpression)
	p.RegisterInfix(token.MINUS, p.parseInfixExpression)
	p.RegisterInfix(token.ASTERISK, p.parseInfixExpression)
	p.RegisterInfix(token.SLASH, p.parseInfixExpression)
	p.RegisterInfix(token.EQUALS, p.parseInfixExpression)
	p.RegisterInfix(token.NOTEQUALS, p.parseInfixExpression)
	p.RegisterInfix(token.LT, p.parseInfixExpression)
	p.RegisterInfix(token.GT, p.parseInfixExpression)
	p.RegisterInfix(token.LPAREN, p.parseCallExpression)

	p.RegisterInfix(token.LBRACKET, p.parseIndexExpression)
	//read 2 tokens so that curr and peek tokens both are set
	p.NextToken()
	p.NextToken()

	return p
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
}

func (p *Parser) NextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for p.currToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.NextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.currToken}
	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.NextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.NextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.currToken}

	p.NextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.NextToken()
	}
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.currToken}

	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.NextToken()
	}
	return stmt
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) currTokenIs(t token.TokenType) bool {
	return p.currToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekToken.Type == t {
		p.NextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) RegisterInfix(token token.TokenType, fn InfixParsefn) {
	p.infixParsefns[token] = fn
}

func (p *Parser) RegisterPrefix(token token.TokenType, fn PrefixParsefn) {
	p.prefixParsefns[token] = fn
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParsefns[p.currToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParsefns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.NextToken()

		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.currToken}
	value, err := strconv.ParseInt(p.currToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
	}
	p.NextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(exp ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
		LeftExpr: exp,
	}
	precedence := p.currPrecedence()
	p.NextToken()
	expression.RightExpr = p.parseExpression(precedence)
	return expression
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix function found for %s", string(t))
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.BooleanExpression{Token: p.currToken, Value: p.currTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpession() ast.Expression {
	p.NextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{
		Token: p.currToken,
	}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.NextToken()
	expression.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.NextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}
	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	blockstmt := &ast.BlockStatement{Token: p.currToken}
	blockstmt.Statements = []ast.Statement{}
	p.NextToken()

	for !p.currTokenIs(token.RBRACE) && !p.currTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			blockstmt.Statements = append(blockstmt.Statements, stmt)
		}
		p.NextToken()
	}
	return blockstmt
}

func (p *Parser) parseFunctionExpression() ast.Expression {
	fnExp := &ast.FunctionLiteral{Token: p.currToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	fnExp.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fnExp.Block = p.parseBlockStatement()

	return fnExp
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	params := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.NextToken()
		return params
	}

	p.NextToken()
	ident := &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
	params = append(params, ident)

	for p.peekTokenIs(token.COMMA) {
		p.NextToken()
		p.NextToken()
		ident := &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
		params = append(params, ident)
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return params
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	expression := &ast.CallExpression{Token: p.currToken, Function: function}
	expression.Arguments = p.parseExpressionList(token.RPAREN)
	return expression
}

func (p *Parser) parseStringExpression() ast.Expression {
	return &ast.StringLiteral{Token: p.currToken, Value: p.currToken.Literal}
}

func (p *Parser) parseArrayExpression() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.currToken}

	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.NextToken()
		return list
	}

	p.NextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.NextToken()
		p.NextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.currToken, Left: left}

	p.NextToken()

	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseHashExpression() ast.Expression {
	hashExp := &ast.HashLiteral{Token: p.currToken}
	hashExp.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.NextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.NextToken()
		value := p.parseExpression(LOWEST)

		hashExp.Pairs[key] = value
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}

	}
	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return hashExp
}
