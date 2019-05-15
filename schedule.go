package middleware

import "time"

// 注册定时任务
//
// 时间单位 秒
func Schedule(name string, timeSchedule int, fun func()) {
	mLogger.InfoF("注册定时任务: %v 间隔时间: %v秒", name, timeSchedule)
	go func(timeSchedule int, fun func()) {
		for {
			time.Sleep(time.Second * time.Duration(timeSchedule))
			fun()
		}
	}(timeSchedule, fun)
}
