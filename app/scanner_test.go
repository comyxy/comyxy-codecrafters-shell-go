package main

import (
	"fmt"
	"testing"
)

func TestNewScanner(t *testing.T) {
	words := NewScanner("\"/tmp/fox/f\\n17\"").Scan()
	for _, w := range words {
		fmt.Println(w)
	}
}
