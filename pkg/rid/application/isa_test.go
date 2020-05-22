package application

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/stretchr/testify/require"
)

func setUpISAApp() ISAApp {
	return &app{
		ISA: &isaStore{
			isas: make(map[dssmodels.ID]*ridmodels.IdentificationServiceArea),
		},
		Subscription: &subscriptionStore{
			subs: make(map[dssmodels.ID]*ridmodels.Subscription),
		},
		clock: fakeClock,
	}
}

// TODO:steeling add owner logic.
type isaStore struct {
	isas map[dssmodels.ID]*ridmodels.IdentificationServiceArea
}

func (store *isaStore) GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error) {
	if isa, ok := store.isas[id]; ok {
		return isa, nil
	}
	return nil, sql.ErrNoRows
}

// DeleteISA deletes the IdentificationServiceArea identified by "id" and owned by "owner".
// Returns the delete IdentificationServiceArea and all IdentificationServiceAreas affected by the delete.
func (store *isaStore) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	isa, ok := store.isas[isa.ID]
	if !ok {
		return nil, nil, sql.ErrNoRows
	}
	delete(store.isas, isa.ID)

	return isa, store.updateNotificationIdxs(ctx, isa), nil
}

// InsertISA inserts or updates an IdentificationServiceArea.
func (store *isaStore) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	storedCopy := *isa
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	store.isas[isa.ID] = &storedCopy

	returnedCopy := storedCopy
	return &returnedCopy, store.updateNotificationIdxs(ctx, &storedCopy), nil
}

func (store *isaStore) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, []*ridmodels.Subscription, error) {
	return store.InsertISA(ctx, isa)
}

func (store *isaStore) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	var isas []*ridmodels.IdentificationServiceArea

	for _, isa := range store.isas {
		if isa.Cells.Intersects(cells) {
			isas = append(isas, isa)
		}
	}
	return isas, nil
}

func TestInsertISA(t *testing.T) {
	ctx := context.Background()
	app := setUpISAApp()

	for _, r := range []struct {
		name                string
		updateFromStartTime time.Time
		updateFromEndTime   time.Time
		startTime           time.Time
		endTime             time.Time
		wantErr             string
		wantStartTime       time.Time
		wantEndTime         time.Time
	}{
		{
			name:    "missing-end-time",
			wantErr: "rpc error: code = InvalidArgument desc = IdentificationServiceArea must have an time_end",
		},
		{
			name:          "start-time-defaults-to-now",
			endTime:       fakeClock.Now().Add(time.Hour),
			wantStartTime: fakeClock.Now(),
		},
		{
			name:      "start-time-in-the-past",
			startTime: fakeClock.Now().Add(-6 * time.Minute),
			endTime:   fakeClock.Now().Add(time.Hour),
			wantErr:   "rpc error: code = InvalidArgument desc = IdentificationServiceArea time_start must not be in the past",
		},
		{
			name:          "start-time-slightly-in-the-past",
			startTime:     fakeClock.Now().Add(-4 * time.Minute),
			endTime:       fakeClock.Now().Add(time.Hour),
			wantStartTime: fakeClock.Now().Add(-4 * time.Minute),
		},
		{
			name:      "end-time-before-start-time",
			startTime: fakeClock.Now().Add(20 * time.Minute),
			endTime:   fakeClock.Now().Add(10 * time.Minute),
			wantErr:   "rpc error: code = InvalidArgument desc = IdentificationServiceArea time_end must be after time_start",
		},
		{
			name:                "updating-keeps-old-times",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(-6 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(6 * time.Hour),
		},
		{
			name:                "changing-start-time-to-past",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			startTime:           fakeClock.Now().Add(-3 * time.Hour),
			wantErr:             "rpc error: code = InvalidArgument desc = IdentificationServiceArea time_start must not be in the past",
		},
		{
			name:                "changing-start-time-to-future",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			startTime:           fakeClock.Now().Add(3 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(3 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(6 * time.Hour),
		},
		{
			name:                "changing-end-time-to-future",
			updateFromStartTime: fakeClock.Now().Add(-6 * time.Hour),
			updateFromEndTime:   fakeClock.Now().Add(6 * time.Hour),
			endTime:             fakeClock.Now().Add(3 * time.Hour),
			wantStartTime:       fakeClock.Now().Add(-6 * time.Hour),
			wantEndTime:         fakeClock.Now().Add(3 * time.Hour),
		},
	} {
		t.Run(r.name, func(t *testing.T) {
			id := dssmodels.ID(uuid.New().String())
			owner := dssmodels.Owner(uuid.New().String())
			var version *dssmodels.Version

			// Insert a pre-existing ISA to simulate updating from something.
			if !r.updateFromStartTime.IsZero() {
				existing, _, err := app.InsertISA(ctx, &ridmodels.IdentificationServiceArea{
					ID:        id,
					Owner:     owner,
					StartTime: &r.updateFromStartTime,
					EndTime:   &r.updateFromEndTime,
				})
				require.NoError(t, err)
				version = existing.Version

				// Can't update if it has a different owner
				isa := *existing
				isa.Owner = "bad-owner"
				_, _, err = app.InsertISA(ctx, &isa)
				require.Error(t, err)
			}

			sa := &ridmodels.IdentificationServiceArea{
				ID:      id,
				Owner:   owner,
				Version: version,
			}
			if !r.startTime.IsZero() {
				sa.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				sa.EndTime = &r.endTime
			}
			isa, _, err := app.InsertISA(ctx, sa)

			if r.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, r.wantErr)
			}

			if !r.wantStartTime.IsZero() {
				require.NotNil(t, isa.StartTime)
				require.Equal(t, r.wantStartTime, *isa.StartTime)
			}
			if !r.wantEndTime.IsZero() {
				require.NotNil(t, isa.EndTime)
				require.Equal(t, r.wantEndTime, *isa.EndTime)
			}
		})
	}
}
