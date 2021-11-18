package middleware

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

type PipelineStatus struct {
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
	Next []LogicLine
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

type PipelineContext struct {
}

func (c *PipelineContext) Put() {

}

func (c *PipelineContext) Get() {

}

func (l *LogicLine) run(c PipelineContext) PipelineContext {
	if l.Type == LINE {
		ctx := l.Node.Runner(c)
		if l.Next == nil || len(l.Next) <= 0 {
			return ctx
		}
		for i := 0; i < len(l.Next); i++ {
			ctx = l.Next[i].run(ctx)
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
