package evaluator

import (
	"Hulk/ast"
	"Hulk/object"
	"fmt"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalStatements(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.BooleanExpression:
		return returnNativeBooleanObject(node.Value, env)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixObject(node.Operator, right, env)

	case *ast.InfixExpression:
		left := Eval(node.LeftExpr, env)
		if isError(left) {
			return left
		}
		right := Eval(node.RightExpr, env)
		if isError(right) {
			return right
		}
		return evalInfixObject(left, node.Operator, right, env)

	case *ast.IfExpression:
		return evalIfExpressionObject(node, env)

	case *ast.BlockStatement:
		return evalBlockStatements(node, env)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return val

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Block
		return &object.Function{Parameters: params, Body: body, Env: env}

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	}
	return nil
}

func evalStatements(program *ast.Program, env *object.Environment) object.Object {
	var obj object.Object

	for _, stmt := range program.Statements {
		obj = Eval(stmt, env)

		switch obj := obj.(type) {
		case *object.ReturnValue:
			return obj.Value
		case *object.Error:
			return obj
		}
	}
	return obj
}

func evalBlockStatements(block *ast.BlockStatement, e *object.Environment) object.Object {
	var obj object.Object

	for _, stmt := range block.Statements {
		obj = Eval(stmt, e)
		if obj != nil {
			if obj.Type() == object.RETURN_VALUE_OBJ || obj.Type() == object.ERROR_OBJ {
				return obj
			}
		}
	}
	return obj
}

func returnNativeBooleanObject(input bool, env *object.Environment) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return NewError("Identifier not found: " + node.Value)
}

func evalPrefixObject(op string, right object.Object, env *object.Environment) object.Object {
	switch op {
	case "!":
		return returnBangOperatorExpression(right, env)
	case "-":
		return returnMinusOperatorExpression(right, env)
	default:
		return NewError("unknown operator: %s %s", op, right.Type())
	}
}

func returnBangOperatorExpression(right object.Object, env *object.Environment) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func returnMinusOperatorExpression(right object.Object, env *object.Environment) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return NewError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixObject(left object.Object, op string, right object.Object, env *object.Environment) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalInfixIntegerExpression(left, op, right, env)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalInfixStringExpression(left, op, right, env)
	case left.Type() != right.Type():
		return NewError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	case op == "==":
		return returnNativeBooleanObject(left == right, env)
	case op == "!=":
		return returnNativeBooleanObject(left != right, env)
	default:
		return NewError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func evalInfixIntegerExpression(left object.Object, op string, right object.Object, env *object.Environment) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch op {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case ">":
		return returnNativeBooleanObject(leftVal > rightVal, env)
	case "<":
		return returnNativeBooleanObject(leftVal < rightVal, env)
	case "==":
		return returnNativeBooleanObject(leftVal == rightVal, env)
	case "!=":
		return returnNativeBooleanObject(leftVal != rightVal, env)
	default:
		return NewError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func evalIfExpressionObject(ifExp *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ifExp.Condition, env)

	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ifExp.Consequence, env)
	} else if ifExp.Alternative != nil {
		return Eval(ifExp.Alternative, env)
	} else {
		return NULL
	}
}

func isTruthy(condition object.Object) bool {
	switch condition {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func NewError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalExpressions(args []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range args {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func evalInfixStringExpression(left object.Object, op string, right object.Object, env *object.Environment) object.Object {
	if op != "+" {
		return NewError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	return &object.String{Value: leftVal + rightVal}
}

func evalIndexExpression(left object.Object, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	default:
		return NewError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(left object.Object, index object.Object) object.Object {
	arrayObj := left.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObj.Elements)) - 1

	if idx < 0 || idx > max {
		return NULL
	}
	return arrayObj.Elements[idx]

}

func evalHashLiteral(hl *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range hl.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}
		hashkey, ok := key.(object.Hashable)
		if !ok {
			return NewError("unusable as hashkey: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashkey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}
	return &object.Hash{Pairs: pairs}
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendedFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.BuiltIn:
		return fn.Fn(args...)
	default:
		return NewError("not a function: %s", fn.Type())
	}
}

func extendedFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}
