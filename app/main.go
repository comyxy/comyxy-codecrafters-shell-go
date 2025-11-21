package main

import (
	"bufio"
	"fmt"
	"os"
)

// // Ensures gofmt doesn't remove the "fmt" and "os" imports in stage 1 (feel free to remove this!)
// var _ = fmt.Fprint
// var _ = os.Stdout
func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		command = command[:len(command)-1]

		if command == "exit" {
			os.Exit(0)
		}

		fmt.Fprintf(os.Stdout, "%s: command not found\n", command)
	}
}
