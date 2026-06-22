package memstore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckpointRestore(t *testing.T) {
	ctx := context.Background()
	r := newRepo()

	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://example.com"))

	cp := r.Checkpoint()

	// Mutate after the checkpoint.
	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-2", "https://other.example.com"))
	md, err := r.GetDSSMetadata(ctx)
	require.NoError(t, err)
	require.Len(t, md, 2)

	// Restore drops dss-2.
	require.NoError(t, r.Restore(cp))
	md, err = r.GetDSSMetadata(ctx)
	require.NoError(t, err)
	require.Len(t, md, 1)
	require.Equal(t, "dss-1", md[0].Locality)
}

func TestCheckpointIsolatesUpsert(t *testing.T) {
	ctx := context.Background()
	r := newRepo()

	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://old.example.com"))

	cp := r.Checkpoint()

	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://new.example.com"))

	require.NoError(t, r.Restore(cp))
	md, err := r.GetDSSMetadata(ctx)
	require.NoError(t, err)
	require.Len(t, md, 1)
	require.Equal(t, "https://old.example.com", md[0].PublicEndpoint)
}

func TestRestoreInvalidType(t *testing.T) {
	require.Error(t, newRepo().Restore("not a checkpoint"))
}
