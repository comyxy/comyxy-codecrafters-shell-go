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

		input = strings.Trim(input, "\n\r")

		tokens := NewScanner(input).Scan()

		cmds := NewParser(tokens).ParsePipeline()

		if len(cmds) == 0 {
			continue
		} else if len(cmds) == 1 {
			cmds[0].Exec()
		} else {
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
				cmds[i].Start()
			}

			for i := 0; i < len(cmds); i++ {
				cmds[i].Wait()
				if i > 0 {
					cmds[i].Stdin.Close()
				}
				if i < len(cmds)-1 {
					cmds[i].Stdout.Close()
				}
			}
		}
	}
}
