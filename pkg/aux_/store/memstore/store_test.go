package memstore

import (
	"context"
	"testing"

	"github.com/interuss/dss/pkg/timestamp"
	"github.com/stretchr/testify/require"
)

func TestCheckpointRestore(t *testing.T) {
	ctx := context.Background()
	ctx = timestamp.NewContext(ctx, fakeClock.Now())

	r := newRepo()

	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://example.com"))

	r.Checkpoint()

	// Mutate after the checkpoint.
	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-2", "https://other.example.com"))
	md, err := r.GetDSSMetadata(ctx)
	require.NoError(t, err)
	require.Len(t, md, 2)

	// Restore drops dss-2.
	r.Restore()
	md, err = r.GetDSSMetadata(ctx)
	require.NoError(t, err)
	require.Len(t, md, 1)
	require.Equal(t, "dss-1", md[0].Locality)
}

func TestCheckpointIsolatesUpsert(t *testing.T) {
	ctx := context.Background()
	ctx = timestamp.NewContext(ctx, fakeClock.Now())
	r := newRepo()

	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://old.example.com"))

	r.Checkpoint()

	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://new.example.com"))

	r.Restore()
	md, err := r.GetDSSMetadata(ctx)
	require.NoError(t, err)
	require.Len(t, md, 1)
	require.Equal(t, "https://old.example.com", md[0].PublicEndpoint)
}
