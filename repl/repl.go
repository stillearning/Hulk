package repl

import (
	"Hulk/ast"
	"Hulk/compiler"
	"Hulk/lexer"
	"Hulk/object"
	"Hulk/parser"
	"Hulk/vm"
	"bufio"
	"fmt"
	"io"
	"os"
)

const PROMPT = "#>"

func Parsing(line string, out io.Writer) *ast.Program {
	l := lexer.New(line)
	p := parser.New(l)

	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		printParseErrors(out, p.Errors())
	}

	return program
}

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	// env := object.NewEnvironment()

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalSize)
	symbolTable := compiler.NewSymbolTable()

	for {
		fmt.Print(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()

		program := Parsing(line, out)

		// evaluated := evaluator.Eval(program, env)
		// if evaluated != nil {
		// 	io.WriteString(out, evaluated.Inspect())
		// 	io.WriteString(out, "\n")
		// }

		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed: \n %s\n", err)
			continue
		}

		code := comp.Bytecode()
		constants = code.Constants

		machine := vm.NewWithGlobalStore(code, globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Bytecode failed: \n %s\n", err)
		}

		LastPoppedStackElem := machine.LastPoppedStackElem()
		io.WriteString(out, LastPoppedStackElem.Inspect())
		io.WriteString(out, "\n")
	}
}

func printParseErrors(out io.Writer, errors []string) {
	io.WriteString(out, " parser errors:\n")
	for _, err := range errors {
		io.WriteString(out, "\t"+err+"\n")
	}
}

func CompileFile(file *os.File, out io.Writer) {
	scanner := bufio.NewScanner(file)

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalSize)
	symbolTable := compiler.NewSymbolTable()

	for scanner.Scan() {
		// Print each line (or handle it as needed)
		// fmt.Println(scanner.Text())

		program := Parsing(scanner.Text(), out)

		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed: \n %s\n", err)
			continue
		}

		code := comp.Bytecode()
		constants = code.Constants

		machine := vm.NewWithGlobalStore(code, globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Bytecode failed: \n %s\n", err)
		}

		LastPoppedStackElem := machine.LastPoppedStackElem()
		io.WriteString(out, LastPoppedStackElem.Inspect())
		io.WriteString(out, "\n")
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

}
