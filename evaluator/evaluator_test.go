package evaluator

import (
	"Hulk/lexer"
	"Hulk/object"
	"Hulk/parser"
	"testing"
)

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	env := object.NewEnvironment()
	return Eval(program, env)
}

func TestIntegerEvalExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func testIntegerObject(t *testing.T, evaluated object.Object, expected int64) bool {
	result, ok := evaluated.(*object.Integer)
	if !ok {
		t.Fatalf("object.Integer not correct, got=%T", result)
		return false
	}
	if result.Value != expected {
		t.Fatalf("integer value not correct, expected=%d but got=%d", expected, result.Value)
		return false
	}
	return true
}

func TestBooleanEvalExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expexcted bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Fatalf("object.Boolean not correct, got=%T", result)
		return false
	}
	if result.Value != expexcted {
		t.Fatalf("value not correct, expected=%t and got=%t", expexcted, result.Value)
		return false
	}
	return true
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!false", true},
		{"!true", false},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)

		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func testNullObject(t *testing.T, evaluated object.Object) bool {
	if evaluated != NULL {
		t.Errorf("object is not null, got=%T(%v)", evaluated, evaluated)
		return false
	}
	return true
}

func TestReturnExpressionObject(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 5;", 5},
		{"return 5*10+9/3;5;5;", 53},
		{"return 2*5;9;", 10},
		{"9;return 2*5;9;", 10},
		{`
		if(10>1){
			if(10>1){
				return 10;}
			return 1;}`, 10},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input       string
		expectedMsg string
	}{
		{
			"5+true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5+true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true;",
			"unknown operator: -BOOLEAN",
		},
		{
			"true+false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true+false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if(10>1){return true+false;}",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
		if(10>1){
			if(10>1){
				return true+false;}
			return 1;}`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"Identifier not found: foobar",
		},
		{
			`"Hello"-"World"`,
			"unknown operator: STRING - STRING",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)

		if !ok {
			t.Errorf("no error object returned, got=%T(%v)", evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expectedMsg {
			t.Errorf("wrong error message, expected=%q, got=%q", tt.expectedMsg, errObj.Message)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a=5;a;", 5},
		{"let a=5*5;a;", 25},
		{"let a=5;let b=a;b;", 5},
		{"let a=5;let b=a;let c=a+b+5;c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObjects(t *testing.T) {
	input := "fn(x) {x+2;};"

	evaluated := testEval(input)

	fnObj, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object not Function, got=%T (%v)", evaluated, evaluated)
	}
	if len(fnObj.Parameters) != 1 {
		t.Fatalf("expected 1 parameter but got=%d", len(fnObj.Parameters))
	}
	if fnObj.Parameters[0].Value != "x" {
		t.Fatalf("expected identifier 'x' but got=%s", fnObj.Parameters[0].Value)
	}
	if fnObj.Body.String() != "(x + 2)" {
		t.Fatalf("body is not (x + 2), got=%s", fnObj.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity=fn(x) {x;}; identity(5);", 5},
		{"let identity=fn(x) {return x;}; identity(5);", 5},
		{"let double=fn(x) {x*2;}; double(5);", 10},
		{"let add=fn(x, y) {x+y;}; add(5,5);", 10},
		{"fn(x) {x;}(5);", 5},
		{"let add=fn(x, y) {x+y;}; add(5+5,add(5,5));", 20},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"input string"`
	evaluated := testEval(input)

	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("evaluated objected not a string, got=%T (%v)", evaluated, evaluated)
	}
	if str.Value != "input string" {
		t.Errorf("string got wrong value, got=%q", str.Value)
	}
}

func TestStringConcat(t *testing.T) {
	input := `"Hello"+" "+"World!";`
	evaluated := testEval(input)

	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("evaluated objected not a string, got=%T (%v)", evaluated, evaluated)
	}
	if str.Value != "Hello World!" {
		t.Errorf("string got wrong value, got=%q", str.Value)
	}
}

func TestBuiltInFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("Hello World")`, 11},
		{`len(1)`, "argument to len() not supported, got INTEGER"},
		{`len("one","two")`, "wrong number of argument. got=2 want=1"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errorObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Fatalf("object is not error, got=%T (%v)", evaluated, evaluated)
				continue
			}
			if errorObj.Message != expected {
				t.Errorf("wrong error messgae, got=%s and expected=%s", errorObj.Message, expected)
			}
		}
	}
}

func TestArrayObject(t *testing.T) {
	input := "[1,2*2,3+3]"

	evaluated := testEval(input)

	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("got wrong object.Array, got=%T (%v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Errorf("expected 3 elements, got=%d", len(result.Elements))
	}
	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1,2,3][0]",
			1,
		},
		{
			"[1,2,3][1]",
			2,
		},
		{
			"[1,2,3][2]",
			3,
		},
		{
			"let i=0;[1,2,3][i]",
			1,
		},
		{
			"[1,2,3][1+1]",
			3,
		},
		{
			"let myArray=[1,2,3];myArray[2]",
			3,
		},
		{
			"let myArray=[1,2,3];myArray[2]+myArray[0]+myArray[1]",
			6,
		},
		{
			"let myArray=[1,2,3];let i=myArray[0];myArray[i]",
			2,
		},
		{
			"[1,2,3][3]",
			nil,
		},
		{
			"[1,2,3][-1]",
			nil,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestHashLiterals(t *testing.T) {
	input := `let two = "two";
			{
			"one": 10 - 9,
			two: 1 + 1,
			"thr" + "ee": 6 / 2,
			4: 4,
			true: 5,
			false: 6
			}`
	evaluated := testEval(input)
	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
	}
	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}
	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(result.Pairs))
	}
	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}
		testIntegerObject(t, pair.Value, expectedValue)
	}
}
