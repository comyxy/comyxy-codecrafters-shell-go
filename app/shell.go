package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

type Shell struct {
	historyList       []string
	appendHistoryList []string

	completer *readline.PrefixCompleter
}

func NewShell() *Shell {
	completer := readline.NewPrefixCompleter(
		readline.PcItem("echo"),
		readline.PcItem("exit"),
	)

	sh := &Shell{
		completer: completer,
	}

	return sh
}

func (sh *Shell) Run() {

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "$ ",
		HistoryFile:  "/tmp/my-shell.history",
		AutoComplete: sh.completer,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()

	historyFile := os.Getenv("HISTFILE")
	if historyFile != "" {
		err := sh.readHistory(historyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}
		defer func() {
			sh.dumpHistory(historyFile)
		}()
	}

	for {

		input, err := rl.Readline()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		input = strings.Trim(input, "\n\r")

		sh.appendHistory(input)

		tokens := NewScanner(input).Scan()

		cmds := NewParser(tokens).ParsePipeline(sh)

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
				if errors.Is(err, errExit) {
					goto finish
				}
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

finish:
}

func (sh *Shell) appendHistory(input string) {
	sh.historyList = append(sh.historyList, input)
	sh.appendHistoryList = append(sh.appendHistoryList, input)
}

func (sh *Shell) readHistory(path string) error {
	historyFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer historyFile.Close()
	reader := bufio.NewReader(historyFile)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		sh.historyList = append(sh.historyList, line)
	}
	return nil
}

func (sh *Shell) dumpHistory(path string) error {
	historyFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer historyFile.Close()
	for _, history := range sh.historyList {
		fmt.Fprintf(historyFile, "%s\n", history)
	}
	return nil
}
