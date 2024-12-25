package vm

import (
	"Hulk/code"
	"Hulk/compiler"
	"Hulk/object"
	"fmt"
)

const StackSize = 2048

const GlobalSize = 65536

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int

	globals []object.Object
}

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: make([]object.Object, GlobalSize),
	}
}

func NewWithGlobalStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {

		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			//jump over to 2 bytes to skip the constant
			ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case code.OpArray:
			numElements := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			array := vm.buildArray(vm.sp-numElements, numElements)
			vm.sp = vm.sp - numElements

			err := vm.push(array)

			if err != nil {
				return err
			}

		case code.OpHash:
			numElements := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)

			if err != nil {
				return err
			}

			vm.sp = vm.sp - numElements

			err = vm.push(hash)

			if err != nil {
				return err
			}

		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpressions(left, index)

			if err != nil {
				return err
			}

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.globals[globalIndex])

			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperations(op)

			if err != nil {
				return err
			}

		case code.OpTrue:
			err := vm.push(True)

			if err != nil {
				return err
			}

		case code.OpFalse:
			err := vm.push(False)

			if err != nil {
				return err
			}

		case code.OpEqual, code.OpGreaterThan, code.OpNotEqual:
			err := vm.executeComparison(op)

			if err != nil {
				return err
			}

		case code.OpBang:
			err := vm.executeBangOperator()

			if err != nil {
				return err
			}

		case code.OpMinus:
			err := vm.executeMinusOperator()

			if err != nil {
				return err
			}

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(vm.instructions[ip+1:])) //pos is the next instruction to jump to if not truthy
			//jump over to 2 bytes to skip the operand
			ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				ip = pos - 1
			}

		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1

		case code.OpNull:
			err := vm.push(Null)

			if err != nil {
				return err
			}

		case code.OpPop:
			vm.pop()
		}
	}
	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) executeBinaryOperations(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := right.Type()
	rightType := left.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperations(op, left, right)
	} else if leftType == object.STRING_OBJ && rightType == object.STRING_OBJ {
		return vm.executeBinaryStringOperations(op, left, right)
	}

	return fmt.Errorf("unssuported types for binary operations: %s %s", leftType, rightType)
}

func (vm *VM) executeBinaryIntegerOperations(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftVal + rightVal
	case code.OpSub:
		result = leftVal - rightVal
	case code.OpMul:
		result = leftVal * rightVal
	case code.OpDiv:
		result = leftVal / rightVal
	default:
		return fmt.Errorf("unkown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperations(op code.Opcode, left, right object.Object) error {

	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	// switch op {
	// case code.OpAdd:
	// 	result = leftVal + rightVal
	// case code.OpSub:
	// 	result = leftVal - rightVal
	// case code.OpMul:
	// 	result = leftVal * rightVal
	// case code.OpDiv:
	// 	result = leftVal / rightVal
	// default:
	// 	return fmt.Errorf("unkown integer operator: %d", op)
	// }

	return vm.push(&object.String{Value: leftVal + rightVal})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBooltoBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBooltoBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d( %s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBooltoBooleanObject(rightVal == leftVal))
	case code.OpNotEqual:
		return vm.push(nativeBooltoBooleanObject(rightVal != leftVal))
	case code.OpGreaterThan:
		return vm.push(nativeBooltoBooleanObject(leftVal > rightVal))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func nativeBooltoBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value

	case *object.Null:
		return false

	default:
		return true
	}
}

func (vm *VM) buildArray(startInd, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startInd)

	for i := startInd; i < endIndex; i++ {
		elements[i-startInd] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startInd, endInd int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startInd; i < endInd; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) executeIndexExpressions(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObj := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObj.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayObj.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}
	return vm.push(pair.Value)
}
