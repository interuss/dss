package operations

import (
	"context"

	ridv1 "github.com/interuss/dss/pkg/api/ridv1"
	ridv2 "github.com/interuss/dss/pkg/api/ridv2"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/locality"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	apiv1 "github.com/interuss/dss/pkg/rid/models/api/v1"
	apiv2 "github.com/interuss/dss/pkg/rid/models/api/v2"
	"github.com/interuss/dss/pkg/rid/repos"
	dssstore "github.com/interuss/dss/pkg/store"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

// ISAResult bundles the affected ISA together with the Subscriptions whose notification
// index changed as a result of the operation. Both are produced within the same store
// transaction, so callers need both from a single result.
type ISAResult struct {
	ISA           *ridmodels.IdentificationServiceArea
	Subscriptions []*ridmodels.Subscription
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
		Decode:  dssstore.DecodeJSON[*ridv1.CreateIdentificationServiceAreaRequest],
		Execute: executeInsertISA,
	}
	Registry[ridv2.CreateIdentificationServiceAreaOperationID] = dssstore.OperationHandler[repos.Repository]{
		Encode:  dssstore.EncodeJSON,
		Decode:  dssstore.DecodeJSON[*ridv2.CreateIdentificationServiceAreaRequest],
		Execute: executeInsertISA,
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
	var (
		rawID    string
		url      string
		clientID *string
		extents  *dssmodels.Volume4D
	)

	switch req := request.(type) {
	case *ridv1.CreateIdentificationServiceAreaRequest:
		if req.Body.FlightsUrl == "" {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required flightsURL")
		}
		if len(req.Body.Extents.SpatialVolume.Footprint.Vertices) == 0 {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing or malformed required extents")
		}
		requestExtents, err := apiv1.FromVolume4D(&req.Body.Extents)
		if err != nil {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err))
		}
		rawID, url, clientID, extents = string(req.Id), string(req.Body.FlightsUrl), req.Auth.ClientID, requestExtents

	case *ridv2.CreateIdentificationServiceAreaRequest:
		if req.Body.UssBaseUrl == "" {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required USS base URL")
		}
		requestExtents, err := apiv2.FromVolume4D(&req.Body.Extents)
		if err != nil {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err))
		}
		rawID, url, clientID, extents = string(req.Id), string(req.Body.UssBaseUrl), req.Auth.ClientID, requestExtents

	default:
		return nil, stacktrace.NewError("unexpected request type %T for operation %q", request, ridv2.CreateIdentificationServiceAreaOperationID)
	}

	id, err := dssmodels.IDFromString(rawID)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:     id,
		Owner:  dssmodels.Owner(*clientID),
		URL:    url,
		Writer: locality.MustFromContext(ctx),
	}
	if err := isa.SetExtents(extents); err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents")
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
