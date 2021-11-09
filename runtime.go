package middleware

import "fmt"

func BuildInfo() string {
	return ""
}

func Printf(formatter string, items ...interface{}) {
	fmt.Printf(formatter+"\n", items...)
}
