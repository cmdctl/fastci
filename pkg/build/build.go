package build

import (
	docker "github.com/fsouza/go-dockerclient"
	"github.com/lensesio/tableprinter"
	"io"
)

// State BuildState represents the state of a build.
type State int

const (
	READY State = iota
	RUNNING
	COMPLETED
	FAILED
)

// Build is a single build in the build queue.
type Build struct {
	Name         string    `json:"name" yaml:"name"`
	Pipeline     *Pipeline `json:"pipeline" yaml:"pipeline"`
	State        State     `json:"state" yaml:"state"`
	Volume       string    `json:"volume" yaml:"volume"`
	OutputStream io.Writer `json:"-" yaml:"-"`
	ErrorStream  io.Writer `json:"-" yaml:"-"`
}

func (b *Build) PrintResult(writer io.Writer) error {
	_, err := writer.Write([]byte(b.Name + ": " + b.State.String() + "\n"))
	if err != nil {
		return err
	}
	printer := tableprinter.New(writer)
	printer.BorderTop, printer.BorderBottom, printer.BorderLeft, printer.BorderRight = true, true, true, true
	printer.ColumnSeparator = "│"
	printer.RowSeparator = "─"
	printer.Print(b.Pipeline.Steps)
	return nil
}

// NewBuild creates a new build.
func NewBuild(name string, pipeline *Pipeline) *Build {
	return &Build{
		Name:     name,
		Pipeline: pipeline,
		State:    READY,
	}
}

// NewBuildWithOutputStreams creates a new build with the given streams for standard and error outputs.
func NewBuildWithOutputStreams(name string, pipeline *Pipeline, outputStream, errorStream io.Writer) *Build {
	return &Build{
		Name:         name,
		Pipeline:     pipeline,
		State:        READY,
		OutputStream: outputStream,
		ErrorStream:  errorStream,
	}
}

// String returns the string representation of a BuildState.
func (b State) String() string {
	switch b {
	case READY:
		return "READY"
	case RUNNING:
		return "RUNNING"
	case COMPLETED:
		return "COMPLETED"
	case FAILED:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// AllStepsSuccessful returns true if all steps in the pipeline are successful.
func (b *Build) AllStepsSuccessful() bool {
	for _, step := range b.Pipeline.Steps {
		if !step.Successful {
			return false
		}
	}
	return true
}

// NextStep returns the next step to be executed.
func (b *Build) NextStep() (step *Step, done bool) {
	if b.AllStepsSuccessful() {
		return nil, true
	}
	for _, step := range b.Pipeline.Steps {
		if !step.Completed {
			return step, false
		}
	}
	return nil, true
}

// Run returns the next step to be executed.
func (b *Build) Run(client *docker.Client) {
	switch b.State {
	case READY:
		prepare(client, b)
	case RUNNING:
		run(client, b)
	}
}

// helper function to run a build
func run(client *docker.Client, build *Build) {
	step, done := build.NextStep()
	if done {
		build.State = COMPLETED
		return
	}
	if build.OutputStream != nil {
		step.OutputStream = build.OutputStream
	}
	if build.ErrorStream != nil {
		step.ErrorStream = build.ErrorStream
	}
	step.Volume = build.Volume
	err := step.Run(client)
	if err != nil {
		build.State = FAILED
		return
	}
	build.Run(client)
}

// helper function to prepare a build
func prepare(client *docker.Client, build *Build) {
	_, done := build.NextStep()
	if done {
		build.State = COMPLETED
		return
	}
	volume, err := client.CreateVolume(docker.CreateVolumeOptions{
		Labels: map[string]string{
			"fastci": build.Name,
		},
	})
	if err != nil {
		return
	}
	build.Volume = volume.Name
	build.State = RUNNING
	build.Run(client)
}
