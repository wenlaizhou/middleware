package middleware

import (
	"errors"
	"fmt"
	"time"
)

type ScheduleData struct {
	Name         string `json:"name"`
	TimeSchedule int    `json:"timeSchedule"`
	handle       chan string
}

var scheduleRunner = map[string]ScheduleData{}

func SchedulePause(name string) {
	s, hasS := scheduleRunner[name]
	if !hasS {
		return
	}
	s.handle <- "pause"
}

func ScheduleContinue(name string) {
	s, hasS := scheduleRunner[name]
	if !hasS {
		return
	}
	s.handle <- "continue"
}

func ScheduleStop(name string) {
	s, hasS := scheduleRunner[name]
	if !hasS {
		return
	}
	s.handle <- "stop"
}

// 注册定时任务
//
// 时间单位 秒
func Schedule(name string, timeSchedule int, fun func()) {
	mLogger.InfoF("注册定时任务: %v 间隔时间: %v秒", name, timeSchedule)
	handle := make(chan string)
	scheduleRunner[name] = ScheduleData{
		Name:         name,
		TimeSchedule: timeSchedule,
		handle:       handle,
	}
	go func(timeSchedule int, sig chan string, fun func()) {
		defer func() {
			if err := recover(); err != nil {
				mLogger.ErrorF("schedule 退出: %#v", err)
			}
		}()
		for {
			select {
			case signal := <-sig:
				switch signal {
				case "pause":
				pause:
					select {
					case nextSignal := <-sig:
						switch nextSignal {
						case "continue":
							break
						default:
							goto pause
						}
						break
					}
				case "continue":
					break
				case "stop":
					panic(errors.New("force stop"))
				default:
					break
				}
			default:
				break
			}
			time.Sleep(time.Second * time.Duration(timeSchedule))
			fun()
		}
	}(timeSchedule, handle, fun)
}

func RegisterScheduleService(path string) []SwaggerPath {

	RegisterHandler(path, func(context Context) {
		context.ApiResponse(0, "", scheduleRunner)
	})

	pausePath := fmt.Sprintf("%v/pause", path)
	pauseSwagger := SwaggerBuildPath(pausePath, "middleware", "post", "middleware schedule pause")
	pauseSwagger.AddParameter(SwaggerParameter{
		Name:        "body",
		Description: "json类型, name:  任务名称",
		Example: `{
	"name" : ""
}`,
		In:       "body",
		Required: true,
	})
	RegisterHandler(pausePath, func(context Context) {
		params, err := context.GetJSON()
		if err != nil {
			context.ApiResponse(-1, err.Error(), nil)
			return
		}
		name, hasName := params["name"]
		if !hasName {
			context.ApiResponse(-1, "no name", nil)
			return
		}
		SchedulePause(fmt.Sprintf("%v", name))
		context.ApiResponse(0, "", nil)
	})

	continuePath := fmt.Sprintf("%v/continue", path)
	continueSwagger := SwaggerBuildPath(continuePath, "middleware", "post", "middleware schedule continue")
	continueSwagger.AddParameter(SwaggerParameter{
		Name:        "body",
		Description: "json类型, name:  任务名称",
		Example: `{
	"name" : ""
}`,
		In:       "body",
		Required: true,
	})
	RegisterHandler(continuePath, func(context Context) {
		params, err := context.GetJSON()
		if err != nil {
			context.ApiResponse(-1, err.Error(), nil)
			return
		}
		name, hasName := params["name"]
		if !hasName {
			context.ApiResponse(-1, "no name", nil)
			return
		}
		ScheduleContinue(fmt.Sprintf("%v", name))
		context.ApiResponse(0, "", nil)
	})

	stopPath := fmt.Sprintf("%v/stop", path)
	stopSwagger := SwaggerBuildPath(stopPath, "middleware", "post", "middleware schedule stop")
	stopSwagger.AddParameter(SwaggerParameter{
		Name:        "body",
		Description: "json类型, name:  任务名称",
		Example: `{
	"name" : ""
}`,
		In:       "body",
		Required: true,
	})
	RegisterHandler(stopPath, func(context Context) {
		params, err := context.GetJSON()
		if err != nil {
			context.ApiResponse(-1, err.Error(), nil)
			return
		}
		name, hasName := params["name"]
		if !hasName {
			context.ApiResponse(-1, "no name", nil)
			return
		}
		ScheduleStop(fmt.Sprintf("%v", name))
		context.ApiResponse(0, "", nil)
	})

	return []SwaggerPath{
		SwaggerBuildPath(path, "middleware", "get", "middleware schedule"),
		pauseSwagger, continueSwagger, stopSwagger,
	}
}
