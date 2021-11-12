package middleware

import (
	"container/list"
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

func (thisSelf *task) run() string {
	thisSelf.Status = "running"
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				thisSelf.Status = "error"
			}
		}()
		thisSelf.Runner()
		done <- true
	}()
	select {
	case <-done:
		thisSelf.Status = "done"
		break
	case <-time.After(time.Duration(thisSelf.TimeoutSeconds) * time.Second):
		thisSelf.Status = "timeout"
		break
	}
	return thisSelf.Status
}

type TaskQueue struct {
	Queue     *list.List
	queueLock sync.RWMutex
	Done      []string
	Errors    []string
	History   []TaskQueueHistory
	Times     int
	Todo      int
	Running   *task
	status    string
}

type TaskQueueHistory struct {
	SerialId   int
	Name       string
	Result     string
	StartEpoch int64
	EndEpoch   int64
}

func CreateTaskQueue() TaskQueue {
	return TaskQueue{
		Queue:   list.New(),
		Done:    []string{},
		Errors:  []string{},
		Todo:    0,
		Times:   0,
		Running: nil,
		status:  "new",
		History: []TaskQueueHistory{},
	}
}

func (thisSelf *TaskQueue) AddTask(name string, timeoutSeconds int, runner func()) {
	thisSelf.queueLock.Lock()
	defer thisSelf.queueLock.Unlock()
	thisSelf.Queue.PushBack(createTask(name, timeoutSeconds, runner))
}

func (thisSelf *TaskQueue) Start() {
	thisSelf.Done = []string{}
	thisSelf.Errors = []string{}
	thisSelf.Todo = thisSelf.Queue.Len()
	thisSelf.Times += 1
	go func() {
		for e := thisSelf.Queue.Front(); e != nil; e = e.Next() {
			thisSelf.runner(e.Value.(task))
		}
	}()
}

func (thisSelf *TaskQueue) runner(t task) {
	thisSelf.Running = &t
	thisSelf.Todo -= 1
	switch t.run() {
	case "error":
		thisSelf.Errors = append(thisSelf.Errors, t.Name)
		break
	default:
		break
	}
	thisSelf.Done = append(thisSelf.Done, t.Name)
}

func (thisSelf *TaskQueue) Status() {

}
