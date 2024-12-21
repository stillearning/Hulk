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
	fmt.Printf("Feel free to type in Commands\n")
	repl.Start(os.Stdin, os.Stdout)
}
