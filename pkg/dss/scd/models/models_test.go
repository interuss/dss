package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOVNFromTimeIsValid(t *testing.T) {
	require.True(t, NewOVNFromTime(time.Now()).Valid())
}
