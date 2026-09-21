package operations

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	ridv1 "github.com/interuss/dss/pkg/api/ridv1"
	ridv2 "github.com/interuss/dss/pkg/api/ridv2"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	"github.com/interuss/dss/pkg/locality"
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

// putISAPayload carries an already-validated rid ISA create/update request: the HTTP handler
// (v1 or v2) does all request-format validation and computes the S2 covering once, so it isn't
// recomputed by every raft node during apply.
type putISAPayload struct {
	Op         string
	ID         dssmodels.ID
	Owner      dssmodels.Owner
	URL        string
	Version    *dssmodels.Version
	Cells      s2.CellUnion
	StartTime  *time.Time
	EndTime    *time.Time
	AltitudeLo *float32
	AltitudeHi *float32
}

func (p *putISAPayload) OperationID() string { return p.Op }

// NewPutISAPayload builds a putISAPayload. id, owner, url, and version are assumed already
// validated by the caller; extents is used only to compute the S2 covering.
func NewPutISAPayload(opID string, id dssmodels.ID, owner dssmodels.Owner, url string, version *dssmodels.Version, extents *dssmodels.Volume4D) (*putISAPayload, error) {
	cells, err := extents.SpatialVolume.Footprint.CalculateCovering()
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents")
	}

	return &putISAPayload{
		Op:         opID,
		ID:         id,
		Owner:      owner,
		URL:        url,
		Version:    version,
		Cells:      cells,
		StartTime:  extents.StartTime,
		EndTime:    extents.EndTime,
		AltitudeLo: extents.SpatialVolume.AltitudeLo,
		AltitudeHi: extents.SpatialVolume.AltitudeHi,
	}, nil
}

func init() {
	Registry[ridv1.DeleteIdentificationServiceAreaOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*ridv1.DeleteIdentificationServiceAreaRequest],
		Execute: executeDeleteISA,
	}
	Registry[ridv2.DeleteIdentificationServiceAreaOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*ridv2.DeleteIdentificationServiceAreaRequest],
		Execute: executeDeleteISA,
	}
	Registry[ridv1.CreateIdentificationServiceAreaOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*putISAPayload],
		Execute: executeInsertISA,
	}
	Registry[ridv2.CreateIdentificationServiceAreaOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*putISAPayload],
		Execute: executeInsertISA,
	}
	Registry[ridv1.UpdateIdentificationServiceAreaOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*putISAPayload],
		Execute: executeUpdateISA,
	}
	Registry[ridv2.UpdateIdentificationServiceAreaOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*putISAPayload],
		Execute: executeUpdateISA,
	}
}

func executeDeleteISA(ctx context.Context, repo repos.Repository, request dssstore.OperationRequest) (any, error) {
	var (
		rawID      string
		rawVersion string
		clientID   *string
	)

	switch req := request.(type) {
	case *ridv1.DeleteIdentificationServiceAreaRequest:
		rawID, rawVersion, clientID = string(req.Id), req.Version, req.Auth.ClientID
	case *ridv2.DeleteIdentificationServiceAreaRequest:
		rawID, rawVersion, clientID = string(req.Id), req.Version, req.Auth.ClientID
	default:
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.DeleteIdentificationServiceAreaOperationID)
	}

	version, err := dssmodels.VersionFromString(rawVersion)
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid version")
	}
	id, err := dssmodels.IDFromString(rawID)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}
	owner := dssmodels.Owner(*clientID)

	return deleteISA(ctx, repo, id, owner, version)
}

func deleteISA(ctx context.Context, repo repos.Repository, id dssmodels.ID, owner dssmodels.Owner, version *dssmodels.Version) (*ISAResult, error) {
	old, err := repo.GetISA(ctx, id, true)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Error getting ISA")
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "ISA %s not found", id.String())
	case !version.Matches(old.Version):
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"ISA currently at version %s but client specified %s", old.Version, version)
	case old.Owner != owner:
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"ISA owned by %s, but %s attempted to delete", old.Owner, owner)
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
	payload, ok := request.(*putISAPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.CreateIdentificationServiceAreaOperationID)
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:         payload.ID,
		Owner:      payload.Owner,
		URL:        payload.URL,
		Cells:      payload.Cells,
		StartTime:  payload.StartTime,
		EndTime:    payload.EndTime,
		AltitudeLo: payload.AltitudeLo,
		AltitudeHi: payload.AltitudeHi,
		Writer:     locality.MustFromContext(ctx),
	}

	return insertISA(ctx, repo, isa)
}

func insertISA(ctx context.Context, repo repos.Repository, isa *ridmodels.IdentificationServiceArea) (*ISAResult, error) {
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
	payload, ok := request.(*putISAPayload)
	if !ok {
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.UpdateIdentificationServiceAreaOperationID)
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:         payload.ID,
		Owner:      payload.Owner,
		URL:        payload.URL,
		Version:    payload.Version,
		Cells:      payload.Cells,
		StartTime:  payload.StartTime,
		EndTime:    payload.EndTime,
		AltitudeLo: payload.AltitudeLo,
		AltitudeHi: payload.AltitudeHi,
		Writer:     locality.MustFromContext(ctx),
	}

	return updateISA(ctx, repo, isa)
}

func updateISA(ctx context.Context, repo repos.Repository, isa *ridmodels.IdentificationServiceArea) (*ISAResult, error) {
	old, err := repo.GetISA(ctx, isa.ID, true)
	switch {
	case err != nil:
		return nil, stacktrace.Propagate(err, "Error getting ISA")
	case old == nil:
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "ISA %s not found", isa.ID)
	case old.Owner != isa.Owner:
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
			"ISA owned by %s, but %s attempted to modify", old.Owner, isa.Owner)
	case !old.Version.Matches(isa.Version):
		return nil, stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
			"ISA currently at version %s but client specified %s", old.Version, isa.Version)
	}

	// Validate and perhaps correct StartTime and EndTime.
	if err := isa.AdjustTimeRange(timestamp.MustFromContext(ctx), old); err != nil {
		return nil, stacktrace.Propagate(err, "Error adjusting time range")
	}

	ret, err := repo.UpdateISA(ctx, isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error updating ISA")
	}

	// TODO steeling, we should change this to a Custom type, to obfuscate
	// some of these metrics and prevent us from doing the wrong thing.
	cells := s2.CellUnionFromUnion(old.Cells, isa.Cells)
	geo.Levelify(&cells)
	// UpdateNotificationIdxsInCells is done in the same transaction as the insert since they
	// are both modifying the store.
	subs, err := repo.UpdateNotificationIdxsInCells(ctx, cells)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error updating notification indices")
	}

	return &ISAResult{ISA: ret, Subscriptions: subs}, nil
}
