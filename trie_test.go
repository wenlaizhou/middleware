package middleware

import "testing"

func TestAddTriePath(t *testing.T) {
	org := map[string]TrieNode{}
	AddTriePath("a/b/c/", func(context Context) {
		println("")
	}, org)

	PrintTrie(org, 0)
}
