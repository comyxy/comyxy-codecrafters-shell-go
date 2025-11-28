package internal

import "fmt"

type Trie struct {
	children map[rune]*Trie
	isEnd    bool
}

func NewTrie() *Trie {
	return &Trie{
		children: make(map[rune]*Trie),
		isEnd:    false,
	}
}

func (t *Trie) Insert(s string) {
	node := t
	for _, c := range s {
		c := c
		_, ok := node.children[c]
		if !ok {
			node.children[c] = NewTrie()
		}
		node = node.children[c]
	}
	node.isEnd = true
	return
}

func (t *Trie) Search(s string) bool {
	node := t.walkTrie(s)
	return node != nil && node.isEnd
}

func (t *Trie) StartsWith(prefix string) bool {
	node := t.walkTrie(prefix)
	return node != nil
}

func (t *Trie) walkTrie(prefix string) *Trie {
	node := t
	for _, c := range prefix {
		c := c
		_, ok := node.children[c]
		if !ok {
			return nil
		}
		node = node.children[c]
	}
	return node
}

func (t *Trie) FindCompletion(prefix string) []string {
	node := t.walkTrie(prefix)
	if node == nil {
		return nil
	}

	res := make([]string, 0, len(node.children))

	var traverse func(node *Trie, ps string)
	traverse = func(node *Trie, ps string) {
		if node == nil {
			return
		}
		if node.isEnd {
			res = append(res, ps)
		}
		for c, nextTrie := range node.children {
			traverse(nextTrie, ps+string(c))
		}
	}
	traverse(node, prefix)

	return res
}

func (t *Trie) Print() {
	res := make([]string, 0)

	var traverse func(node *Trie, ps string)
	traverse = func(node *Trie, ps string) {
		if node == nil {
			return
		}
		if node.isEnd {
			res = append(res, ps)
		}
		for c, nextTrie := range node.children {
			traverse(nextTrie, ps+string(c))
		}
	}

	traverse(t, ``)
	for _, v := range res {
		fmt.Println(v)
	}
}
