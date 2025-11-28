package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
	"github.com/codecrafters-io/shell-starter-go/internal"
)

const (
	prompt = "$ "
)

type Shell struct {
	historyList       []string
	appendHistoryList []string

	completer readline.AutoCompleter
}

func NewShell() *Shell {
	completer := NewMyAutoCompleter()

	sh := &Shell{
		completer: completer,
	}

	return sh
}

func (sh *Shell) Run() {

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       prompt,
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

type myAutoCompleter struct {
	trie *internal.Trie

	tabPressed bool
}

func NewMyAutoCompleter() readline.AutoCompleter {
	trie := internal.NewTrie()

	for _, cmd := range getExternCommand() {
		trie.Insert(cmd)
	}

	trie.Insert("echo")
	trie.Insert("exit")

	//trie.Insert("xyz_ant")
	//trie.Insert("xyz_ant_owl")
	//trie.Insert("xyz_ant_owl_pig")

	//trie.Print()

	return &myAutoCompleter{
		trie: trie,
	}
}

func (m *myAutoCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	strLine := string(line)
	l := strings.Fields(strLine)
	if len(l) == 0 {
		return nil, 0
	}

	prefix := l[0]
	completion := m.trie.FindCompletion(prefix)

	if len(completion) == 0 {
		m.tabPressed = false
		fmt.Fprintf(os.Stdout, "\x07")
		return nil, 0
	} else if len(completion) == 1 {
		// 直接补全需要依赖readline
		m.tabPressed = false
		strCompletion0 := completion[0]
		strCompletion0, _ = strings.CutPrefix(strCompletion0, prefix)
		strCompletion0 += " "
		return [][]rune{[]rune(strCompletion0)}, len(prefix)
	} else {
		sort.Slice(completion, func(i, j int) bool {
			return len(completion[i]) < len(completion[j])
		})
		isChain := isCompletionPrefixChain(completion)
		if isChain {
			strCompletion0 := completion[0]
			strCompletion0, _ = strings.CutPrefix(strCompletion0, prefix)
			return [][]rune{[]rune(strCompletion0)}, len(prefix)
		} else if !m.tabPressed {
			m.tabPressed = true
			fmt.Fprintf(os.Stdout, "\x07")
			return nil, 0
		} else {
			// 为了通过测试需要手动输出到stdout
			m.tabPressed = false
			sort.Strings(completion)
			fmt.Fprintf(os.Stdout, "\n%s\n", strings.Join(completion, "  "))
			fmt.Fprintf(os.Stdout, "%s%s", prompt, strLine)
			return nil, 0
		}
	}
}

func isCompletionPrefixChain(strs []string) bool {
	for i := 0; i < len(strs)-1; i++ {
		cur := strs[i]
		next := strs[i+1]

		ok := strings.HasPrefix(next, cur)
		if !ok {
			return false
		}
	}
	return true
}

func getExternCommand() []string {
	var cmds []string

	path := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if path == dir {
				return nil
			}

			r, err := filepath.Rel(dir, path)
			if err != nil {
				return nil
			}
			cmds = append(cmds, r)
			return nil
		})
	}
	return cmds
}
