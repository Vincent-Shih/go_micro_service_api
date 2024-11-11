package auth

import "strings"

type TrieNode struct {
	children map[string]*TrieNode
	isEnd    bool
	wildcard bool
}

type Trie struct {
	root *TrieNode
}

// NewTrie creates a new Trie
func NewTrie() *Trie {
	return &Trie{root: &TrieNode{children: make(map[string]*TrieNode)}}
}

// Insert adds a pattern to the Trie
func (t Trie) Insert(pattern string) {
	node := t.root
	parts := strings.Split(pattern, "/")
	for _, part := range parts {
		if part == "*" || part == "*any" {
			node.wildcard = true
		}
		if _, ok := node.children[part]; !ok {
			node.children[part] = &TrieNode{children: make(map[string]*TrieNode)}
		}
		node = node.children[part]
	}
	node.isEnd = true
}

// Search checks if a path matches any pattern in the Trie
func (t *Trie) Search(path string) bool {
	parts := strings.Split(path, "/")
	return t.searchHelper(t.root, parts, 0)
}

func (t *Trie) searchHelper(node *TrieNode, parts []string, index int) bool {
	if index == len(parts) {
		return node.isEnd
	}

	if node.wildcard {
		for i := index; i < len(parts); i++ {
			if t.searchHelper(node, parts, i+1) {
				return true
			}
		}
	}

	part := parts[index]
	if child, exists := node.children[part]; exists {
		if t.searchHelper(child, parts, index+1) {
			return true
		}
	}

	if child, exists := node.children["*"]; exists {
		if t.searchHelper(child, parts, index+1) {
			return true
		}
	}

	return false
}
