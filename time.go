package middleware

import (
	"fmt"
	"time"
)

const TimeFormatForClickhouse = "2006-01-02 15:04:05"

// TimeEpoch 获取毫秒时间戳
func TimeEpoch() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func TimeParse() {
	t, _ := time.ParseInLocation(TimeFormatForClickhouse, "2023-02-15 11:05:00", time.Local)
	fmt.Println(t)
	fmt.Println(time.Now())
	fmt.Println(time.Since(t))
	bef := time.Now().Add(-30 * time.Second)
	fmt.Println(time.Since(bef))
	fmt.Println(time.Since(t))
	return
}
