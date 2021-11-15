package middleware

import (
	"regexp"
	"strings"
)

/*
注册过滤器

handle : return false 拦截请求
*/
func (t *Server) RegisterFilter(path string, handle func(Context) bool) {
	t.Lock()
	defer t.Unlock()
	if len(path) <= 0 {
		return
	}
	if strings.HasSuffix(path, "/") {
		path = path + ".*"
	}
	t.filter = append(t.filter, filterProcessor{
		handler: handle,
		pathReg: regexp.MustCompile(path),
	})

}

/*
注册过滤器

handle : return false 拦截请求
*/
func RegisterFilter(path string, handle func(Context) bool) {
	globalServer.RegisterFilter(path, handle)
}

type filterProcessor struct {
	pathReg *regexp.Regexp
	params  []string
	handler func(Context) bool
}
