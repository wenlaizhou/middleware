package middleware

import "strings"

type TrieNode struct {
	Path    string
	Handler func(Context)
	Next    map[string]TrieNode
}

func AddPath(path string, handler func(Context), root TrieNode) {
	if strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-2]
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	paths := strings.Split(path, "/")
	addNext(paths, handler, root)
}

func addNext(paths []string, handler func(Context), node TrieNode) {
	if len(paths) <= 0 {
		return
	}
	path := paths[0]
	org, has := node.Next[path]
	if len(paths) == 1 {
		// 开始做事
		if has {
			org.Handler = handler
		} else {
			node.Next[path] = TrieNode{
				Path:    path,
				Handler: handler,
			}
		}
		return
	}
	if !has {
		node.Next[path] = TrieNode{
			Path: path,
			Next: map[string]TrieNode{},
		}
		addNext(paths[1:], handler, node.Next[path])
	} else {
		addNext(paths[1:], handler, node.Next[path])
	}
}

func FindPath(path string, root TrieNode) func(Context) {
	if strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-2]
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	paths := strings.Split(path, "/")
	return findNext(paths, root)
}

func findNext(paths []string, root TrieNode) func(Context) {
	if len(paths) <= 0 {
		return nil
	}
	path := paths[0]
	if len(paths) == 1 {
		return root.Handler
	}
	org, has := root.Next[path]
	if !has {
		return root.Handler
	}
	return findNext(paths[1:], org)
}
