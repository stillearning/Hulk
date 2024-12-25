package main

import (
	"Hulk/repl"
	"fmt"
	"os"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s!, This is Hulk Programming language!\n", user.Username)

	arguments := os.Args

	if len(arguments) == 1 {
		fmt.Printf("Feel free to type in Commands\n")
		repl.Start(os.Stdin, os.Stdout)
	} else if len(os.Args) == 3 {
		if os.Args[1] == "rune" {
			fileName := os.Args[2]

			// Open the file
			file, err := os.Open(fileName)
			if err != nil {
				fmt.Printf("Error opening file: %v\n", err)
				return
			}
			defer file.Close()

			// Pass the file to a function for processing
			repl.CompileFile(file, os.Stdout)
		} else {
			fmt.Errorf("invalid arguments: %s", arguments[1])
		}
	} else {
		fmt.Errorf("Invalid number of arguments")
	}
}
