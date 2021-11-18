package middleware

import (
	"errors"
)

const (
	LINE   = 0
	SWITCH = 1
	LOOP   = 2
	IF     = 3
	ASYNC  = 4
)

// pipeline
type Pipeline struct {

	// pipeline 名称
	Name string

	// 根节点
	Root *LogicLine

	// 总逻辑数
	Total int
}

func (p *Pipeline) AddLogic(logic LogicLine) {
	if p.Root == nil {
		p.Root = &logic
		return
	}
	curr := p.Root
	next := p.Root
	for ; next.Children != nil; next = next.Children {
		curr = next
	}
	curr.Children = &logic
}

func StartPipeline(pipeline Pipeline, input interface{}) PipelineContext {

	ctx := &PipelineContext{
		Input: input,
	}

	go runLogic(*pipeline.Root, ctx)

	return *ctx
}

func runLogic(line LogicLine, ctx *PipelineContext) PipelineContext {
	result := LogicResult{
		Name:  line.Name,
		Type:  line.Type,
		Node:  line.Node.Name,
		Start: TimeEpoch(),
		Input: ctx.Input,
		Error: nil,
	}
	defer func() {
		if err := recover(); err != nil {
			result.Error = errors.New("panic")
			ctx.Done[line.Name] = result
		}
	}()
	ctx.Current = line.Name
	input := line.InputFilter(ctx.Input)
	output := line.Node.Runner(input)
	output = line.OutputFilter(output)
	result.Output = output
	result.End = TimeEpoch()
	ctx.Done[line.Name] = result
	ctx.Input = output
	if line.Children != nil {
		return runLogic(*line.Children, ctx)
	}
	return *ctx
}

type PipelineStatus struct {
	Start   int64         `json:"start"`
	End     int64         `json:"end"`
	Current string        `json:"current"`
	Done    []LogicResult `json:"done"`
	Total   int           `json:"total"`
}

// 逻辑线
type LogicLine struct {

	// 类型
	Type int

	// 输入过滤
	InputFilter func(interface{}) interface{}

	// 输出过滤
	OutputFilter func(interface{}) interface{}

	// 名称
	Name string

	// 节点
	Node PipelineNode

	// 配置
	Config map[string]string

	// 逻辑数量
	Len int

	// 子逻辑
	Children *LogicLine
}

// 节点
type PipelineNode struct {

	// 节点名称
	Name string

	// 节点配置
	Config map[string]string

	// 节点处理器
	Runner func(interface{}) interface{}
}

// 上下文对象
type PipelineContext struct {

	// 输入参数
	Input interface{}

	// 当前处理状态
	Current string

	// 已完成的流程上下文
	Done map[string]LogicResult
}

type LogicResult struct {
	Name   string      `json:"name"`
	Type   int         `json:"type"`
	Node   string      `json:"node"`
	Start  int64       `json:"start"`
	End    int64       `json:"end"`
	Input  interface{} `json:"input"`
	Output interface{} `json:"output"`
	Error  error       `json:"error"`
}
