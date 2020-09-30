package middleware

import "strings"

type TrieNode struct {
	Key      string
	Handler  func(Context)
	Children map[string]TrieNode
}

func AddTriePath(path string, handler func(Context), nodes map[string]TrieNode) {
	if len(path) <= 0 || handler == nil || nodes == nil {
		return
	}
	if strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-2]
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	if !strings.Contains(path, "/") {
		originNode, has := nodes[path]
		if has {
			originNode.Handler = handler
			return
		}
		node := TrieNode{
			Key:      path,
			Handler:  handler,
			Children: nil,
		}
		nodes[path] = node
		return
	}
	paths := strings.Split(path, "/")
	var tmpNodes = &nodes
	for i := 0; i < len(paths); i++ {
		p := paths[i]
		originNode, has := (*tmpNodes)[p]
		if !has {
			(*tmpNodes)[p] = *createTrie(paths[i:], handler)
			return
		}
		if originNode.Children == nil {
			originNode.Children = map[string]TrieNode{}
		}
		tmpNodes = &originNode.Children
		continue
	}
}

func createTrie(paths []string, handler func(Context)) *TrieNode {
	if len(paths) <= 0 {
		return nil
	}
	if len(paths) == 1 {
		return &TrieNode{
			Key:      paths[0],
			Handler:  handler,
			Children: nil,
		}
	}
	result := TrieNode{
		Key:      paths[0],
		Handler:  nil,
		Children: nil,
	}
	var tmp = &result
	for i := 1; i < len(paths); i++ {
		if i == len(paths)-1 {
			tmp.Children = map[string]TrieNode{
				paths[i]: {
					Key:     paths[i],
					Handler: handler,
				},
			}
		} else {
			tmp.Children = map[string]TrieNode{
				paths[i]: {
					Key: paths[i],
				},
			}
		}
		var x = tmp.Children[paths[i]]
		tmp = &x
	}
	return &result
}

func pick(nodes map[string]TrieNode, path string) func(Context) {
	query := nodes
	if strings.HasSuffix(path, "/") {
		path = path[0 : len(path)-2]
	}
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	var pre TrieNode
	paths := strings.Split(path, "/")
	for i := 0; i < len(paths); i++ {
		p := paths[i]
		if i == len(paths)-1 {
			node, has := query[p]
			if has {
				return node.Handler
			} else {
				return pre.Handler
			}
		}
		node, has := query[p]
		if has {
			pre = node
			query = node.Children
			continue
		}
		return pre.Handler
	}
	return nil
}

func PrintTrie(nodes map[string]TrieNode, level int) {
	if nodes == nil {
		return
	}
	for k, v := range nodes {
		println(strings.Repeat("--", level), k, ":", v.Handler)
		PrintTrie(v.Children, level+1)
	}
}
