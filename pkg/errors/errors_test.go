package errors

import (
	"errors"
	"testing"

	"github.com/palantir/stacktrace"
	"github.com/stretchr/testify/assert"
)

func TestStacktraceUnwrap(t *testing.T) {
	cause := errors.New("test")
	assert.Equal(t, cause, errors.Unwrap(stacktrace.Propagate(cause, "test")))
}
