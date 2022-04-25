package middleware

import (
	"sync"
	"sync/atomic"
)

var pipelineLogger = GetLogger("pipeline")

// 流程
type PipeLine struct {

	// 名称
	Name string

	// 根逻辑节点
	Root *Logic

	// 总逻辑数(预估)
	Total int
}

type PipelineManager struct {
	locker sync.RWMutex

	Results []*PipelineResult

	// pipeline
}

func (p *PipelineManager) Start(pipe PipeLine, input interface{}) {
	// p.Pipelines[pipe.Name] = map[string][]LogicResult{}
	p.locker.Lock()
	defer p.locker.Unlock()
	trace := Guid()
	result := &PipelineResult{
		Name:   pipe.Name,
		Trace:  trace,
		Start:  TimeEpoch(),
		Status: "start",
	}
	result.Current(0, pipe.Root.Name)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				pipelineLogger.ErrorF("%v", err)
			}
		}()
		var span int32 = 0
		res, output, next := StartLogic(pipe.Name, trace, span, pipe.Root, input)
		span++
		result.AddResult(res)
		if next == nil {
			result.Status = "done"
			return
		}
		NextLogics(pipe.Name, trace, &span, next, output, result)
		result.Status = "done"
	}()
	p.Results = append(p.Results, result)
}

func NextLogics(pipeline string, trace string, span *int32, logics []*Logic, input interface{}, result *PipelineResult) {
	if logics == nil {
		return
	}
	for _, n := range logics {
		atomic.AddInt32(span, 1)
		result.Current(*span, n.Name)
		res, output, next := StartLogic(pipeline, trace, *span, n, input)
		result.AddResult(res)
		if next != nil {
			NextLogics(pipeline, trace, span, logics, output, result)
		}
		continue
	}
	return
}

func StartLogic(pipeline string, trace string, span int32, logic *Logic, input interface{}) (LogicResult, interface{}, []*Logic) {
	result := LogicResult{
		Pipeline: pipeline,
		Name:     logic.Name,
		Trace:    trace,
		Span:     int(span),
		Start:    TimeEpoch(),
	}
	parsedInput := logic.Before(input)

	if !logic.Condition(parsedInput) {
		output := logic.After(parsedInput)
		next := logic.Selector(logic, output)
		result.End = TimeEpoch()
		result.Result = "condition not passed"
		return result, output, next
	}

	output := logic.After(logic.Runner(parsedInput))

	next := logic.Selector(logic, output)

	result.End = TimeEpoch()
	result.Result = "success"

	return result, output, next
}

// 逻辑节点
type Logic struct {

	// 名称
	Name string

	// 输入过滤器
	Before func(input interface{}) interface{}

	// 条件判断器
	Condition func(input interface{}) bool

	// 执行器
	Runner func(interface{}) interface{}

	// 输出过滤器, 返回结果类型
	After func(output interface{}) interface{}

	// 返回逻辑节点
	Selector func(logic *Logic, output interface{}) []*Logic

	// 子节点
	Children []*Logic
}

func CreatePipeline(name string, root *Logic) *PipeLine {
	return &PipeLine{
		Name:  name,
		Root:  root,
		Total: 0,
	}
}

func PipelineAddLogic(p *PipeLine, logic []*Logic) {
	p.Root.Children = logic
}

// func PipelineAddLogicSpec(p *PipeLine, deep int, number int, logic []*Logic) error {
//	if deep == 0 {
//		if number >= len(p.Root.Children) {
//			return errors.New("length over flow")
//		}
//		p.Root.Children[number].Children = append(p.Root.Children[number].Children, logic...)
//		return nil
//	}
//
//	for i := 0; i < deep; i++ {
//
//	}
// }

func CreateLogic(name string, before func(input interface{}) interface{},
	condition func(input interface{}) bool, runner func(interface{}) interface{},
	after func(output interface{}) interface{}, selector func(logic *Logic, output interface{}) []*Logic) *Logic {
	return &Logic{
		Name:      name,
		Before:    before,
		Condition: condition,
		Runner:    runner,
		After:     after,
		Selector:  selector,
		Children:  nil,
	}
}

func AddLogicChild(logic *Logic, child []*Logic) {
	logic.Children = append(logic.Children, child...)
}

func AddLogicChildNext(logic *Logic, deep int, next int, child []*Logic) error {
	// todo
	AddLogicChild(logic.Children[next], child)
	return nil
}

type PipelineResult struct {

	// pipeline name
	Name string

	// start time
	Start int64

	// status
	Status string

	// 结束时间
	End int64

	// trace id
	Trace string

	// 当前span
	CurrentSpan int32

	// 当前逻辑
	CurrentLogic string

	// 锁数据
	locker sync.RWMutex

	// 执行结果
	logicResults []LogicResult
}

func (p *PipelineResult) AddResult(l LogicResult) {
	p.locker.Lock()
	defer p.locker.Unlock()
	p.logicResults = append(p.logicResults, l)
}

func (p *PipelineResult) Results() []LogicResult {
	return p.logicResults
}

func (p *PipelineResult) Current(span int32, logic string) {
	p.locker.Lock()
	defer p.locker.Unlock()
	p.CurrentSpan = span
	p.CurrentLogic = logic
}

type LogicResult struct {

	// 流程名称
	Pipeline string `json:"pipeline"`

	// 逻辑节点名称
	Name string `json:"name"`

	// trace id
	Trace string `json:"trace"`

	// span id
	Span int `json:"span"`

	// 开始时间
	Start int64 `json:"start"`

	// 结束时间
	End int64 `json:"end"`

	// 输入
	// Input interface{}

	// 输出
	// Output interface{}

	// 执行结果
	Result string `json:"result"`
}
