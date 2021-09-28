package middleware

import (
	"fmt"
	"testing"
	"time"
)

func TestGetToken(t *testing.T) {
	start := time.Now().UnixMilli()
	fmt.Printf("%+v\n", time.Now().Format(time.RFC3339))
	for i := 0; i < 10; i++ {
		token, _ := GetToken("service", "appId",
			"appSecret")
		fmt.Printf("token为: %v, 单次耗时: %vms\n", token, time.Now().UnixMilli()-start)
	}
	fmt.Printf("总体耗时: %vms\n", time.Now().UnixMilli()-start)
}
