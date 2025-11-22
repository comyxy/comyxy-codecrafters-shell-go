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

type Redirect struct {
	TokenType TokenType
	FileName  string
}

type Command struct {
	Args           []string
	RedirectOutput Redirect
	RedirectErr    Redirect
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

	_ = c.execExternal()
}

func (c *Command) execInternal() {
	cmdName := c.Args[0]

	switch cmdName {
	case cmdExit:
		c.execExit()
	case cmdEcho:
		_ = c.execEcho()
	case cmdType:
		c.execType()
	}
}

func (c *Command) execExit() {
	os.Exit(0)
}

func (c *Command) execEcho() error {
	options := c.Args[1:]

	writer := os.Stdout

	redirectOutput := c.RedirectOutput
	if redirectOutput.TokenType == TokenRedirectOut {
		f, err := os.Create(redirectOutput.FileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s: %s\n", "echo", redirectOutput.FileName, err)
			return err
		}
		defer f.Close()
		writer = f
	}

	redirectErr := c.RedirectErr
	if redirectErr.TokenType == TokenRedirectErr {
		f, err := os.Create(redirectErr.FileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s: %s\n", "echo", redirectErr.FileName, err)
			return err
		}
		defer f.Close()
	}

	r := strings.Join(options, " ")
	fmt.Fprintln(writer, r)
	return nil
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

func (c *Command) execExternal() error {
	cmdName := c.Args[0]
	options := c.Args[1:]

	absPath, err := exec.LookPath(cmdName)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintf(os.Stdout, "%s: command not found\n", cmdName)
			return nil
		}
		fmt.Printf("fail to LookPath: %v\n", err)
		return err
	}

	execCmd := exec.Command(absPath, options...)
	// Set argv to use original command name as argv[0]
	execCmd.Args[0] = cmdName
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	if c.RedirectOutput.TokenType == TokenRedirectOut {
		redirect := c.RedirectOutput
		f, err := os.Create(redirect.FileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s: %s\n", cmdName, redirect.FileName, err)
			return err
		}
		defer f.Close()
		execCmd.Stdout = f
	}
	if c.RedirectErr.TokenType == TokenRedirectErr {
		redirect := c.RedirectErr
		f, err := os.Create(redirect.FileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s: %s\n", cmdName, redirect.FileName, err)
			return err
		}
		defer f.Close()
		execCmd.Stderr = f
	}

	err = execCmd.Run()
	if err != nil {
		return err
	}
	return nil
}
