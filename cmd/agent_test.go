package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgent(t *testing.T) {
	cmd := agentCmd
	err := cmd.Execute()
	assert.NoError(t, err)
}
