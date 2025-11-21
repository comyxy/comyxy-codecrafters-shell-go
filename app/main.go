package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// // Ensures gofmt doesn't remove the "fmt" and "os" imports in stage 1 (feel free to remove this!)
// var _ = fmt.Fprint
// var _ = os.Stdout
func main() {
	initCommands()

	for {
		fmt.Fprint(os.Stdout, "$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		evaluate(input)
	}
}

var (
	once sync.Once
	// COMMANDS set once, read only
	COMMANDS = make(map[string]func(options []string))
)

func initCommands() {
	once.Do(func() {
		COMMANDS = map[string]func(options []string){
			"exit": exit,
			"echo": echo,
			"type": typeFn,
		}
	})
}

func evaluate(input string) {
	input = strings.Trim(input, "\n\r ")

	args := NewParser().Parse(input)

	cmd, options := args[0], args[1:]

	fn, ok := COMMANDS[cmd]
	if !ok {
		external(cmd, options)
		return
	}

	fn(options)
}

func exit(options []string) {
	os.Exit(0)
}

func echo(options []string) {
	r := strings.Join(options, " ")
	fmt.Fprintln(os.Stdout, r)
}

func typeFn(options []string) {
	if len(options) == 0 {
		return
	}
	cmd := options[0]

	_, ok := COMMANDS[cmd]
	if ok {
		fmt.Fprintf(os.Stdout, "%s is a shell builtin\n", cmd)
		return
	}

	absPath, err := exec.LookPath(cmd)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintf(os.Stdout, "%s: not found\n", cmd)
			return
		}
		fmt.Printf("fail to LookPath: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stdout, "%s is %s\n", cmd, absPath)
	return
}

func external(cmd string, options []string) {
	absPath, err := exec.LookPath(cmd)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintf(os.Stdout, "%s: command not found\n", cmd)
			return
		}
		fmt.Printf("fail to LookPath: %v\n", err)
		return
	}

	c := exec.Command(absPath, options...)
	// Set argv to use original command name as argv[0]
	c.Args[0] = cmd
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	err = c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to exec %s: %v\n", absPath, err)
		return
	}
}
