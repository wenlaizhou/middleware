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

	// 根节点
	Root LogicLine

	// 总逻辑数
	Total   int
	Running []PipelineStatus
	History []PipelineStatus
}

func (p *Pipeline) Start() PipelineStatus {

}

type PipelineStatus struct {
}

func (p *PipelineStatus) Status() {

}

// 逻辑线
type LogicLine struct {

	// 类型
	Type int

	// 名称
	Name string

	// 节点
	Node PipelineNode

	// 配置
	Config map[string]string

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
	Runner func(ctx PipelineContext) PipelineContext

	// 节点是否在线
	Health func() bool

	// 节点上线
	Online func() bool
}

// 上下文对象
type PipelineContext struct {

	// 输入, 每过一个流程, Input都会变更
	Input interface{}

	// 配置
	Config map[string]interface{}

	// 已完成的流程上下文
	Done map[string]NodeResult
}

type NodeResult struct {
	Start  int64
	End    int64
	Input  interface{}
	Output interface{}
}

func (c *PipelineContext) Put() {

}

func (c *PipelineContext) Get() {

}

func (l *LogicLine) run(c PipelineContext) PipelineContext {
	if l.Type == LINE {
		ctx := l.Node.Runner(c)
		if l.Children == nil || len(l.Children) <= 0 {
			return ctx
		}
		for i := 0; i < len(l.Children); i++ {
			ctx = l.Children[i].run(ctx)
		}
		return ctx
	}
	if l.Type == ASYNC {
		l.Type = LINE
		go l.run(c)
		return c
	}
	return c
}
