package memstore

import (
	dssmodels "github.com/interuss/dss/pkg/models"
	modelspb "github.com/interuss/dss/pkg/pb"
	"github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/dss/pkg/pb/scdpb/snapshotpb"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *repo) GetSnapshot() (*raftpb.Snapshot, error) {
	snap := &snapshotpb.Snapshot{
		Constraints:        make(map[string]*snapshotpb.Constraint, len(r.state.Constraints)),
		Subscriptions:      make(map[string]*snapshotpb.Subscription, len(r.state.Subscriptions)),
		OperationalIntents: make(map[string]*snapshotpb.OperationalIntent, len(r.state.OperationalIntents)),
		Availabilities:     make(map[string]*snapshotpb.Availability, len(r.state.Availabilities)),
	}
	for id, rec := range r.state.Constraints {
		snap.Constraints[id.String()] = &snapshotpb.Constraint{
			Manager:    rec.Manager.String(),
			Version:    int32(rec.Version),
			UssBaseUrl: rec.USSBaseURL,
			Volume:     rec.Volume.ToProto(),
			UpdatedAt:  timestamppb.New(rec.UpdatedAt),
		}
	}
	for id, rec := range r.state.Subscriptions {
		snap.Subscriptions[id.String()] = &snapshotpb.Subscription{
			Manager:                     rec.Manager.String(),
			NotificationIndex:           int64(rec.NotificationIndex),
			UssBaseUrl:                  rec.USSBaseURL,
			NotifyForOperationalIntents: rec.NotifyForOperationalIntents,
			NotifyForConstraints:        rec.NotifyForConstraints,
			ImplicitSubscription:        rec.ImplicitSubscription,
			StartTime:                   timestamppb.New(rec.StartTime),
			EndTime:                     timestamppb.New(rec.EndTime),
			Cells:                       dssmodels.CellUnionToProto(rec.Cells),
			UpdatedAt:                   timestamppb.New(rec.UpdatedAt),
		}
	}
	for id, rec := range r.state.OperationalIntents {
		state, err := rec.State.ToProto()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to encode state of operational intent %s", id)
		}
		var subscriptionID *string
		if rec.SubscriptionID != nil {
			subscriptionID = new(rec.SubscriptionID.String())
		}
		snap.OperationalIntents[id.String()] = &snapshotpb.OperationalIntent{
			Manager:         rec.Manager.String(),
			Version:         int32(rec.Version),
			State:           state,
			UssBaseUrl:      rec.USSBaseURL,
			SubscriptionId:  subscriptionID,
			UssRequestedOvn: scdmodels.OVN(rec.USSRequestedOVN).ToProto(),
			PastOvns:        scdmodels.OVNsToProto(rec.PastOVNs),
			UpdatedAt:       timestamppb.New(rec.UpdatedAt),
			Volume:          rec.CellsVolume4D.ToProto(),
		}
	}
	for uss, rec := range r.state.Availabilities {
		availability, err := rec.Availability.ToProto()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to encode availability of USS %s", uss)
		}
		snap.Availabilities[uss.String()] = &snapshotpb.Availability{
			Availability: availability,
			UpdatedAt:    timestamppb.New(rec.UpdatedAt),
		}
	}

	if r.state.EvictionConfig != nil {
		snap.EvictionConfig = r.state.EvictionConfig.ToProto()
	}

	return &raftpb.Snapshot{AppSnapshot: &raftpb.Snapshot_ScdSnapshot{ScdSnapshot: snap}}, nil
}

func (r *repo) RestoreFromSnapshot(s *raftpb.Snapshot) error {
	snap := s.GetScdSnapshot()
	if snap == nil {
		return stacktrace.NewError("Expected an SCD snapshot, got %T", s.GetAppSnapshot())
	}

	restored := state{
		Constraints:        make(map[dssmodels.ID]*constraintRecord, len(snap.GetConstraints())),
		Subscriptions:      make(map[dssmodels.ID]*subscriptionRecord, len(snap.GetSubscriptions())),
		OperationalIntents: make(map[dssmodels.ID]*operationalIntentRecord, len(snap.GetOperationalIntents())),
		Availabilities:     make(map[dssmodels.Manager]*availabilityRecord, len(snap.GetAvailabilities())),
	}
	for key, pb := range snap.GetConstraints() {
		id := dssmodels.ID(key)
		restored.Constraints[id] = &constraintRecord{
			ID:         id,
			Manager:    dssmodels.Manager(pb.GetManager()),
			Version:    scdmodels.VersionNumber(pb.GetVersion()),
			USSBaseURL: pb.GetUssBaseUrl(),
			Volume:     volumeFromProto(pb.GetVolume()),
			UpdatedAt:  pb.GetUpdatedAt().AsTime(),
		}
	}
	for key, pb := range snap.GetSubscriptions() {
		id := dssmodels.ID(key)
		restored.Subscriptions[id] = &subscriptionRecord{
			ID:                          id,
			Manager:                     dssmodels.Manager(pb.GetManager()),
			NotificationIndex:           int(pb.GetNotificationIndex()),
			USSBaseURL:                  pb.GetUssBaseUrl(),
			NotifyForOperationalIntents: pb.GetNotifyForOperationalIntents(),
			NotifyForConstraints:        pb.GetNotifyForConstraints(),
			ImplicitSubscription:        pb.GetImplicitSubscription(),
			StartTime:                   pb.GetStartTime().AsTime(),
			EndTime:                     pb.GetEndTime().AsTime(),
			Cells:                       dssmodels.CellUnionFromProto(pb.GetCells()),
			UpdatedAt:                   pb.GetUpdatedAt().AsTime(),
		}
	}
	for key, pb := range snap.GetOperationalIntents() {
		id := dssmodels.ID(key)
		var subscriptionID *dssmodels.ID
		if pb.SubscriptionId != nil {
			subscriptionID = new(dssmodels.ID(pb.GetSubscriptionId()))
		}
		var oiState scdmodels.OperationalIntentState
		if err := oiState.FromProto(pb.GetState()); err != nil {
			return stacktrace.Propagate(err, "Failed to decode state of operational intent %s", id)
		}
		var requestedOVN scdmodels.OVN
		if err := requestedOVN.FromProto(pb.GetUssRequestedOvn()); err != nil {
			return stacktrace.Propagate(err, "Failed to decode requested OVN of operational intent %s", id)
		}
		pastOVNs, err := scdmodels.OVNsFromProto(pb.GetPastOvns())
		if err != nil {
			return stacktrace.Propagate(err, "Failed to decode past OVNs of operational intent %s", id)
		}
		restored.OperationalIntents[id] = &operationalIntentRecord{
			ID:              id,
			Manager:         dssmodels.Manager(pb.GetManager()),
			Version:         scdmodels.VersionNumber(pb.GetVersion()),
			State:           oiState,
			USSBaseURL:      pb.GetUssBaseUrl(),
			SubscriptionID:  subscriptionID,
			USSRequestedOVN: requestedOVN.String(),
			PastOVNs:        pastOVNs,
			UpdatedAt:       pb.GetUpdatedAt().AsTime(),
			CellsVolume4D:   volumeFromProto(pb.GetVolume()),
		}
	}
	for key, pb := range snap.GetAvailabilities() {
		uss := dssmodels.Manager(key)
		var availability scdmodels.UssAvailabilityState
		if err := availability.FromProto(pb.GetAvailability()); err != nil {
			return stacktrace.Propagate(err, "Failed to decode availability of USS %s", uss)
		}
		restored.Availabilities[uss] = &availabilityRecord{
			Uss:          uss,
			Availability: availability,
			UpdatedAt:    pb.GetUpdatedAt().AsTime(),
		}
	}

	if snap.GetEvictionConfig() != nil {
		restored.EvictionConfig = new(dssmodels.EvictionConfig)
		if err := restored.EvictionConfig.FromProto(snap.GetEvictionConfig()); err != nil {
			return stacktrace.Propagate(err, "invalid eviction config in snapshot")
		}
	}

	r.state = restored
	return nil
}

func volumeFromProto(pb *modelspb.CellsVolume4D) *dssmodels.CellsVolume4D {
	if pb == nil {
		return nil
	}
	volume := new(dssmodels.CellsVolume4D)
	volume.FromProto(pb)
	return volume
}
