package memstore

import (
	"context"
	"testing"
	"time"

	auxmodels "github.com/interuss/dss/pkg/aux_/models"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/require"
)

func TestSaveOwnMetadataValidation(t *testing.T) {
	ctx := context.Background()
	r := newRepo()

	require.Equal(t, dsserr.BadRequest, stacktrace.GetCode(r.SaveOwnMetadata(ctx, "", "https://example.com")))
	require.Equal(t, dsserr.BadRequest, stacktrace.GetCode(r.SaveOwnMetadata(ctx, "dss-1", "")))
}

func TestSaveOwnMetadataRoundTrip(t *testing.T) {
	ctx := context.Background()
	r := newRepo()

	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://example.com"))

	md, err := r.GetDSSMetadata(ctx)
	require.NoError(t, err)

	require.Len(t, md, 1)
	require.Equal(t, "dss-1", md[0].Locality)
	require.Equal(t, "https://example.com", md[0].PublicEndpoint)
	require.NotNil(t, md[0].UpdatedAt)

	// No heartbeat recorded yet.
	require.False(t, md[0].LatestTimestamp.Source.Valid)
	require.Nil(t, md[0].LatestTimestamp.Timestamp)
}

func TestSaveOwnMetadataUpsert(t *testing.T) {
	ctx := context.Background()
	r := newRepo()

	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://old.example.com"))
	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://new.example.com"))

	md, err := r.GetDSSMetadata(ctx)
	require.NoError(t, err)

	require.Len(t, md, 1)
	require.Equal(t, "https://new.example.com", md[0].PublicEndpoint)
}

func TestRecordHeartbeatValidation(t *testing.T) {
	ctx := context.Background()
	r := newRepo()

	require.Equal(t, dsserr.BadRequest, stacktrace.GetCode(r.RecordHeartbeat(ctx, auxmodels.Heartbeat{Source: "source1"})))
	require.Equal(t, dsserr.BadRequest, stacktrace.GetCode(r.RecordHeartbeat(ctx, auxmodels.Heartbeat{Locality: "dss-1"})))

	ts := time.Now()
	before := ts.Add(-time.Minute)
	err := r.RecordHeartbeat(ctx, auxmodels.Heartbeat{
		Locality:                    "dss-1",
		Source:                      "source1",
		Timestamp:                   &ts,
		NextHeartbeatExpectedBefore: &before,
	})

	require.Equal(t, dsserr.BadRequest, stacktrace.GetCode(err))
}

func TestRecordHeartbeatDefaultsTimestamp(t *testing.T) {
	ctx := context.Background()
	r := newRepo()

	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://example.com"))
	require.NoError(t, r.RecordHeartbeat(ctx, auxmodels.Heartbeat{Locality: "dss-1", Source: "source1"}))

	md, err := r.GetDSSMetadata(ctx)
	require.NoError(t, err)

	require.Len(t, md, 1)
	require.True(t, md[0].LatestTimestamp.Source.Valid)
	require.NotNil(t, md[0].LatestTimestamp.Timestamp)
}

func TestGetDSSMetadataPicksLatestHeartbeat(t *testing.T) {
	ctx := context.Background()
	r := newRepo()

	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://example.com"))

	older := time.Now().Add(-time.Hour)
	newer := time.Now()
	require.NoError(t, r.RecordHeartbeat(ctx, auxmodels.Heartbeat{Locality: "dss-1", Source: "source1", Timestamp: &older, Reporter: "uss1"}))
	require.NoError(t, r.RecordHeartbeat(ctx, auxmodels.Heartbeat{Locality: "dss-1", Source: "source2", Timestamp: &newer, Reporter: "uss2"}))

	md, err := r.GetDSSMetadata(ctx)
	require.NoError(t, err)

	require.Len(t, md, 1)
	require.True(t, md[0].LatestTimestamp.Timestamp.Equal(newer))
	require.Equal(t, "source2", md[0].LatestTimestamp.Source.String)
	require.Equal(t, "uss2", md[0].LatestTimestamp.Reporter.String)
}

func TestGetDSSMetadataUpdatesHeartbeatPerSource(t *testing.T) {
	ctx := context.Background()
	r := newRepo()

	require.NoError(t, r.SaveOwnMetadata(ctx, "dss-1", "https://example.com"))

	first := time.Now().Add(-time.Hour)
	second := time.Now()
	require.NoError(t, r.RecordHeartbeat(ctx, auxmodels.Heartbeat{Locality: "dss-1", Source: "source1", Timestamp: &first}))
	require.NoError(t, r.RecordHeartbeat(ctx, auxmodels.Heartbeat{Locality: "dss-1", Source: "source1", Timestamp: &second}))

	md, err := r.GetDSSMetadata(ctx)
	require.NoError(t, err)

	require.Len(t, md, 1)
	require.True(t, md[0].LatestTimestamp.Timestamp.Equal(second))
}
