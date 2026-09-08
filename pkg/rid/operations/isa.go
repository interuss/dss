package operations

import (
	"context"

	ridv1 "github.com/interuss/dss/pkg/api/ridv1"
	ridv2 "github.com/interuss/dss/pkg/api/ridv2"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	dssstore "github.com/interuss/dss/pkg/store"
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
