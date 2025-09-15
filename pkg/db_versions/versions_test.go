package dbversions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestVersionFileIsRead checks that the schema version file is read and parsed.
func TestVersionIsParsed(t *testing.T) {
	types := []string{Aux, Rid, Scd}

	for _, tp := range types {
		_, err := GetCurrentMajorCRDBSchemaVersion(tp)
		require.NoError(t, err)
	}
}
