package middleware

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type task struct {
	Name       string `json:"name"`
	StartEpoch int64  `json:"startEpoch"`
	EndEpoch   int64  `json:"endEpoch"`
	Runner     func()
	// new running done error timeout
	Status         string `json:"status"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
}

func createTask(name string, timeoutSeconds int, runner func()) task {
	return task{
		Name:           name,
		StartEpoch:     0,
		EndEpoch:       0,
		Runner:         runner,
		Status:         "new",
		TimeoutSeconds: timeoutSeconds,
	}
}

type TaskQueue struct {
	Queue      *list.List
	queueLock  sync.RWMutex
	Done       []string
	Errors     []string
	History    []TaskQueueHistory
	Times      int
	StartEpoch int64
	EndEpoch   int64
	Todo       int
	Running    *task
	status     string
	signal     chan string
}

type TaskQueueHistory struct {
	SerialId   int
	Name       string
	Result     string
	StartEpoch int64
	EndEpoch   int64
}

type TaskQueueInfo struct {
	Length     int
	Done       []string
	Errors     []string
	StartEpoch int64
	EndEpoch   int64
	Running    string
	Times      int
}

func (thisSelf *task) run() string {
	thisSelf.Status = "running"
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				thisSelf.Status = "error"
				done <- false
			}
		}()
		thisSelf.Runner()
		done <- true
	}()
	select {
	case res := <-done:
		if res {
			thisSelf.Status = "done"
		} else {
			thisSelf.Status = "error"
		}
		break
	case <-time.After(time.Duration(thisSelf.TimeoutSeconds) * time.Second):
		thisSelf.Status = "timeout"
		break
	}
	return thisSelf.Status
}

// 创建任务队列框架(异步, 可监测, 完整运行记录)
func CreateTaskQueue() TaskQueue {
	return TaskQueue{
		Queue:      list.New(),
		Done:       []string{},
		Errors:     []string{},
		Todo:       0,
		Times:      0,
		StartEpoch: 0,
		EndEpoch:   0,
		Running:    nil,
		status:     "new",
		History:    []TaskQueueHistory{},
		signal:     make(chan string),
	}
}

func (thisSelf *TaskQueue) AddTask(name string, timeoutSeconds int, runner func()) {
	thisSelf.queueLock.Lock()
	defer thisSelf.queueLock.Unlock()
	thisSelf.Queue.PushBack(createTask(name, timeoutSeconds, runner))
}

// 执行一次任务队列, 异步
func (thisSelf *TaskQueue) Start() (error, chan string) {
	if thisSelf.status != "new" {
		return errors.New("队列正在运行中"), nil
	}
	done := make(chan string)
	thisSelf.status = "running"
	thisSelf.Done = []string{}
	thisSelf.Errors = []string{}
	thisSelf.Todo = thisSelf.Queue.Len()
	thisSelf.Times += 1
	thisSelf.StartEpoch = TimeEpoch()
	thisSelf.EndEpoch = 0
	go func() {
		for e := thisSelf.Queue.Front(); e != nil; e = e.Next() {
			thisSelf.runner(e.Value.(task))
		receive:
			select {
			case sig := <-thisSelf.signal:
				switch sig {
				case "pause":
				pause:
					select {
					case sig := <-thisSelf.signal:
						switch sig {
						case "continue":
							goto continueTask
							break
						default:
							goto pause
						}
					}
					break
				case "continue":
					goto continueTask
					break
				case "stop":
					panic(errors.New("force stop"))
					break
				default:
					goto receive
					break
				}
			}
		continueTask:
			thisSelf.signal <- "continue"
		}
		thisSelf.EndEpoch = TimeEpoch()
		done <- "done"
		thisSelf.status = "new"
	}()
	return nil, done
}

func (thisSelf *TaskQueue) Pause() {
	thisSelf.signal <- "pause"
}

func (thisSelf *TaskQueue) Continue() {
	thisSelf.signal <- "continue"
}

func (thisSelf *TaskQueue) Stop() {
	thisSelf.signal <- "stop"
}

func (thisSelf *TaskQueue) Status() TaskQueueInfo {

	return TaskQueueInfo{
		Length:     thisSelf.Queue.Len(),
		Done:       thisSelf.Done,
		Errors:     thisSelf.Errors,
		StartEpoch: thisSelf.StartEpoch,
		EndEpoch:   thisSelf.EndEpoch,
		Running:    thisSelf.Running.Name,
		Times:      thisSelf.Times,
	}

}

func (thisSelf *TaskQueue) runner(t task) {
	thisSelf.Running = &t
	thisSelf.Todo -= 1
	history := TaskQueueHistory{
		SerialId: thisSelf.Times,
		Name:     t.Name,

		StartEpoch: TimeEpoch(),
	}
	switch t.run() {
	case "error":
		thisSelf.Errors = append(thisSelf.Errors, t.Name)
		break
	default:
		break
	}
	thisSelf.Done = append(thisSelf.Done, t.Name)
	history.EndEpoch = TimeEpoch()
	history.Result = t.Status
	thisSelf.History = append(thisSelf.History, history)
}
