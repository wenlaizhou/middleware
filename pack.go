package middleware

import (
	"os"
	"path/filepath"
)

// todo
// 静态文件编译到变量中
// 模板文件编译到变量中
// 将模板文本注册到全局数据体中
func packDir(root string) string {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		return nil
	})
	return ""
}
