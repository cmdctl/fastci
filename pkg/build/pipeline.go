package build

// Pipeline is a collection of steps.
type Pipeline struct {
	Steps []*Step `json:"steps" yaml:"steps"`
}

// NewPipeline creates a new pipeline.
func NewPipeline(steps ...*Step) *Pipeline {
	return &Pipeline{
		Steps: steps,
	}
}
