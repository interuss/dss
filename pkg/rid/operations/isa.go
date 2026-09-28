package operations

import (
	"context"

	"github.com/golang/geo/s2"
	ridv2 "github.com/interuss/dss/pkg/api/ridv2"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

// ISAResult bundles the affected ISA together with the relevant Subscriptions
type ISAResult struct {
	ISA           *ridmodels.IdentificationServiceArea
	Subscriptions []*ridmodels.Subscription
}

type deleteISAPayload struct {
	ID      dssmodels.ID
	Owner   dssmodels.Owner
	Version *dssmodels.Version
}

func (p *deleteISAPayload) OperationID() string {
	return ridv2.DeleteIdentificationServiceAreaOperationID
}

// NewDeleteISAPayload performs the request validation that can be done ahead of the transaction
// for an ISA deletion request.
func NewDeleteISAPayload(id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) dssstore.OperationRequest {
	return &deleteISAPayload{ID: id, Owner: owner, Version: version}
}

type insertISAPayload struct {
	ISA *ridmodels.IdentificationServiceArea
}

func (p *insertISAPayload) OperationID() string {
	return ridv2.CreateIdentificationServiceAreaOperationID
}

// NewInsertISAPayload performs the request validation that can be done ahead of the transaction
// for an ISA creation request.
func NewInsertISAPayload(id dssmodels.ID, owner dssmodels.Owner, url string, writer string, extents *dssmodels.Volume4D, allowHTTPBaseUrls bool) (dssstore.OperationRequest, error) {
	isa, err := newISA(id, owner, url, writer, nil, extents, allowHTTPBaseUrls)
	if err != nil {
		return nil, err
	}
	return &insertISAPayload{ISA: isa}, nil
}

type updateISAPayload struct {
	ISA *ridmodels.IdentificationServiceArea
}

func (p *updateISAPayload) OperationID() string {
	return ridv2.UpdateIdentificationServiceAreaOperationID
}

// NewUpdateISAPayload performs the request validation that can be done ahead of the transaction
// for an ISA update request.
func NewUpdateISAPayload(id dssmodels.ID, owner dssmodels.Owner, url string, writer string, version *dssmodels.Version, extents *dssmodels.Volume4D, allowHTTPBaseUrls bool) (dssstore.OperationRequest, error) {
	isa, err := newISA(id, owner, url, writer, version, extents, allowHTTPBaseUrls)
	if err != nil {
		return nil, err
	}
	return &updateISAPayload{ISA: isa}, nil
}

// newISA performs the request validation that can be done ahead of the transaction.
func newISA(id dssmodels.ID, owner dssmodels.Owner, url string, writer string, version *dssmodels.Version, extents *dssmodels.Volume4D, allowHTTPBaseUrls bool) (*ridmodels.IdentificationServiceArea, error) {
	if !allowHTTPBaseUrls {
		if err := ridmodels.ValidateURL(url); err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate ISA URL")
		}
	}

	cells, err := extents.SpatialVolume.Footprint.CalculateCovering()
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents")
	}

	return &ridmodels.IdentificationServiceArea{
		ID:      id,
		Owner:   owner,
		URL:     url,
		Writer:  writer,
		Version: version,
		CellsVolume4D: &dssmodels.CellsVolume4D{
			Cells:      cells,
			StartTime:  extents.StartTime,
			EndTime:    extents.EndTime,
			AltitudeLo: extents.SpatialVolume.AltitudeLo,
			AltitudeHi: extents.SpatialVolume.AltitudeHi,
		},
	}, nil
}

func init() {
	Registry[ridv2.DeleteIdentificationServiceAreaOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*deleteISAPayload],
		Execute: executeDeleteISA,
	}
	Registry[ridv2.CreateIdentificationServiceAreaOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*insertISAPayload],
		Execute: executeInsertISA,
	}
	Registry[ridv2.UpdateIdentificationServiceAreaOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*updateISAPayload],
		Execute: executeUpdateISA,
	}
}

func executeDeleteISA(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*deleteISAPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.DeleteIdentificationServiceAreaOperationID)
	}

	old, err := repo.GetISA(ctx, payload.ID, true)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Error getting ISA")
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "ISA %s not found", payload.ID.String())
	case !payload.Version.Matches(old.Version):
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"ISA currently at version %s but client specified %s", old.Version, payload.Version)
	case old.Owner != payload.Owner:
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"ISA owned by %s, but %s attempted to delete", old.Owner, payload.Owner)
	}

	ret, err := repo.DeleteISA(ctx, old)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error deleting ISA")
	}

	subs, err := repo.UpdateNotificationIdxsInCells(ctx, old.Cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error updating notification indices")
	}

	return &ISAResult{ISA: ret, Subscriptions: subs}, nil
}

func executeInsertISA(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*insertISAPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.CreateIdentificationServiceAreaOperationID)
	}

	isa := payload.ISA

	// Validate and perhaps correct StartTime and EndTime.
	if err := isa.AdjustTimeRange(timestamp.MustFromContext(ctx), nil); err != nil {
		return nil, stacktrace.Propagate(err, "Error adjusting time range")
	}

	// ensure it doesn't exist yet
	old, err := repo.GetISA(ctx, isa.ID, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error getting ISA")
	}
	if old != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.AlreadyExists, "ISA %s already exists", isa.ID)
	}

	// UpdateNotificationIdxsInCells is done in the same transaction as the insert since they
	// are both modifying the store.
	subs, err := repo.UpdateNotificationIdxsInCells(ctx, isa.Cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error updating notification indices")
	}

	ret, err := repo.InsertISA(ctx, isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error inserting ISA")
	}

	return &ISAResult{ISA: ret, Subscriptions: subs}, nil
}

func executeUpdateISA(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	payload, ok := request.(*updateISAPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.UpdateIdentificationServiceAreaOperationID)
	}

	isa := payload.ISA

	old, err := repo.GetISA(ctx, isa.ID, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error getting ISA")
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := isa.AdjustTimeRange(timestamp.MustFromContext(ctx), old); err != nil {
		return nil, stacktrace.Propagate(err, "Error adjusting time range")
	}

	switch {
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "ISA %s not found", isa.ID.String())
	case !isa.Version.Matches(old.Version):
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"ISA currently at version %s but client specified %s", old.Version, isa.Version)
	case old.Owner != isa.Owner:
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"ISA owned by %s, but %s attempted to modify", old.Owner, isa.Owner)
	}

	// TODO steeling, we should change this to a Custom type, to obfuscate
	// some of these metrics and prevent us from doing the wrong thing.
	cells := s2.CellUnionFromUnion(old.Cells, isa.Cells)
	geo.Levelify(&cells)

	// UpdateNotificationIdxsInCells is done in the same transaction as the update since they
	// are both modifying the store.
	subs, err := repo.UpdateNotificationIdxsInCells(ctx, cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error updating notification indices")
	}

	ret, err := repo.UpdateISA(ctx, isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error updating ISA in repo")
	}
	return &ISAResult{ISA: ret, Subscriptions: subs}, nil
}
