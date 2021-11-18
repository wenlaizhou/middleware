package middleware

import ()

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
	Root LogicLine

	// 总逻辑数
	Total int
}

func StartPipeline(pipeline Pipeline) PipelineStatus {

	result := PipelineStatus{
		Start:   TimeEpoch(),
		Current: pipeline.Name,
	}

	return result
}

func runLogic(line LogicLine, ctx PipelineContext) []LogicLine {
	return nil
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
	Children []LogicLine
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
}
