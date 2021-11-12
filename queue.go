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
	Queue              *list.List
	queueLock          sync.RWMutex
	Done               []string
	Errors             []string
	TaskQueueHistories map[int64][]TaskQueueHistory
	Times              int
	StartEpoch         int64
	EndEpoch           int64
	Todo               int
	Running            *task
	status             string
	signal             chan string
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
	Status     string
}

func (t *task) run() string {
	t.Status = "running"
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Status = "error"
				done <- false
			}
		}()
		t.Runner()
		done <- true
	}()
	select {
	case res := <-done:
		if res {
			t.Status = "done"
		} else {
			t.Status = "error"
		}
		break
	case <-time.After(time.Duration(t.TimeoutSeconds) * time.Second):
		t.Status = "timeout"
		break
	}
	return t.Status
}

// 创建任务队列框架(异步, 可监测, 完整运行记录)
func CreateTaskQueue() TaskQueue {
	return TaskQueue{
		Queue:              list.New(),
		Done:               []string{},
		Errors:             []string{},
		Todo:               0,
		Times:              0,
		StartEpoch:         0,
		EndEpoch:           0,
		Running:            nil,
		status:             "new",
		TaskQueueHistories: map[int64][]TaskQueueHistory{},
		signal:             make(chan string),
	}
}

func (q *TaskQueue) AddTask(name string, timeoutSeconds int, runner func()) {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	q.Queue.PushBack(createTask(name, timeoutSeconds, runner))
}

// 执行一次任务队列, 异步
func (q *TaskQueue) Start() (error, chan string) {
	if q.status != "new" && q.status != "done" {
		return errors.New("队列正在运行中"), nil
	}
	done := make(chan string, 1)
	q.status = "running"
	q.Done = []string{}
	q.Errors = []string{}
	q.Todo = q.Queue.Len()
	q.Times += 1
	q.StartEpoch = TimeEpoch()
	q.EndEpoch = 0
	q.TaskQueueHistories[q.StartEpoch] = []TaskQueueHistory{}
	go func() {
		for e := q.Queue.Front(); e != nil; e = e.Next() {
			q.runner(e.Value.(task))
			select {
			case sig := <-q.signal:
				switch sig {
				case "pause":
					q.status = "pause"
				pause:
					select {
					case sig := <-q.signal:
						switch sig {
						case "continue":
							q.status = "running"
							break
						default:
							goto pause
						}
						break
					}
				case "continue":
					break
				case "stop":
					q.status = "done"
					panic(errors.New("force stop"))
				default:
					break
				}
			default:
				break
			}
		}
		q.EndEpoch = TimeEpoch()
		done <- "done"
		q.Running = nil
		q.status = "done"
	}()
	return nil, done
}

func (q *TaskQueue) Pause() {
	if q.status != "running" {
		return
	}
	q.signal <- "pause"
}

func (q *TaskQueue) Continue() {
	if q.status != "pause" {
		return
	}
	q.signal <- "continue"
}

func (q *TaskQueue) Stop() {
	q.signal <- "stop"
}

func (q *TaskQueue) Status() TaskQueueInfo {
	running := ""
	if q.Running != nil {
		running = q.Running.Name
	}
	return TaskQueueInfo{
		Length:     q.Queue.Len(),
		Done:       q.Done,
		Errors:     q.Errors,
		StartEpoch: q.StartEpoch,
		EndEpoch:   q.EndEpoch,
		Running:    running,
		Times:      q.Times,
		Status:     q.status,
	}

}

func (q *TaskQueue) History() map[int64][]TaskQueueHistory {
	return q.TaskQueueHistories
}

func (q *TaskQueue) runner(t task) {
	q.Running = &t
	q.Todo -= 1
	history := TaskQueueHistory{
		SerialId: q.Times,
		Name:     t.Name,

		StartEpoch: TimeEpoch(),
	}
	switch t.run() {
	case "error":
		q.Errors = append(q.Errors, t.Name)
		break
	default:
		break
	}
	q.Done = append(q.Done, t.Name)
	history.EndEpoch = TimeEpoch()
	history.Result = t.Status
	q.TaskQueueHistories[q.StartEpoch] =
		append(q.TaskQueueHistories[q.StartEpoch], history)
}
