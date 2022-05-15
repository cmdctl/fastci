package core

import (
	docker "github.com/fsouza/go-dockerclient"
	"io"
)

// BuildState represents the state of a build.
type BuildState int

const (
	READY BuildState = iota
	RUNNING
	COMPLETED
	FAILED
)

// Build is a single build in the build queue.
type Build struct {
	Name         string
	Pipeline     *Pipeline
	State        BuildState
	Volume       string
	OutputStream io.Writer
	ErrorStream  io.Writer
}

// NewBuild creates a new build.
func NewBuild(name string, pipeline *Pipeline) *Build {
	return &Build{
		Name:     name,
		Pipeline: pipeline,
		State:    READY,
	}
}

// NewBuildWithLogStreams creates a new build with the given streams for standard and error outputs.
func NewBuildWithLogStreams(name string, pipeline *Pipeline, outputStream, errorStream io.Writer) *Build {
	return &Build{
		Name:         name,
		Pipeline:     pipeline,
		State:        READY,
		OutputStream: outputStream,
		ErrorStream:  errorStream,
	}
}

// String returns the string representation of a BuildState.
func (b BuildState) String() string {
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
func (build *Build) AllStepsSuccessful() bool {
	for _, step := range build.Pipeline.Steps {
		if !step.Successful {
			return false
		}
	}
	return true
}

// NextStep returns the next step to be executed.
func (build *Build) NextStep() (step *Step, done bool) {
	if build.AllStepsSuccessful() {
		return nil, true
	}
	for _, step := range build.Pipeline.Steps {
		if !step.Completed {
			return step, false
		}
	}
	return nil, true
}

// Run returns the next step to be executed.
func (build *Build) Run(client *docker.Client) {
	switch build.State {
	case READY:
		prepare(client, build)
	case RUNNING:
		run(client, build)
	}
	return
}

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
