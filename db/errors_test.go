package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultsPredicate(t *testing.T) {
	assert := assert.New(t)

	assert.False(ResultsNotFound(errors.New("foo")))
	assert.False(ResultsNotFound(nil))
	assert.True(ResultsNotFound(errors.New("not found")))
}
