package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolStat(t *testing.T) {
	err := ToolStat()
	assert.Equal(t, nil, err)
}
