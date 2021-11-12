package middleware

type task struct {
	Name       string `json:"name"`
	StartEpoch int64  `json:"startEpoch"`
	EndEpoch   int64  `json:"endEpoch"`
	Runner     func()
	// new running done error
	Status string `json:"status"`
}

func CreateTask(name string, runner func()) task {
	return task{
		Name:       name,
		StartEpoch: 0,
		EndEpoch:   0,
		Runner:     runner,
		Status:     "new",
	}
}

func (thisSelf *task) run() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				thisSelf.Status = "error"
			}
		}()
		thisSelf.Runner()
		thisSelf.Status = "done"
	}()
	thisSelf.Status = "running"
}

type TaskQueue struct {
	Queue   []task `json:"queue"`
	Done    []task `json:"done"`
	Todo    []task `json:"todo"`
	Running task   `json:"running"`
}
