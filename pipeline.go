package middleware

const (
	LOGIC_TYPE_NOMAR = 0
)

type LogicLine struct {
	Type int

	Name string

	Node PipelineNode

	Config map[string]string

	Next []LogicLine
}

type PipelineNode struct {
	Config map[string]string

	Runner func(ctx PipelineContext) PipelineContext
}

type PipelineContext struct {
}

func (c *PipelineContext) Put() {

}

func (c *PipelineContext) Get() {

}

func (l *LogicLine) run(c *PipelineContext) {

}
