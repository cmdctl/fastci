package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func CaseImageWithTag(t *testing.T) {
	step := NewStep("image", "image:tag", []string{})
	assert.Equal(t, "image", step.Image)
	assert.Equal(t, "tag", step.ImageTag)
}

func CaseImageWithoutTag(t *testing.T) {
	step := NewStep("image", "image", []string{})
	assert.Equal(t, "image", step.Image)
	assert.Equal(t, "latest", step.ImageTag)
}

func TestNewStep(t *testing.T) {
	var testCases = map[string]func(t *testing.T){
		"CaseImageWithTag":    CaseImageWithTag,
		"CaseImageWithoutTag": CaseImageWithoutTag,
	}
	for name, test := range testCases {
		t.Run(name, test)
	}
}
