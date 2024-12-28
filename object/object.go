package object

import (
	"Hulk/ast"
	"Hulk/code"
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJ           = "INTEGER"
	BOOLEAN_OBJ           = "BOOLEAN"
	NULL_OBJ              = "NULL"
	RETURN_VALUE_OBJ      = "RETURN_VALUE"
	ERROR_OBJ             = "ERROR"
	FUNCTION_OBJ          = "FUNCTION"
	STRING_OBJ            = "STRING"
	BUILTIN_OBJ           = "BUILTIN"
	ARRAY_OBJ             = "ARRAY"
	HASH_OBJ              = "HASH"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION_OBJ"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}

func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType {
	return BOOLEAN_OBJ
}

func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
}

type Null struct{}

func (n *Null) Type() ObjectType {
	return NULL_OBJ
}

func (n *Null) Inspect() string {
	return "null"
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType {
	return RETURN_VALUE_OBJ
}

func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}

type Error struct {
	Message string
}

func (err *Error) Type() ObjectType {
	return ERROR_OBJ
}

func (err *Error) Inspect() string {
	return "ERROR: " + err.Message
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (fn *Function) Type() ObjectType {
	return FUNCTION_OBJ
}

func (fn *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fn.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(fn.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type String struct {
	Value string
}

func (str *String) Type() ObjectType {
	return STRING_OBJ
}

func (str *String) Inspect() string {
	return str.Value
}

type BuiltinFunction func(args ...Object) Object

type BuiltIn struct {
	Fn BuiltinFunction
}

func (bltin *BuiltIn) Type() ObjectType {
	return BUILTIN_OBJ
}

func (bltin *BuiltIn) Inspect() string {
	return "builtin function"
}

type Array struct {
	Elements []Object
}

func (arr *Array) Type() ObjectType {
	return ARRAY_OBJ
}

func (arr *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, ar := range arr.Elements {
		elements = append(elements, ar.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type Hashable interface {
	HashKey() HashKey
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (str *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(str.Value))
	return HashKey{Type: str.Type(), Value: h.Sum64()}
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType {
	return HASH_OBJ
}

func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

type CompiledFunction struct {
	Instructions code.Instructions
	NumLocals    int
}

func (cf *CompiledFunction) Type() ObjectType {
	return COMPILED_FUNCTION_OBJ
}

func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}
