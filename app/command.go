package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	cmdPwd  = "pwd"
	cmdCd   = "cd"
	cmdExit = "exit"
	cmdEcho = "echo"
	cmdType = "type"
)

var (
	builtinMap = map[string]bool{
		cmdPwd:  true,
		cmdCd:   true,
		cmdExit: true,
		cmdEcho: true,
		cmdType: true,
	}
)

type RedirectFile struct {
	*os.File

	TokenType TokenType
}

func NewRedirectFile(file *os.File, tokenType TokenType) *RedirectFile {
	return &RedirectFile{File: file, TokenType: tokenType}
}

func (f *RedirectFile) Close() error {
	if (f.TokenType == TokenRedirectOut || f.TokenType == TokenRedirectOutAppend) && f.File == os.Stdout {
		return f.File.Close()
	}

	if (f.TokenType == TokenRedirectErr || f.TokenType == TokenRedirectErrAppend) && f.File == os.Stderr {
		return f.File.Close()
	}

	return nil
}

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
	case cmdPwd:
		_ = c.execPwd()
	case cmdCd:
		_ = c.execCd()
	case cmdExit:
		c.execExit()
	case cmdEcho:
		_ = c.execEcho()
	case cmdType:
		c.execType()
	}
}

func (c *Command) execPwd() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	outWriter, err := c.getOutFile()
	if err != nil {
		return err
	}
	defer outWriter.Close()

	fmt.Fprintln(outWriter, dir)

	return nil
}

func (c *Command) execCd() error {
	if len(c.Args) < 2 {
		return nil
	}

	errFile, err := c.getErrFile()
	if err != nil {
		return err
	}
	defer errFile.Close()

	dir := c.Args[1]
	if dir == "~" {
		dir, err = os.UserHomeDir()
		if err != nil {
			return err
		}
	}

	err = os.Chdir(dir)
	if err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			fmt.Fprintf(errFile, "%s: %s: %s\n", "cd", pathErr.Path, "No such file or directory")
		}
		return err
	}
	return nil
}

func (c *Command) execExit() {
	os.Exit(0)
}

func (c *Command) execEcho() error {
	options := c.Args[1:]

	writer, err := c.getOutFile()
	if err != nil {
		return err
	}
	defer writer.Close()

	errWriter, err := c.getErrFile()
	if err != nil {
		return err
	}
	defer errWriter.Close()

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
	execCmd.Stdin = os.Stdin

	writer, err := c.getOutFile()
	if err != nil {
		return err
	}
	defer writer.Close()
	execCmd.Stdout = writer

	errWriter, err := c.getErrFile()
	if err != nil {
		return err
	}
	defer errWriter.Close()
	execCmd.Stderr = errWriter

	err = execCmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (c *Command) getOutFile() (*RedirectFile, error) {
	writer := os.Stdout

	redirectOutput := c.RedirectOutput
	if redirectOutput.TokenType == TokenRedirectOut || redirectOutput.TokenType == TokenRedirectOutAppend {
		var f *os.File
		var err error
		if redirectOutput.TokenType == TokenRedirectOut {
			f, err = os.Create(redirectOutput.FileName)
		} else {
			f, err = os.OpenFile(redirectOutput.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s: %s\n", c.Args[0], redirectOutput.FileName, err)
			return nil, err
		}
		writer = f
	}

	return NewRedirectFile(writer, c.RedirectOutput.TokenType), nil
}

func (c *Command) getErrFile() (*RedirectFile, error) {
	writer := os.Stderr
	redirectErr := c.RedirectErr
	if redirectErr.TokenType == TokenRedirectErr || redirectErr.TokenType == TokenRedirectErrAppend {
		var f *os.File
		var err error
		if redirectErr.TokenType == TokenRedirectErr {
			f, err = os.Create(redirectErr.FileName)
		} else {
			f, err = os.OpenFile(redirectErr.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s: %s\n", c.Args[0], redirectErr.FileName, err)
			return nil, err
		}
		writer = f
	}
	return NewRedirectFile(writer, c.RedirectErr.TokenType), nil
}
