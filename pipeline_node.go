package scribe

// A PipelineNode
type PipelineNode interface {
	Requires() []Argument
}
