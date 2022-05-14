package test

import (
	docker "github.com/fsouza/go-dockerclient"
	"testing"
)

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
