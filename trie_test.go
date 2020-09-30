package middleware

import "testing"

func TestPrintTrie(t *testing.T) {
	PrintTrie(nil, 0)
}

func TestAddTriePath(t *testing.T) {
	org := map[string]TrieNode{}
	AddTriePath("a/b/c/", func(context Context) {
		println("")
	}, org)

	PrintTrie(org, 0)
}
