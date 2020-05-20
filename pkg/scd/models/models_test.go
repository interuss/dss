package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOVNFromTimeIsValid(t *testing.T) {
	require.True(t, NewOVNFromTime(time.Now(), uuid.New().String()).Valid())
}
