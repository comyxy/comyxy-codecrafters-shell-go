package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// // Ensures gofmt doesn't remove the "fmt" and "os" imports in stage 1 (feel free to remove this!)
// var _ = fmt.Fprint
// var _ = os.Stdout
func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		input = strings.Trim(input, "\n\r ")

		evaluate(input)
	}
}

var COMMANDS = map[string]func(options []string){
	"exit": exit,
	"echo": echo,
}

func evaluate(input string) {
	args := strings.Split(input, " ")

	cmd, options := args[0], args[1:]

	fn, ok := COMMANDS[cmd]
	if !ok {
		fmt.Fprintf(os.Stdout, "%s: command not found\n", cmd)
		return
	}

	fn(options)
}

func exit(options []string) {
	os.Exit(0)
}

func echo(options []string) {
	if len(options) == 0 {

	}
	r := strings.Join(options, " ")
	fmt.Fprintln(os.Stdout, r)
}
