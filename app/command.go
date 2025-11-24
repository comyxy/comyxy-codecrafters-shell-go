package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const (
	cmdPwd     = "pwd"
	cmdCd      = "cd"
	cmdExit    = "exit"
	cmdEcho    = "echo"
	cmdType    = "type"
	cmdHistory = "history"
)

var (
	builtinMap = map[string]bool{
		cmdPwd:     true,
		cmdCd:      true,
		cmdExit:    true,
		cmdEcho:    true,
		cmdType:    true,
		cmdHistory: true,
	}
)

type IoFile struct {
	File *os.File

	TokenType TokenType
}

func NewIoFile(file *os.File, tokenType TokenType) *IoFile {
	return &IoFile{File: file, TokenType: tokenType}
}

func (f *IoFile) Close() error {
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
	RedirectIn     Redirect

	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File

	waitFunc func() error
	sh       *Shell
}

func NewCommand(sh *Shell) *Command {
	return &Command{
		Args:           nil,
		RedirectOutput: Redirect{},
		RedirectErr:    Redirect{},
		RedirectIn:     Redirect{},
		Stdin:          os.Stdin,
		Stdout:         os.Stdout,
		Stderr:         os.Stderr,
		sh:             sh,
	}
}

func (c *Command) Start() error {
	if len(c.Args) == 0 {
		return nil
	}

	cmdName := c.Args[0]

	if builtinMap[cmdName] {
		return c.startInternal()
	}

	return c.startExternal()
}

func (c *Command) Wait() error {
	if len(c.Args) == 0 {
		return nil
	}

	if c.waitFunc != nil {
		err := c.waitFunc()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) startExternal() error {
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

	reader, err := c.getInFile()
	if err != nil {
		return err
	}
	defer reader.Close()
	execCmd.Stdin = reader.File

	writer, err := c.getOutFile()
	if err != nil {
		return err
	}
	defer writer.Close()
	execCmd.Stdout = writer.File

	errWriter, err := c.getErrFile()
	if err != nil {
		return err
	}
	defer errWriter.Close()
	execCmd.Stderr = errWriter.File

	err = execCmd.Start()
	if err != nil {
		return err
	}
	c.waitFunc = func() error {
		return execCmd.Wait()
	}
	return nil
}

func (c *Command) startInternal() error {
	var wg sync.WaitGroup
	var err error

	wg.Add(1)
	go func() {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("panic: %v", e)
			}
		}()

		err = c.execInternal()
		wg.Done()
	}()

	c.waitFunc = func() error {
		wg.Wait()
		return nil
	}
	return err
}

func (c *Command) execInternal() error {
	cmdName := c.Args[0]

	var err error
	switch cmdName {
	case cmdPwd:
		err = c.execPwd()
	case cmdCd:
		err = c.execCd()
	case cmdExit:
		c.execExit()
	case cmdEcho:
		err = c.execEcho()
	case cmdType:
		c.execType()
	case cmdHistory:
		err = c.execHistory()
	}
	return err
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

	fmt.Fprintln(outWriter.File, dir)

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
			fmt.Fprintf(errFile.File, "%s: %s: %s\n", "cd", pathErr.Path, "No such file or directory")
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
	fmt.Fprintln(writer.File, r)
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

func (c *Command) execHistory() error {
	file, err := c.getOutFile()
	if err != nil {
		return err
	}
	defer file.Close()

	for idx, history := range c.sh.historyList {
		fmt.Fprintf(file.File, "    %d  %s", idx+1, history)
	}
	return nil
}

func (c *Command) getOutFile() (*IoFile, error) {
	writer := c.Stdout

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

	return NewIoFile(writer, c.RedirectOutput.TokenType), nil
}

func (c *Command) getErrFile() (*IoFile, error) {
	writer := c.Stderr

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
	return NewIoFile(writer, c.RedirectErr.TokenType), nil
}

func (c *Command) getInFile() (*IoFile, error) {
	reader := c.Stdin

	return NewIoFile(reader, c.RedirectIn.TokenType), nil
}
