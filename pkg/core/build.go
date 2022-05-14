package core

import docker "github.com/fsouza/go-dockerclient"

type BuildState int

const (
	READY BuildState = iota
	RUNNING
	COMPLETED
	FAILED
)

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

// Build is a single build in the build queue.
type Build struct {
	Name     string
	Pipeline *Pipeline
	State    BuildState
	Volume   string
}

// NewBuild creates a new build.
func NewBuild(name string, pipeline *Pipeline) *Build {
	return &Build{
		Name:     name,
		Pipeline: pipeline,
		State:    READY,
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
func (build *Build) NextStep() (*Step, bool) {
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
	case RUNNING:
		step, done := build.NextStep()
		if done {
			build.State = COMPLETED
			return
		}
		step.Volume = build.Volume
		err := step.Run(client)
		if err != nil {
			build.State = FAILED
			return
		}
		build.Run(client)
	}
	return
}
