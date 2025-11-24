package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	for {
		fmt.Fprint(os.Stdout, "$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		input = strings.Trim(input, "\n\r")

		tokens := NewScanner(input).Scan()

		cmds := NewParser(tokens).ParsePipeline()

		if len(cmds) == 0 {
			continue
		}

		for i := 0; i < len(cmds)-1; i++ {
			curCmd := cmds[i]
			nextCmd := cmds[i+1]

			pr, pw, err := os.Pipe()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				break
			}

			nextCmd.Stdin = pr
			curCmd.Stdout = pw
		}

		for i := len(cmds) - 1; i >= 0; i-- {
			err := cmds[i].Start()
			if err != nil {
				break
			}
		}

		for i := 0; i < len(cmds); i++ {
			err := cmds[i].Wait()
			if err != nil {
				break
			}
			if i > 0 {
				cmds[i].Stdin.Close()
			}
			if i < len(cmds)-1 {
				cmds[i].Stdout.Close()
			}
		}
	}
}
