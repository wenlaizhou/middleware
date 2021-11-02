package middleware

import "time"

func TimeEpoch() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
