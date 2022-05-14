package core

import (
	"github.com/cmdctl/fastci/test"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
	"testing"
)

// SuccessCase runs a test case with a success expectation.
func SuccessCase(t *testing.T) {
	client, _ := docker.NewClient("unix:///var/run/docker.sock")
	step1 := NewStep("step1", "ubuntu", []string{"echo", "hello"})
	step2 := NewStep("step2", "ubuntu", []string{"echo", "world"})
	pipeline := NewPipeline(step1, step2)
	build := NewBuild("test", pipeline)
	build.Run(client)
	assert.Equal(t, build.State.String(), COMPLETED.String())
	assert.Equal(t, true, build.Pipeline.Steps[0].Successful)
	assert.Equal(t, true, build.Pipeline.Steps[0].Completed)
	assert.Equal(t, 0, len(build.Pipeline.Steps[0].Errors))
	assert.Equal(t, true, build.Pipeline.Steps[1].Successful)
	assert.Equal(t, true, build.Pipeline.Steps[1].Completed)
	assert.Equal(t, 0, len(build.Pipeline.Steps[1].Errors))
}

// FailureCase runs a test case with a failure expectation.
func FailureCase(t *testing.T) {
	client, _ := docker.NewClient("unix:///var/run/docker.sock")
	step1 := NewStep("step1", "ubuntu", []string{"echo", "hello"})
	step2 := NewStep("step2", "ubuntu", []string{"exit", "1"})
	pipeline := NewPipeline(step1, step2)
	build := NewBuild("test", pipeline)
	build.Run(client)
	assert.Equal(t, FAILED.String(), build.State.String())
	assert.Equal(t, true, build.Pipeline.Steps[0].Successful)
	assert.Equal(t, true, build.Pipeline.Steps[0].Completed)
	assert.Equal(t, 0, len(build.Pipeline.Steps[0].Errors))
	assert.Equal(t, false, build.Pipeline.Steps[1].Successful)
	assert.Equal(t, true, build.Pipeline.Steps[1].Completed)
	assert.Equal(t, 1, len(build.Pipeline.Steps[1].Errors))
}

func VolumeMountCase(t *testing.T) {
	client, _ := docker.NewClient("unix:///var/run/docker.sock")
	step1 := NewStep("step1", "ubuntu", []string{"touch", "text.txt"})
	step2 := NewStep("step2", "ubuntu", []string{"cat", "text.txt"})
	pipeline := NewPipeline(step1, step2)
	build := NewBuild("test", pipeline)
	build.Run(client)
	assert.Equal(t, build.State.String(), COMPLETED.String())
	assert.Equal(t, true, build.Pipeline.Steps[0].Successful)
	assert.Equal(t, true, build.Pipeline.Steps[0].Completed)
	assert.Equal(t, 0, len(build.Pipeline.Steps[0].Errors))
	assert.Equal(t, true, build.Pipeline.Steps[1].Successful)
	assert.Equal(t, true, build.Pipeline.Steps[1].Completed)
	assert.Equal(t, 0, len(build.Pipeline.Steps[1].Errors))
}

var testCases = map[string]func(t *testing.T){
	"SuccessCase":     SuccessCase,
	"FailureCase":     FailureCase,
	"VolumeMountCase": VolumeMountCase,
}

func TestBuild_Run(t *testing.T) {
	for name, testCase := range testCases {
		t.Run(name, testCase)
		test.Cleanup(t)
	}
}
