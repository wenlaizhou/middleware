package middleware

import "strings"

type TrieNode struct {
	Path    string
	Handler func(Context) //root handler 为默认处理器
	Next    map[string]TrieNode
}

func NewTrieNode(defaultHandler func(Context)) TrieNode {
	return TrieNode{
		Path:    "/",
		Handler: defaultHandler,
		Next:    map[string]TrieNode{},
	}
}

func (this TrieNode) AddPath(path string, handler func(Context)) {
	if strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-1]
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	paths := strings.Split(path, "/")
	addNext(paths, handler, this)
}

func addNext(paths []string, handler func(Context), node TrieNode) {
	if len(paths) <= 0 {
		return
	}
	path := paths[0]
	if node.Next == nil {
		node.Next = map[string]TrieNode{}
	}
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

func (this TrieNode) FindPath(path string) func(Context) {
	if strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-1]
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	paths := strings.Split(path, "/")
	return findNext(paths, this)
}

func findNext(paths []string, root TrieNode) func(Context) {
	if len(paths) <= 0 {
		return nil
	}
	path := paths[0]
	org, has := root.Next[path]
	if len(paths) == 1 {
		if has {
			return org.Handler
		}
		return root.Handler
	}
	if !has {
		return root.Handler
	}
	return findNext(paths[1:], org)
}
