package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrent(t *testing.T) {
	// Make sure that parsing on init is permissive.
	assert.NotEmpty(t, Current().String())
}
