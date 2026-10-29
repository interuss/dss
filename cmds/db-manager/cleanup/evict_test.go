package cleanup

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestDeleteExpiredEntities(t *testing.T) {
	threshold := time.Now()
	makeIDs := func(n int) []dssmodels.ID {
		ids := make([]dssmodels.ID, n)
		for i := range n {
			ids[i] = dssmodels.ID(fmt.Sprintf("id-%d", i+1))
		}
		return ids
	}

	tests := []struct {
		name           string
		totalExpired   int
		limit          int
		batchSize      int
		wantDeletedLen int
		wantCalls      []int
	}{
		{
			name:           "single query unlimited",
			totalExpired:   5,
			limit:          0,
			batchSize:      0,
			wantDeletedLen: 5,
			wantCalls:      []int{0},
		},
		{
			name:           "single query with limit",
			totalExpired:   5,
			limit:          3,
			batchSize:      0,
			wantDeletedLen: 3,
			wantCalls:      []int{3},
		},
		{
			name:           "batched without limit non-multiple total",
			totalExpired:   5,
			limit:          0,
			batchSize:      2,
			wantDeletedLen: 5,
			wantCalls:      []int{2, 2, 2},
		},
		{
			name:           "batched without limit exact multiple total",
			totalExpired:   4,
			limit:          0,
			batchSize:      2,
			wantDeletedLen: 4,
			wantCalls:      []int{2, 2, 2},
		},
		{
			name:           "batched with limit exact multiple of batch size",
			totalExpired:   6,
			limit:          4,
			batchSize:      2,
			wantDeletedLen: 4,
			wantCalls:      []int{2, 2},
		},
		{
			name:           "batched with limit exceeded by batch_size - 1",
			totalExpired:   6,
			limit:          3,
			batchSize:      2,
			wantDeletedLen: 4,
			wantCalls:      []int{2, 2},
		},
		{
			name:           "batched with limit larger than available expired entities",
			totalExpired:   3,
			limit:          10,
			batchSize:      2,
			wantDeletedLen: 3,
			wantCalls:      []int{2, 2},
		},
		{
			name:           "batched with batch size larger than limit",
			totalExpired:   6,
			limit:          2,
			batchSize:      4,
			wantDeletedLen: 4,
			wantCalls:      []int{4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remaining := makeIDs(tt.totalExpired)
			var calls []int

			deleteFn := func(_ context.Context, gotThreshold time.Time, queryLimit int) ([]dssmodels.ID, error) {
				require.Equal(t, threshold, gotThreshold)
				calls = append(calls, queryLimit)
				n := len(remaining)
				if queryLimit > 0 && queryLimit < n {
					n = queryLimit
				}
				batch := slices.Clone(remaining[:n])
				remaining = remaining[n:]
				return batch, nil
			}

			deleted, err := deleteExpiredEntities(context.Background(), deleteFn, threshold, tt.limit, tt.batchSize)
			require.NoError(t, err)
			require.Equal(t, makeIDs(tt.wantDeletedLen), deleted)
			require.Equal(t, tt.wantCalls, calls)
		})
	}
}

func TestDeleteExpiredEntitiesError(t *testing.T) {
	wantErr := errors.New("db error")
	calls := 0
	deleteFn := func(_ context.Context, _ time.Time, _ int) ([]dssmodels.ID, error) {
		calls++
		if calls == 2 {
			return nil, wantErr
		}
		return []dssmodels.ID{"id-1", "id-2"}, nil
	}

	_, err := deleteExpiredEntities(context.Background(), deleteFn, time.Now(), 0, 2)
	require.ErrorIs(t, err, wantErr)
	require.Equal(t, 2, calls)
}

func TestEvictFlagValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "negative scd_limit",
			args:    []string{"--delete", "--scd_limit=-1"},
			wantErr: "scd_limit must be equal to or greater than 0, got -1",
		},
		{
			name:    "negative rid_limit",
			args:    []string{"--delete", "--rid_limit=-1"},
			wantErr: "rid_limit must be equal to or greater than 0, got -1",
		},
		{
			name:    "negative scd_batch_size",
			args:    []string{"--delete", "--scd_batch_size=-1"},
			wantErr: "scd_batch_size must be equal to or greater than 0, got -1",
		},
		{
			name:    "negative rid_batch_size",
			args:    []string{"--delete", "--rid_batch_size=-1"},
			wantErr: "rid_batch_size must be equal to or greater than 0, got -1",
		},
		{
			name:    "scd_limit without delete",
			args:    []string{"--scd_limit=10"},
			wantErr: "scd_limit can only be set together with --delete",
		},
		{
			name:    "rid_limit without delete",
			args:    []string{"--rid_limit=10"},
			wantErr: "rid_limit can only be set together with --delete",
		},
		{
			name:    "scd_batch_size without delete",
			args:    []string{"--scd_batch_size=10"},
			wantErr: "scd_batch_size can only be set together with --delete",
		},
		{
			name:    "rid_batch_size without delete",
			args:    []string{"--rid_batch_size=10"},
			wantErr: "rid_batch_size can only be set together with --delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:  "evict",
				RunE: evict,
			}
			cmd.Flags().AddFlagSet(flags)
			t.Cleanup(func() {
				*deleteExpired = false
				*scdLimit = 0
				*ridLimit = 0
				*scdBatchSize = 0
				*ridBatchSize = 0
				for _, name := range []string{"delete", "scd_limit", "rid_limit", "scd_batch_size", "rid_batch_size"} {
					if f := flags.Lookup(name); f != nil {
						f.Changed = false
					}
				}
			})

			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}
