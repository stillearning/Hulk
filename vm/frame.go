package vm

import (
	"Hulk/code"
	"Hulk/object"
)

// fn points to compiled function reference by the frame, and the ip is the instruction in this frame for this function
type Frame struct {
	fn          *object.CompiledFunction
	ip          int
	basePointer int
}

func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{fn: fn, ip: -1, basePointer: basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
