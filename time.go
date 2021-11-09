package middleware

import "time"

// 获取毫秒时间戳
func TimeEpoch() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
