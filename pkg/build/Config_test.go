package build

import (
	"github.com/cmdctl/fastci/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFromYaml(t *testing.T) {
	file, err := test.YamlTestFS.ReadFile("data/correct-config.yaml")
	if err != nil {
		t.Fatal("Failed to read file:", err)
	}
	config, err := FromYaml(file)
	if err != nil {
		t.Fatal("Failed to parse config:", err)
	}
	assert.Equal(t, "1", config.Version)
	assert.Equal(t, "build app", config.Build.Name)
	assert.Equal(t, 3, len(config.Build.Pipeline.Steps))
}
