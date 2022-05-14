package core

import (
	"errors"
	docker "github.com/fsouza/go-dockerclient"
	"strconv"
	"strings"
)

// Step is a single step in a pipeline.
type Step struct {
	Successful bool
	Completed  bool
	Name       string
	Image      string
	Commands   []string
	Errors     []string
	Volume     string
}

// NewStep creates a new step.
func NewStep(name, image string, commands []string) *Step {
	return &Step{
		Successful: false,
		Completed:  false,
		Name:       name,
		Image:      image,
		Commands:   commands,
		Errors:     []string{},
	}
}

// Run executes the step.
func (s *Step) Run(client *docker.Client) error {
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: s.Name,
		HostConfig: &docker.HostConfig{
			Binds: []string{s.Volume + ":/app"},
		},
		Config: &docker.Config{
			Image: s.Image,
			Cmd:   []string{"/bin/sh", "-c", strings.Join(s.Commands, " ")},
			Labels: map[string]string{
				"fastci": "true",
			},
			WorkingDir: "/app",

			Env: []string{
				"FASTCI_STEP_NAME=" + s.Name,
				"FASTCI_STEP_IMAGE=" + s.Image,
			},
		},
	})
	if err != nil {
		s.Successful = false
		s.Completed = true
		s.Errors = append(s.Errors, err.Error())
		return err
	}
	err = client.StartContainer(container.ID, nil)
	if err != nil {
		s.Successful = false
		s.Completed = true
		s.Errors = append(s.Errors, err.Error())
		return err
	}
	status, err := client.WaitContainer(container.ID)
	if err != nil {
		s.Successful = false
		s.Completed = true
		s.Errors = append(s.Errors, err.Error())
		return err
	}
	if status != 0 {
		s.Successful = false
		s.Completed = true
		s.Errors = append(s.Errors, "exit status "+strconv.Itoa(status))
		return errors.New("exit status " + strconv.Itoa(status))
	}

	s.Successful = true
	s.Completed = true
	return nil
}

// String returns the string representation of the step.
func (s *Step) String() string {
	return s.Name
}
