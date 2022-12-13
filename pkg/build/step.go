package build

import (
	"errors"
	docker "github.com/fsouza/go-dockerclient"
	"io"
	"strconv"
	"strings"
)

// Step is a single step in a pipeline.
type Step struct {
	Successful   bool      `json:"successful" yaml:"-" header:"successful"`
	Completed    bool      `json:"completed" yaml:"-" header:"completed"`
	Name         string    `json:"name" yaml:"name" header:"name"`
	Image        string    `json:"image" yaml:"image" header:"image"`
	Commands     []string  `json:"commands" yaml:"commands" header:"commands"`
	Errors       []string  `json:"errors" yaml:"-" header:"errors"`
	Volume       string    `json:"volume" yaml:"-"`
	OutputStream io.Writer `json:"-" yaml:"-"`
	ErrorStream  io.Writer `json:"-" yaml:"-"`
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

// toImageAndTag parses the image into name and tag.
func toImageAndTag(image string) (name string, tag string) {
	tag = "latest"
	split := strings.Split(image, ":")
	if len(split) > 1 {
		name = split[0]
		tag = split[1]
	} else {
		name = image
	}
	return name, tag
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
			Cmd:   []string{"/bin/sh", "-c", strings.Join(s.Commands, "\n")},
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
		if err == docker.ErrNoSuchImage {
			image, tag := toImageAndTag(s.Image)
			err := client.PullImage(docker.PullImageOptions{
				Repository: image,
				Tag:        tag,
			}, docker.AuthConfiguration{})
			if err != nil {
				s.Successful = false
				s.Completed = true
				s.Errors = append(s.Errors, err.Error())
				return err
			}
			container, err = client.CreateContainer(docker.CreateContainerOptions{
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
		} else {
			s.Successful = false
			s.Completed = true
			s.Errors = append(s.Errors, err.Error())
			return err
		}
	}
	err = client.StartContainer(container.ID, nil)
	if err != nil {
		s.Successful = false
		s.Completed = true
		s.Errors = append(s.Errors, err.Error())
		return err
	}
	err = client.Logs(docker.LogsOptions{
		Container:    container.ID,
		OutputStream: s.OutputStream,
		ErrorStream:  s.ErrorStream,
		Stdout:       true,
		Stderr:       true,
		Follow:       true,
		Tail:         "all",
	})
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
