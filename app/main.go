package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
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

		input = strings.Trim(input, "\n\r ")

		evaluate(input)
	}
}

var (
	once sync.Once
	// set once, read only
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

	pathEnv := os.Getenv("PATH")
	dirs := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			var pathErr *fs.PathError
			if errors.As(err, &pathErr) {
				continue
			}
			fmt.Printf("fail to read dir: %v\n", err)
			return
		}

		// 遍历所有条目
		for _, entry := range entries {
			if entry.Name() == cmd {
				pos := strings.Join([]string{dir, cmd}, string(os.PathSeparator))
				fmt.Fprintf(os.Stdout, "%s is in %s\n", cmd, pos)
				return
			}
		}
	}
	fmt.Fprintf(os.Stdout, "%s: not found\n", cmd)
}
