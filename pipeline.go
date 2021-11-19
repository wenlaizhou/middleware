package middleware

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
	Pipelines map[string]PipeLine
}

func (p *PipelineManager) Start(pipe PipeLine, input interface{}) {

}

func StartLogic(logic *Logic) (LogicResult, *Logic) {
	return LogicResult{}, nil
}

// 逻辑节点
type Logic struct {

	// 输入过滤器
	Before func(input interface{}) interface{}

	// 条件判断器
	Condition func(input interface{}) bool

	// 执行器
	Runner func(interface{}) interface{}

	// 输出过滤器, 返回结果类型
	After func(output interface{}) interface{}

	// 返回逻辑节点
	Selector func(output interface{}) *Logic

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
	Result int `json:"result"`
}
