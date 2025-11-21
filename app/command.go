package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	cmdExit = "exit"
	cmdEcho = "echo"
	cmdType = "type"
)

var (
	builtinMap = map[string]bool{
		cmdExit: true,
		cmdEcho: true,
		cmdType: true,
	}
)

type Command struct {
	Args []string
}

func NewCommand(args []string) *Command {
	return &Command{
		Args: args,
	}
}

func (c *Command) Exec() {
	if len(c.Args) == 0 {
		return
	}

	cmdName := c.Args[0]

	if builtinMap[cmdName] {
		c.execInternal()
		return
	}

	c.execExternal()
}

func (c *Command) execInternal() {
	cmdName := c.Args[0]

	switch cmdName {
	case cmdExit:
		c.execExit()
	case cmdEcho:
		c.execEcho()
	case cmdType:
		c.execType()
	}
}

func (c *Command) execExternal() {
	cmdName := c.Args[0]
	options := c.Args[1:]

	absPath, err := exec.LookPath(cmdName)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintf(os.Stdout, "%s: command not found\n", cmdName)
			return
		}
		fmt.Printf("fail to LookPath: %v\n", err)
		return
	}

	execCmd := exec.Command(absPath, options...)
	// Set argv to use original command name as argv[0]
	execCmd.Args[0] = cmdName
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	err = execCmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to exec %s: %v\n", absPath, err)
		return
	}
}

func (c *Command) execExit() {
	os.Exit(0)
}

func (c *Command) execEcho() {
	options := c.Args[1:]

	r := strings.Join(options, " ")
	fmt.Fprintln(os.Stdout, r)
}

func (c *Command) execType() {
	options := c.Args[1:]

	if len(options) == 0 {
		return
	}

	cmdName := options[0]

	if builtinMap[cmdName] {
		fmt.Fprintf(os.Stdout, "%s is a shell builtin\n", cmdName)
		return
	}

	absPath, err := exec.LookPath(cmdName)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintf(os.Stdout, "%s: not found\n", cmdName)
			return
		}
		fmt.Printf("fail to LookPath: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stdout, "%s is %s\n", cmdName, absPath)
	return
}
