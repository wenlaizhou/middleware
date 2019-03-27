package middleware

import "github.com/wenlaizhou/middleware"

import "testing"

func TestIsActive(t *testing.T) {
	println(middleware.IsActive("www.baidu.com", 80))
}
