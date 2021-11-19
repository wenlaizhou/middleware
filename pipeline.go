package middleware

import "sync/atomic"

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

	// pipeline
}

func (p *PipelineManager) Start(pipe PipeLine, input interface{}) {
	//p.Pipelines[pipe.Name] = map[string][]LogicResult{}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				pipelineLogger.ErrorF("%v", err)
			}
		}()
		result := []LogicResult{}
		trace := Guid()
		var span int32 = 0
		res, output, next := StartLogic(pipe.Name, trace, span, pipe.Root, input)
		span++
		result = append(result, res)
		if next == nil {
			return
		}
		NextLogics(pipe.Name, trace, &span, next, output, result)
	}()
}

func NextLogics(pipeline string, trace string, span *int32, logics []*Logic, input interface{}, results []LogicResult) {
	if logics == nil {
		return
	}
	for _, n := range logics {
		atomic.AddInt32(span, 1)
		res, output, next := StartLogic(pipeline, trace, *span, n, input)
		results = append(results, res)
		if next != nil {
			NextLogics(pipeline, trace, span, logics, output, results)
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
	//Input interface{}

	// 输出
	//Output interface{}

	// 执行结果
	Result string `json:"result"`
}
