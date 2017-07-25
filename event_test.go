package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnevenEventArguments(t *testing.T) {
	assert := assert.New(t)
	_, err := NewEvent("test_event", "Testing {hehe}", 1, 2, 3)
	assert.NotNil(err)
}

func TestMapping(t *testing.T) {
	assert := assert.New(t)
	e, err := NewEvent("test_event",
		"Testing {hehe}",
		"hehe", 1,
		"foo", 2,
		"person", nil)
	assert.Nil(err)

	assert.Equal(e.Items["hehe"], 1)
	assert.Equal(e.Items["foo"], 2)
}
