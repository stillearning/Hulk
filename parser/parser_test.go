package parser

import (
	"Hulk/ast"
	"Hulk/lexer"
	"fmt"
	"testing"
)

func TestReturnStatements(t *testing.T) {
	input := `
	return 5;
	return 10;
	return 58582;`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 3 {
		t.Fatalf("program.statements does not have 3 statements but have %d", len(program.Statements))
	}
	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.returnStatement got =%T", stmt)
			continue
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral not 'return', got %q", returnStmt.TokenLiteral())
		}
	}
}

func TestLetStatements(t *testing.T) {
	input := `
	let x =5;
	let y =10;
	let foobaar=8383;
	`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram return nil!\n")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("Program does not contain 3 statements, got=%d", len(program.Statements))
	}
	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobaar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testLetStatement(t *testing.T, stmt ast.Statement, name string) bool {
	if stmt.TokenLiteral() != "let" {
		t.Errorf("Illegal start of statement, no let found\n")
	}
	letStmt, ok := stmt.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", stmt)
		return false
	}
	if letStmt.Name.Value != name {
		t.Errorf("expected letStmt.Name.Value=%s, got=%s", name, letStmt.Name.Value)
	}
	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("expected letStmt.Name.TokenLiteral()=%s, got=%s", name, letStmt.Name.Value)
	}
	return true
}

func TestIdentfierExpression(t *testing.T) {
	input := "foobar"

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Errorf("expected 1 statement but got %d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Errorf("%T statement is not ast.expressionStatement", stmt)
	}
	identstmt, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Errorf("%T statement is not ast.Identifier", identstmt)
	}
	if identstmt.Value != "foobar" {
		t.Errorf("expected foobar but got %s", identstmt.Value)
	}
	if identstmt.TokenLiteral() != "foobar" {
		t.Errorf("expected foobar but got %s", identstmt.Value)
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5"
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Errorf("expected 1 statement but got %d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Errorf("expressionStatement is not correct, got %T", stmt)
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)

	if !ok {
		t.Errorf("expression not IntegerLiteral, got %T", literal)
	}
	if literal.Value != 5 {
		t.Errorf("literal.value is not 5, got %d", literal.Value)
	}
	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral() is not 5, got %s", literal.TokenLiteral())
	}
}

func TestPrefixOperators(t *testing.T) {
	tests := []struct {
		input      string
		operator   string
		integerVal int64
	}{
		{"!5", "!", 5},
		{"-15", "-", 15},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Errorf("expected 1 statement but got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("expressionStatement is not correct, got %T", stmt)
		}

		prefixepr, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Errorf("prefix statement is not correct, got %T", prefixepr)
		}
		if prefixepr.Operator != tt.operator {
			t.Errorf("got wrong operator, expected %s but got %s", prefixepr.Operator, tt.operator)
		}
		if !testIntegerLiteral(t, prefixepr.Right, tt.integerVal) {
			return
		}
	}
}

func TestInfixOperators(t *testing.T) {
	tests := []struct {
		input     string
		leftExpr  int64
		operator  string
		rightExpr int64
	}{
		{"5+5;", 5, "+", 5},
		{"5-5;", 5, "-", 5},
		{"5*5;", 5, "*", 5},
		{"5/5;", 5, "/", 5},
		{"5>5;", 5, ">", 5},
		{"5<5;", 5, "<", 5},
		{"5==5;", 5, "==", 5},
		{"5!=5;", 5, "!=", 5},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Errorf("expected 1 statement got %d", len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("expression statement not correct, got= %T", stmt)
		}
		infixStmt, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Errorf("infix expression not correct, got=%T", infixStmt)
		}
		if !testIntegerLiteral(t, infixStmt.LeftExpr, tt.leftExpr) {
			t.Errorf("left expr not correct, expected=%d and got %s", tt.leftExpr, infixStmt.LeftExpr.String())
		}
		if infixStmt.Operator != tt.operator {
			t.Errorf("left expr not correct, expected=%s and got %s", tt.operator, infixStmt.Operator)
		}
		if !testIntegerLiteral(t, infixStmt.RightExpr, tt.rightExpr) {
			t.Errorf("left expr not correct, expected=%d and got %s", tt.rightExpr, infixStmt.RightExpr.String())
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"( 2 + 2) * 5",
			"((2 + 2) * 5)",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		actual := program.String()

		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func testIntegerLiteral(t *testing.T, il ast.Expression, val int64) bool {
	intg, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("cannot convert expression to integer literal, got=%T", intg)
		return false
	}
	if intg.Value != val {
		t.Errorf("got wrong integer value, expected=%d got %d", intg.Value, val)
		return false
	}
	if intg.TokenLiteral() != fmt.Sprintf("%d", val) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", val, intg.TokenLiteral())
		return false
	}
	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier, got=%T", exp)
		return false
	}
	if ident.Value != value {
		t.Errorf("ident.value=%s but got value=%s", ident.Value, value)
		return false
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral()=%s but got value=%s", ident.TokenLiteral(), value)
		return false
	}
	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	boolExp, ok := exp.(*ast.BooleanExpression)
	if !ok {
		t.Errorf("exp not *ast.Identifier, got=%T", exp)
		return false
	}
	if boolExp.Value != value {
		t.Errorf("ident.value=%t but got value=%t", boolExp.Value, value)
		return false
	}
	if boolExp.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("ident.TokenLiteral()=%s but got value=%t", boolExp.TokenLiteral(), value)
		return false
	}
	return true
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled, got=%T", exp)
	return false
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.OperatorExpression, got=%T(%s)", exp, exp)
	}
	if !testLiteralExpression(t, opExp.LeftExpr, left) {
		return false
	}
	if opExp.Operator != operator {
		t.Errorf("expected operator=%s, but got %s", operator, opExp.Operator)
		return false
	}
	if !testLiteralExpression(t, opExp.RightExpr, right) {
		return false
	}
	return true
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			"true",
			true,
		},
		{
			"false",
			false,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		exp, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("didn't get correct expression statement, got=%T", exp)
		}
		boolExp, ok := exp.Expression.(*ast.BooleanExpression)
		if !ok {
			t.Errorf("didn't get correct boolean expression statement, got=%T", boolExp)
		}
		if tt.expected != boolExp.Value {
			t.Errorf("boolExp.Value is not correct, expected=%t, but got=%t", tt.expected, boolExp.Value)
		}
	}
}

func TestPrecedenceOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		actual := program.String()

		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!foobar;", "!", "foobar"},
		{"-foobar;", "-", "foobar"},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpression. got=%T", stmt.Expression)
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s",
				tt.operator, exp.Operator)
		}
		if !testLiteralExpression(t, exp.Right, tt.value) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"foobar + barfoo;", "foobar", "+", "barfoo"},
		{"foobar - barfoo;", "foobar", "-", "barfoo"},
		{"foobar * barfoo;", "foobar", "*", "barfoo"},
		{"foobar / barfoo;", "foobar", "/", "barfoo"},
		{"foobar > barfoo;", "foobar", ">", "barfoo"},
		{"foobar < barfoo;", "foobar", "<", "barfoo"},
		{"foobar == barfoo;", "foobar", "==", "barfoo"},
		{"foobar != barfoo;", "foobar", "!=", "barfoo"},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		if !testInfixExpression(t, stmt.Expression, tt.leftValue,
			tt.operator, tt.rightValue) {
			return
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if(x<y){x}`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	ifExp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfExpression. got=%T",
			ifExp)
	}
	if !testInfixExpression(t, ifExp.Condition, "x", "<", "y") {
		return
	}
	if len(ifExp.Consequence.Statements) != 1 {
		t.Fatalf("expected 1 consequence, got=%d", len(ifExp.Consequence.Statements))
	}
	conseq, ok := ifExp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
			ifExp.Consequence.Statements[0])
	}
	if !testIdentifier(t, conseq.Expression, "x") {
		return
	}
	if ifExp.Alternative != nil {
		t.Errorf("exp.Alternative.Statements was not nil. got=%+v", ifExp.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if(x<y) {x} else {y}`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	ifExp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfExpression. got=%T",
			ifExp)
	}
	if !testInfixExpression(t, ifExp.Condition, "x", "<", "y") {
		return
	}
	if len(ifExp.Consequence.Statements) != 1 {
		t.Fatalf("expected 1 consequence, got=%d", len(ifExp.Consequence.Statements))
	}
	conseq, ok := ifExp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("if, Statements[0] is not ast.ExpressionStatement. got=%T",
			ifExp.Consequence.Statements[0])
	}
	if !testIdentifier(t, conseq.Expression, "x") {
		return
	}
	alternate, ok := ifExp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("else, Statements[0] is not ast.ExpressionStatement. got=%T",
			ifExp.Alternative.Statements[0])
	}
	if !testIdentifier(t, alternate.Expression, "y") {
		return
	}
}

func TestFunctionExpression(t *testing.T) {
	input := `fn(x,y) {x+y;}`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)

	if !ok {
		t.Fatalf("got wrong expression statement, got=%T", stmt)
	}
	fnExp, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("got wrong expression statement, got=%T", fnExp)
	}
	if len(fnExp.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got=%d", len(fnExp.Parameters))
	}
	testLiteralExpression(t, fnExp.Parameters[0], "x")
	testLiteralExpression(t, fnExp.Parameters[1], "y")

	if len(fnExp.Block.Statements) != 1 {
		t.Fatalf("expected 1 statement, got=%d", len(fnExp.Block.Statements))
	}

	blockExp, ok := fnExp.Block.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("got wrong expressionStatement from block of function, got=%T", blockExp)
	}
	testInfixExpression(t, blockExp.Expression, "x", "+", "y")

}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1,2*3,4+5);"

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements expects 1 statement, got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement, got=%T", stmt)
	}
	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.expression is not call expression, got=%T", exp)
	}
	if !testIdentifier(t, exp.Function, "add") {
		return
	}
	if len(exp.Arguments) != 3 {
		t.Fatalf("got wrong number of argument, got=%d", len(exp.Arguments))
	}
	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"string expression"`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral, got=%T", literal)
	}
	if literal.Value != "string expression" {
		t.Errorf("string value not correct, got=%s", literal.Value)
	}
}

func TestArrayLiteralExpression(t *testing.T) {
	input := "[1,2*2,3+3]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("array expression not correct, got=%T", stmt.Expression)
	}

	if len(array.Elements) != 3 {
		t.Errorf("expected 3 elements, got=%d", len(array.Elements))
	}
	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestIndexExpressions(t *testing.T) {
	input := "myArray[1+1]"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	indExp, ok := stmt.Expression.(*ast.IndexExpression)

	if !ok {
		t.Fatalf("got wrong index expression, got=%T", stmt.Expression)
	}

	if !testIdentifier(t, indExp.Left, "myArray") {
		t.Errorf("left expression not correct, expected=myArray and got=%s", indExp.Left)
	}

	if !testInfixExpression(t, indExp.Index, 1, "+", 1) {
		return
	}
}

func TestHashLiteralStringKeys(t *testing.T) {
	input := `{"one":1, "two":2, "three":3}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hashExp, ok := stmt.Expression.(*ast.HashLiteral)

	if !ok {
		t.Fatalf("got wrong hash literal, got=%T", hashExp)
	}
	if len(hashExp.Pairs) != 3 {
		t.Errorf("expected 3 pairs, got=%d", len(hashExp.Pairs))
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, value := range hashExp.Pairs {
		literal, ok := key.(*ast.StringLiteral)

		if !ok {
			t.Fatalf("expected string literal as a key, got=%T", literal)
		}

		expectedVal := expected[literal.Value]
		testIntegerLiteral(t, value, expectedVal)
	}
}
func TestEmptyHashLiteral(t *testing.T) {
	input := "{}"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hashExp, ok := stmt.Expression.(*ast.HashLiteral)

	if !ok {
		t.Fatalf("got wrong hash literal, got=%T", hashExp)
	}
	if len(hashExp.Pairs) != 0 {
		t.Errorf("expected 0 pairs, got=%d", len(hashExp.Pairs))
	}
}

func TestHashLiteralWithExpressions(t *testing.T) {
	input := `{"one":0+1, "two":10-8, "three":15/5}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hashExp, ok := stmt.Expression.(*ast.HashLiteral)

	if !ok {
		t.Fatalf("got wrong hash literal, got=%T", hashExp)
	}
	if len(hashExp.Pairs) != 3 {
		t.Errorf("expected 3 pairs, got=%d", len(hashExp.Pairs))
	}

	tests := map[string]func(ast.Expression){
		"one": func(e ast.Expression) {
			testInfixExpression(t, e, 0, "+", 1)
		},
		"two": func(e ast.Expression) {
			testInfixExpression(t, e, 10, "-", 8)
		},
		"three": func(e ast.Expression) {
			testInfixExpression(t, e, 15, "/", 5)
		},
	}
	for key, value := range hashExp.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Fatalf("expected string literal as a key, got=%T", key)
			continue
		}
		testFunc, ok := tests[literal.String()]
		if !ok {
			t.Fatalf("no test function returned for %q found", literal.String())
			continue
		}
		testFunc(value)
	}
}
