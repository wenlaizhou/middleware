package middleware

import "time"

var schedule []map[string]func()

//注册定时任务
//
//时间单位 秒
func Schedule(name string, timeSchedule int, fun func()) {
	mLogger.InfoLn("注册定时任务: %v 间隔时间: %v", name, timeSchedule)
	go func(timeSchedule int, fun func()) {
		for {
			time.Sleep(time.Second * time.Duration(timeSchedule))
			fun()
		}
	}(timeSchedule, fun)
}
