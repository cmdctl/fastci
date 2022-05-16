package test

import (
	docker "github.com/fsouza/go-dockerclient"
	"testing"
)

// MockWriter is a mock implementation of io.Writer
type MockWriter struct {
	Content []byte
}

// Write is a mock implementation of io.Writer.Write
func (m *MockWriter) Write(p []byte) (n int, err error) {
	m.Content = append(m.Content, p...)
	return len(p), nil
}

// Cleanup cleans all containers and images from the test environment.
func Cleanup(t *testing.T) {
	dockerClient, _ := docker.NewClient("unix:///var/run/docker.sock")
	containers, err := dockerClient.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"label": {"fastci"},
		},
	})
	if err != nil {
		t.Errorf("Error listing containers: %s", err)
	}
	for _, container := range containers {
		err := dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID: container.ID,
		})
		if err != nil {
			t.Errorf("Error removing container: %s", err)
		}
	}
	volumes, err := dockerClient.ListVolumes(docker.ListVolumesOptions{
		Filters: map[string][]string{
			"label": {"fastci"},
		},
	})
	if err != nil {
		t.Errorf("Error listing volumes: %s", err)
	}
	for _, volume := range volumes {
		err := dockerClient.RemoveVolumeWithOptions(docker.RemoveVolumeOptions{
			Name: volume.Name,
		})
		if err != nil {
			t.Errorf("Error removing volume: %s", err)
		}
	}
}
