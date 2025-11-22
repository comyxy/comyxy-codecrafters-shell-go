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
