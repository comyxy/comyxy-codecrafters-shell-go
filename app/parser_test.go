package main

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	tokens := NewScanner("ls /tmp/baz > /tmp/foo/baz.md").Scan()
	cmd := NewParser(tokens).Parse()

	fmt.Println(cmd)
}

func TestParser2(t *testing.T) {
	tokens := NewScanner("echo test | head").Scan()
	cmd := NewParser(tokens).Parse()

	fmt.Println(cmd)
}
