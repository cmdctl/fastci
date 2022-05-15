package core

import (
	"github.com/cmdctl/fastci/test"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
	"testing"
)

// CaseSuccess runs a test case with a success expectation.
func CaseSuccess(t *testing.T) {
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

// CaseFailure runs a test case with a failure expectation.
func CaseFailure(t *testing.T) {
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

// CaseVolumeMount runs a test case with a volume mount.
func CaseVolumeMount(t *testing.T) {
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

// CaseLogCollection runs a test case with mock std out and err.
func CaseLogCollection(t *testing.T) {
	stdOut := &test.MockWriter{Content: []byte{}}
	stdErr := &test.MockWriter{Content: []byte{}}

	client, _ := docker.NewClient("unix:///var/run/docker.sock")
	step1 := NewStep("logCollection1", "ubuntu", []string{"echo", "hello"})
	step2 := NewStep("logCollection2", "ubuntu", []string{"echo", "world"})
	pipeline := NewPipeline(step1, step2)
	build := NewBuildWithOutputStreams("testLogCollection", pipeline, stdOut, stdErr)
	build.Run(client)
	assert.Equal(t, COMPLETED.String(), build.State.String())
	assert.Equal(t, true, build.Pipeline.Steps[0].Successful)
	assert.Equal(t, true, build.Pipeline.Steps[0].Completed)
	assert.Equal(t, 0, len(build.Pipeline.Steps[0].Errors))
	assert.Equal(t, true, build.Pipeline.Steps[1].Successful)
	assert.Equal(t, true, build.Pipeline.Steps[1].Completed)
	assert.Equal(t, 0, len(build.Pipeline.Steps[1].Errors))
	assert.Equal(t, "hello\nworld\n", string(stdOut.Content))
	assert.Equal(t, "", string(stdErr.Content))
}

// CasePullingImages runs a test case with pulling busybox image.
func CasePullingImages(t *testing.T) {
	client, _ := docker.NewClient("unix:///var/run/docker.sock")
	step1 := NewStep("pullingImagesStep1", "busybox", []string{"echo", "hello"})
	step2 := NewStep("pullingImagesStep2", "busybox", []string{"echo", "world"})
	pipeline := NewPipeline(step1, step2)
	build := NewBuild("testPullingImages", pipeline)
	build.Run(client)
	assert.Equal(t, COMPLETED.String(), build.State.String())
	assert.Equal(t, true, build.Pipeline.Steps[0].Successful)
	assert.Equal(t, true, build.Pipeline.Steps[0].Completed)
	assert.Equal(t, 0, len(build.Pipeline.Steps[0].Errors))
	assert.Equal(t, true, build.Pipeline.Steps[1].Successful)
	assert.Equal(t, true, build.Pipeline.Steps[1].Completed)
	assert.Equal(t, 0, len(build.Pipeline.Steps[1].Errors))
}

// TestBuild_Run runs test cases with a build.
func TestBuild_Run(t *testing.T) {
	var testCases = map[string]func(t *testing.T){
		"CaseSuccess":       CaseSuccess,
		"CaseFailure":       CaseFailure,
		"CaseVolumeMount":   CaseVolumeMount,
		"CaseLogCollection": CaseLogCollection,
		"CasePullingImages": CasePullingImages,
	}
	for name, testCase := range testCases {
		t.Run(name, testCase)
		test.Cleanup(t)
	}
}
