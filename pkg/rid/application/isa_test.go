package application

import (
	"context"
	"testing"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	_ ISAApp = &app{}
)

func setUpISAApp(ctx context.Context, t *testing.T) (*app, func()) {
	l := zap.L()
	transactor, cleanup := setUpStore(ctx, t, l)
	return NewFromTransactor(transactor, l).(*app), cleanup
}

// TODO:steeling add owner logic.
type isaStore struct {
	isas map[dssmodels.ID]*ridmodels.IdentificationServiceArea
}

func (store *isaStore) GetISA(ctx context.Context, id dssmodels.ID, forUpdate bool) (*ridmodels.IdentificationServiceArea, error) {
	if isa, ok := store.isas[id]; ok {
		return isa, nil
	}
	return nil, nil
}

// Implements repos.ISA.DeleteISA
func (store *isaStore) DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	isa, ok := store.isas[isa.ID]
	if !ok {
		return nil, nil
	}
	delete(store.isas, isa.ID)

	return isa, nil
}

// Implements repos.ISA.InsertISA
func (store *isaStore) InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	storedCopy := *isa
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	store.isas[isa.ID] = &storedCopy

	returnedCopy := storedCopy
	return &returnedCopy, nil
}

// Implements repos.ISA.UpdateISA
func (store *isaStore) UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error) {
	storedCopy := *isa
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	store.isas[isa.ID] = &storedCopy
	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (store *isaStore) GetVersion(ctx context.Context) (*semver.Version, error) {
	return semver.NewVersion("v3.1.1")
}

// Implements repos.ISA.SearchISA
func (store *isaStore) SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	var isas []*ridmodels.IdentificationServiceArea

	for _, isa := range store.isas {
		if isa.Cells.Intersects(cells) {
			isas = append(isas, isa)
		}
	}
	return isas, nil
}

// Implements repos.ISA.ListExpiredISAs
func (store *isaStore) ListExpiredISAs(ctx context.Context, writer string, threshold time.Time) ([]*ridmodels.IdentificationServiceArea, error) {
	return make([]*ridmodels.IdentificationServiceArea, 0), nil
}

// Implements repos.ISA.CountISAs
func (store *isaStore) CountISAs(ctx context.Context) (int64, error) {
	return int64(len(store.isas)), nil
}

func TestUpdateISA(t *testing.T) {
	ctx := context.Background()
	app, cleanup := setUpISAApp(ctx, t)

	defer cleanup()

	for _, r := range []struct {
		name                string
		updateFromStartTime time.Time
		updateFromEndTime   time.Time
		startTime           time.Time
		endTime             time.Time
		wantErr             stacktrace.ErrorCode
		wantStartTime       time.Time
		wantEndTime         time.Time
	}{
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
			wantErr:             dsserr.BadRequest,
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
			repo, err := app.store.Interact(ctx)
			require.NoError(t, err)
			// Insert a pre-existing ISA to simulate updating from something.
			existing, err := repo.InsertISA(ctx, &ridmodels.IdentificationServiceArea{
				ID:        id,
				Owner:     owner,
				StartTime: &r.updateFromStartTime,
				EndTime:   &r.updateFromEndTime,
				Cells:     s2.CellUnion{12494535935418957824},
			})
			require.NoError(t, err)

			sa := &ridmodels.IdentificationServiceArea{
				ID:      id,
				Owner:   owner,
				Version: existing.Version,
				Cells:   s2.CellUnion{12494535935418957824},
			}
			if !r.startTime.IsZero() {
				sa.StartTime = &r.startTime
			}
			if !r.endTime.IsZero() {
				sa.EndTime = &r.endTime
			}
			isa, _, err := app.UpdateISA(ctx, sa)

			if r.wantErr == stacktrace.ErrorCode(0) {
				require.NoError(t, err)
			} else {
				require.Equal(t, r.wantErr, stacktrace.GetCode(err))
			}

			if !r.wantStartTime.IsZero() {
				require.NotNil(t, isa.StartTime)
				require.Equal(t, r.wantStartTime.UTC().Truncate(time.Microsecond), (*isa.StartTime).UTC().Truncate(time.Microsecond))
			}
			if !r.wantEndTime.IsZero() {
				require.NotNil(t, isa.EndTime)
				require.Equal(t, r.wantEndTime.UTC().Truncate(time.Microsecond), (*isa.EndTime).UTC().Truncate(time.Microsecond))
			}
		})
	}
}
