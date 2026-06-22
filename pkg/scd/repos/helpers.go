package repos

import (
	"context"

	"github.com/golang/geo/s2"

	restapi "github.com/interuss/dss/pkg/api/scdv1"

	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

type ValidOIRParams struct {
	ID                   dssmodels.ID
	OVN                  scdmodels.OVN
	NewOVN               scdmodels.OVN
	State                scdmodels.OperationalIntentState
	UExtent              *dssmodels.Volume4D
	Cells                s2.CellUnion
	SubscriptionID       dssmodels.ID
	USSBaseURL           string
	ImplicitSubscription struct {
		Requested      bool
		ID             dssmodels.ID
		BaseURL        string
		ForConstraints bool
	}
	Key map[scdmodels.OVN]bool
}

func (vp *ValidOIRParams) ToOIR(manager dssmodels.Manager, attachedSub *scdmodels.Subscription, version scdmodels.VersionNumber, pastOVNs []scdmodels.OVN) *scdmodels.OperationalIntent {
	// For OIR's in the accepted state, we may not have a attachedSub available,
	// in such cases the attachedSub ID on scdmodels.OperationalIntent will be nil
	// and will be replaced with the 'NullV4UUID' when sent over to a client.
	var subID *dssmodels.ID
	if attachedSub != nil {
		// Note: do _not_ use vp.subscriptionID here, as it may be empty
		subID = &attachedSub.ID
	}
	return &scdmodels.OperationalIntent{
		ID:       vp.ID,
		Manager:  manager,
		Version:  version,
		OVN:      vp.NewOVN, // non-empty only if the USS has requested an OVN
		PastOVNs: pastOVNs,

		StartTime:     vp.UExtent.StartTime,
		EndTime:       vp.UExtent.EndTime,
		AltitudeLower: vp.UExtent.SpatialVolume.AltitudeLo,
		AltitudeUpper: vp.UExtent.SpatialVolume.AltitudeHi,
		Cells:         vp.Cells,

		USSBaseURL:     vp.USSBaseURL,
		SubscriptionID: subID,
		State:          vp.State,
	}
}

// SubscriptionIsImplicitAndOnlyAttachedToOIR will check if:
// - the subscription is defined and is implicit
// - the subscription is attached to the specified operational intent
// - the subscription is not attached to any other operational intent
//
// This is to be used in contexts where an implicit subscription may need to be cleaned up: if true is returned,
// the subscription can be safely removed after the operational intent is deleted or attached to another subscription.
//
// NOTE: this should eventually be pushed down the datastore as part of the queries being executed in the callers of this method.
//
//	See https://github.com/interuss/dss/issues/1059 for more details
func SubscriptionIsImplicitAndOnlyAttachedToOIR(ctx context.Context, r Repository, oirID dssmodels.ID, subscription *scdmodels.Subscription) (bool, error) {
	if subscription == nil {
		return false, nil
	}
	if !subscription.ImplicitSubscription {
		return false, nil
	}
	// Get the Subscription's dependent OperationalIntents
	dependentOps, err := r.GetDependentOperationalIntents(ctx, subscription.ID)
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not find dependent OperationalIntents")
	}
	if len(dependentOps) == 0 {
		return false, stacktrace.NewError("An implicit Subscription had no dependent OperationalIntents")
	} else if len(dependentOps) == 1 && dependentOps[0] == oirID {
		return true, nil
	}
	return false, nil
}

// ValidateUpsertRequestAgainstPreviousOIR checks that the client requesting an OIR upsert has the necessary permissions and that the request is valid.
// On success, the version of the OIR is returned:
//   - upon initial creation (if no previous OIR exists), it is 0
//   - otherwise, it is the version of the previous OIR
func ValidateUpsertRequestAgainstPreviousOIR(
	requestingManager dssmodels.Manager,
	providedOVN scdmodels.OVN,
	previousOIR *scdmodels.OperationalIntent,
) error {

	if previousOIR != nil {
		if previousOIR.Manager != requestingManager {
			return stacktrace.NewErrorWithCode(dsserr.PermissionDenied,
				"OperationalIntent owned by %s, but %s attempted to modify", previousOIR.Manager, requestingManager)
		}
		if previousOIR.OVN != providedOVN {
			return stacktrace.NewErrorWithCode(dsserr.VersionMismatch,
				"Current version is %s but client specified version %s", previousOIR.OVN, providedOVN)
		}

		return nil
	}

	if providedOVN != "" {
		return stacktrace.NewErrorWithCode(dsserr.NotFound, "OperationalIntent does not exist and therefore is not version %s", providedOVN)
	}

	return nil
}

// CreateAndStoreNewImplicitSubscription will create a brand new implicit subscription based on the provided parameters,
// store it and return it.
func CreateAndStoreNewImplicitSubscription(ctx context.Context, r Repository, manager dssmodels.Manager, validParams *ValidOIRParams) (*scdmodels.Subscription, error) {
	subToUpsert := scdmodels.Subscription{
		ID:                          validParams.ImplicitSubscription.ID,
		Manager:                     manager,
		StartTime:                   validParams.UExtent.StartTime,
		EndTime:                     validParams.UExtent.EndTime,
		AltitudeLo:                  validParams.UExtent.SpatialVolume.AltitudeLo,
		AltitudeHi:                  validParams.UExtent.SpatialVolume.AltitudeHi,
		Cells:                       validParams.Cells,
		USSBaseURL:                  validParams.ImplicitSubscription.BaseURL,
		NotifyForOperationalIntents: true,
		NotifyForConstraints:        validParams.ImplicitSubscription.ForConstraints,
		ImplicitSubscription:        true,
	}

	return r.UpsertSubscription(ctx, &subToUpsert)
}

// ComputeNotificationVolume computes the volume that needs to be queried for subscriptions
// given the requested extent and the (possibly nil) previous operational intent.
// The returned volume is either the union of the requested extent and the previous OIR's extent, or just the requested extent
// if the previous OIR is nil.
func ComputeNotificationVolume(
	previousOIR *scdmodels.OperationalIntent,
	requestedExtent *dssmodels.Volume4D) (*dssmodels.Volume4D, error) {

	if previousOIR == nil {
		return requestedExtent, nil
	}

	// Compute total affected Volume4D for notification purposes
	oldVolume := &dssmodels.Volume4D{
		StartTime: previousOIR.StartTime,
		EndTime:   previousOIR.EndTime,
		SpatialVolume: &dssmodels.Volume3D{
			AltitudeHi: previousOIR.AltitudeUpper,
			AltitudeLo: previousOIR.AltitudeLower,
			Footprint: dssmodels.GeometryFunc(func() (s2.CellUnion, error) {
				return previousOIR.Cells, nil
			}),
		},
	}
	notifyVolume, err := dssmodels.UnionVolumes4D(requestedExtent, oldVolume)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error constructing 4D volumes union")
	}

	return notifyVolume, nil
}

// GetRelevantSubscriptionsAndIncrementIndices retrieves the subscriptions relevant to the passed volume and increments their notification indices
// before returning them.
func GetRelevantSubscriptionsAndIncrementIndices(
	ctx context.Context,
	r Repository,
	notifyVolume *dssmodels.Volume4D,
) (Subscriptions, error) {

	// Find the Subscriptions interested in OperationalIntents and increment their
	// notification indices
	subs, err := r.IncrementNotificationIndicesForOperationalIntents(ctx, notifyVolume)

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to increment notification indices of relevant subscriptions")
	}

	return subs, nil
}

// ValidateKeyAndProvideConflictResponse ensures that the provided key contains all the necessary OVNs relevant for the area covered by the OperationalIntent.
// - If all required keys are provided, (nil, nil) will be returned.
// - If keys are missing, the conflict response to be sent back as well as an error with the dsserr.MissingOVNs code will be returned.
// - In case of any other error, (nil, error) will be returned.
func ValidateKeyAndProvideConflictResponse(
	ctx context.Context,
	r Repository,
	requestingManager dssmodels.Manager,
	params *ValidOIRParams,
	attachedSubscription *scdmodels.Subscription,
) (*restapi.AirspaceConflictResponse, error) {

	// Identify OperationalIntents missing from the key
	var missingOps []*scdmodels.OperationalIntent
	relevantOps, err := r.SearchOperationalIntents(ctx, params.UExtent)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to SearchOperations")
	}
	for _, relevantOp := range relevantOps {
		_, ok := params.Key[relevantOp.OVN]
		// Note: The OIR being mutated does not need to be specified in the key:
		if !ok && relevantOp.RequiresKey() && relevantOp.ID != params.ID {
			missingOps = append(missingOps, relevantOp)
		}
	}

	// Identify Constraints missing from the key
	var missingConstraints []*scdmodels.Constraint
	if attachedSubscription != nil && attachedSubscription.NotifyForConstraints {
		constraints, err := r.SearchConstraints(ctx, params.UExtent)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Unable to SearchConstraints")
		}
		for _, relevantConstraint := range constraints {
			if _, ok := params.Key[relevantConstraint.OVN]; !ok {
				missingConstraints = append(missingConstraints, relevantConstraint)
			}
		}
	}

	// If the client is missing some OVNs, provide the pointers to the
	// information they need
	if len(missingOps) > 0 || len(missingConstraints) > 0 {
		msg := "Current OVNs not provided for one or more OperationalIntents or Constraints"
		responseConflict := &restapi.AirspaceConflictResponse{Message: &msg}

		if len(missingOps) > 0 {
			responseConflict.MissingOperationalIntents = new([]restapi.OperationalIntentReference)
			for _, missingOp := range missingOps {
				p := missingOp.ToRest()
				// We scrub the OVNs of entities not owned by the requesting manager to make sure
				// they have really contacted the managing USS
				if missingOp.Manager != requestingManager {
					noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
					p.Ovn = &noOvnPhrase
				}
				*responseConflict.MissingOperationalIntents = append(*responseConflict.MissingOperationalIntents, *p)
			}
		}

		if len(missingConstraints) > 0 {
			responseConflict.MissingConstraints = new([]restapi.ConstraintReference)
			for _, missingConstraint := range missingConstraints {
				c := missingConstraint.ToRest()
				// We scrub the OVNs of entities not owned by the requesting manager to make sure
				// they have really contacted the managing USS
				if missingConstraint.Manager != requestingManager {
					noOvnPhrase := restapi.EntityOVN(scdmodels.NoOvnPhrase)
					c.Ovn = &noOvnPhrase
				}
				*responseConflict.MissingConstraints = append(*responseConflict.MissingConstraints, *c)
			}
		}

		return responseConflict, stacktrace.NewErrorWithCode(dsserr.MissingOVNs, "Missing OVNs: %v", msg)
	}

	return nil, nil
}

// EnsureSubscriptionCoversOIR ensures that the subscription covers the requested geo-temporal extent, extending it if both possible and required,
// or failing otherwise.
// After this method returns successfully, the subscription will cover the requested geo-temporal extent.
func EnsureSubscriptionCoversOIR(ctx context.Context, r Repository, sub *scdmodels.Subscription, params *ValidOIRParams) (*scdmodels.Subscription, error) {

	updateSub := false
	if sub.StartTime != nil && sub.StartTime.After(*params.UExtent.StartTime) {
		if sub.ImplicitSubscription {
			sub.StartTime = params.UExtent.StartTime
			updateSub = true
		} else {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not begin until after the OperationalIntent starts")
		}
	}
	if sub.EndTime != nil && sub.EndTime.Before(*params.UExtent.EndTime) {
		if sub.ImplicitSubscription {
			sub.EndTime = params.UExtent.EndTime
			updateSub = true
		} else {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription ends before the OperationalIntent ends")
		}
	}
	if !sub.Cells.Contains(params.Cells) {
		if sub.ImplicitSubscription {
			sub.Cells = s2.CellUnionFromUnion(sub.Cells, params.Cells)
			updateSub = true
		} else {
			return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Subscription does not cover entire spatial area of the OperationalIntent")
		}
	}
	if updateSub {
		upsertedSub, err := r.UpsertSubscription(ctx, sub)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to update existing Subscription")
		}
		return upsertedSub, nil
	}

	return sub, nil
}

func MakeSubscribersToNotify(subscriptions []*scdmodels.Subscription) []restapi.SubscriberToNotify {
	result := []restapi.SubscriberToNotify{}

	subscriptionsByURL := map[string][]restapi.SubscriptionState{}
	for _, sub := range subscriptions {
		subState := restapi.SubscriptionState{
			SubscriptionId:    restapi.SubscriptionID(sub.ID.String()),
			NotificationIndex: restapi.SubscriptionNotificationIndex(sub.NotificationIndex),
		}
		subscriptionsByURL[sub.USSBaseURL] = append(subscriptionsByURL[sub.USSBaseURL], subState)
	}
	for url, states := range subscriptionsByURL {
		result = append(result, restapi.SubscriberToNotify{
			UssBaseUrl:    restapi.SubscriptionUssBaseURL(url),
			Subscriptions: states,
		})
	}

	return result
}
