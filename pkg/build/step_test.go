package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToImageAndTag(t *testing.T) {
	imageName := "ubuntu"
	image, tag := toImageAndTag(imageName)
	assert.Equal(t, "ubuntu", image)
	assert.Equal(t, "latest", tag)
}