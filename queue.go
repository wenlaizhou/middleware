package middleware

import (
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
	Queue              []task
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
	SerialId   int    `json:"serialId"`
	Span       int    `json:"span"`
	Name       string `json:"name"`
	Result     string `json:"result"`
	StartEpoch int64  `json:"startEpoch"`
	EndEpoch   int64  `json:"endEpoch"`
}

type TaskQueueInfo struct {
	Length     int        `json:"length"`
	Tasks      []TaskInfo `json:"tasks"`
	Done       []string   `json:"done"`
	Errors     []string   `json:"errors"`
	StartEpoch int64      `json:"startEpoch"`
	EndEpoch   int64      `json:"endEpoch"`
	Running    string     `json:"running"`
	Times      int        `json:"times"`
	Status     string     `json:"status"`
}

type TaskInfo struct {
	Name    string `json:"name"`
	Timeout int    `json:"timeout"`
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
		Queue:              []task{},
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
	q.Queue = append(q.Queue, createTask(name, timeoutSeconds, runner))
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
	q.Todo = len(q.Queue)
	q.Times += 1
	q.StartEpoch = TimeEpoch()
	q.EndEpoch = 0
	q.TaskQueueHistories[q.StartEpoch] = []TaskQueueHistory{}
	go func() {
		spanId := 0
		for _, task := range q.Queue {
			q.runner(task, spanId)
			spanId++
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
	tasks := []TaskInfo{}
	for _, task := range q.Queue {
		tasks = append(tasks, TaskInfo{
			Name:    task.Name,
			Timeout: task.TimeoutSeconds,
		})
	}
	return TaskQueueInfo{
		Length:     len(q.Queue),
		Tasks:      tasks,
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

func (q *TaskQueue) runner(t task, span int) {
	q.Running = &t
	q.Todo -= 1
	history := TaskQueueHistory{
		SerialId:   q.Times,
		Name:       t.Name,
		Span:       span,
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
