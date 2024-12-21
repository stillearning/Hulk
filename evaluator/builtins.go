package evaluator

import (
	"Hulk/object"
	"fmt"
)

var builtins = map[string]*object.BuiltIn{
	"len": &object.BuiltIn{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("wrong number of argument. got=%d want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}

			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}

			default:
				return NewError("argument to len() not supported, got %s",
					args[0].Type())
			}
		},
	},
	"first": &object.BuiltIn{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("wrong number of argument. got=%d want=1", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("expected an array for function FIRST, got=%s", args[0].Type())
			}
			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}
			return NULL
		},
	},
	"last": &object.BuiltIn{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("wrong number of argument. got=%d want=1", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("expected an array for function LAST, got=%s", args[0].Type())
			}
			arr := args[0].(*object.Array)
			size := len(arr.Elements)
			if size > 0 {
				return arr.Elements[size-1]
			}
			return NULL
		},
	},
	"rest": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return NewError("wrong number of argument. got=%d want=1", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("expected an array for function REST, got=%s", args[0].Type())
			}
			arr := args[0].(*object.Array)
			size := len(arr.Elements)
			if size > 0 {
				newEles := make([]object.Object, size-1, size-1)
				copy(newEles, arr.Elements[1:size])
				return &object.Array{Elements: newEles}
			}
			return NULL
		},
	},
	"push": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return NewError("wrong number of argument. got=%d want=2", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return NewError("expected an array for function PUSH, got=%s", args[0].Type())
			}
			arr := args[0].(*object.Array)
			size := len(arr.Elements)
			newEles := make([]object.Object, size+1, size+1)
			copy(newEles, arr.Elements[1:size])
			newEles[size] = args[1]
			return &object.Array{Elements: newEles}
		},
	},
	"puts": {
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
}
