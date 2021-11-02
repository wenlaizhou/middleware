package middleware

import (
	"fmt"
	"testing"
	"time"
)

func TestGetToken(t *testing.T) {
	start := TimeEpoch()
	fmt.Printf("%+v\n", time.Now().Format(time.RFC3339))
	for i := 0; i < 10; i++ {
		token, err := GetToken("service", "appId",
			"appSecret")
		if err != nil {
			println(err.Error())
		}
		fmt.Printf("token为: %v, 单次耗时: %vms\n", token, TimeEpoch()-start)
	}
	fmt.Printf("总体耗时: %vms\n", TimeEpoch()-start)
}
