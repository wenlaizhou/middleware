package middleware

import "time"

const TimeFormatForClickhouse = "2006-01-02 15:04:05"

// TimeEpoch 获取毫秒时间戳
func TimeEpoch() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
