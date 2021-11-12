package middleware

import (
	"container/list"
	"sync"
)

type task struct {
	Name       string `json:"name"`
	StartEpoch int64  `json:"startEpoch"`
	EndEpoch   int64  `json:"endEpoch"`
	Runner     func()
	// new running done error
	Status string `json:"status"`
	signal chan bool
}

func createTask(name string, runner func()) task {
	return task{
		Name:       name,
		StartEpoch: 0,
		EndEpoch:   0,
		Runner:     runner,
		Status:     "new",
		signal:     make(chan bool),
	}
}

func (thisSelf *task) run() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				thisSelf.signal <- false
				thisSelf.Status = "error"
			}
		}()
		thisSelf.Runner()
		thisSelf.Status = "done"
		thisSelf.signal <- true
	}()
	thisSelf.Status = "running"
}

type TaskQueue struct {
	Queue     *list.List
	queueLock sync.RWMutex
	Done      []string
	Errors    []string
	Todo      int
	Running   *task
	status    string
}

func CreateTaskQueue() TaskQueue {
	return TaskQueue{
		Queue:   list.New(),
		Done:    []string{},
		Errors:  []string{},
		Todo:    0,
		Running: nil,
		status:  "new",
	}
}

func (thisSelf *TaskQueue) AddTask(name string, runner func()) {
	thisSelf.queueLock.Lock()
	defer thisSelf.queueLock.Unlock()
	thisSelf.Queue.PushBack(createTask(name, runner))
}

func (thisSelf *TaskQueue) Start() {
	thisSelf.Done = []string{}
	thisSelf.Errors = []string{}
	thisSelf.Todo = thisSelf.Queue.Len()
	go func() {
		for e := thisSelf.Queue.Front(); e != nil; e = e.Next() {
			thisSelf.runner(e.Value.(task))
		}
	}()
}

func (thisSelf *TaskQueue) runner(t task) {
	t.run()
	thisSelf.Running = &t
	thisSelf.Todo -= 1
	select {
	case res := <-t.signal:
		if !res {
			thisSelf.Errors = append(thisSelf.Errors, t.Name)
		}
		thisSelf.Done = append(thisSelf.Done, t.Name)
	}
}

func (thisSelf *TaskQueue) Status() {

}
