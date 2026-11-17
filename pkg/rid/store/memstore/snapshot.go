package memstore

import (
	"github.com/interuss/dss/pkg/memstore/utils"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/dss/pkg/pb/ridpb/snapshotpb"
	"github.com/interuss/stacktrace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *repo) GetSnapshot() (*raftpb.Snapshot, error) {
	snap := &snapshotpb.Snapshot{
		Isas:          make(map[string]*snapshotpb.ISA, len(r.state.ISAs)),
		Subscriptions: make(map[string]*snapshotpb.Subscription, len(r.state.Subscriptions)),
	}
	for id, rec := range r.state.ISAs {
		snap.Isas[id.String()] = &snapshotpb.ISA{
			Url:        rec.URL,
			Owner:      rec.Owner.String(),
			Cells:      dssmodels.CellUnionToProto(rec.Cells),
			StartTime:  timestamppb.New(rec.StartTime),
			EndTime:    timestamppb.New(rec.EndTime),
			AltitudeHi: utils.ClonePtr(rec.AltitudeHi),
			AltitudeLo: utils.ClonePtr(rec.AltitudeLo),
			Writer:     rec.Writer,
			UpdatedAt:  timestamppb.New(rec.UpdatedAt),
		}
	}
	for id, rec := range r.state.Subscriptions {
		snap.Subscriptions[id.String()] = &snapshotpb.Subscription{
			Url:               rec.URL,
			NotificationIndex: int32(rec.NotificationIndex),
			Owner:             rec.Owner.String(),
			Cells:             dssmodels.CellUnionToProto(rec.Cells),
			StartTime:         timestamppb.New(rec.StartTime),
			EndTime:           timestamppb.New(rec.EndTime),
			AltitudeHi:        utils.ClonePtr(rec.AltitudeHi),
			AltitudeLo:        utils.ClonePtr(rec.AltitudeLo),
			Writer:            rec.Writer,
			UpdatedAt:         timestamppb.New(rec.UpdatedAt),
		}
	}

	if r.state.EvictionConfig != nil {
		snap.EvictionConfig = r.state.EvictionConfig.ToProto()
	}

	return &raftpb.Snapshot{AppSnapshot: &raftpb.Snapshot_RidSnapshot{RidSnapshot: snap}}, nil
}

func (r *repo) RestoreFromSnapshot(s *raftpb.Snapshot) error {
	snap := s.GetRidSnapshot()
	if snap == nil {
		return stacktrace.NewError("Expected an RID snapshot, got %T", s.GetAppSnapshot())
	}

	restored := state{
		ISAs:          make(map[dssmodels.ID]*isaRecord, len(snap.GetIsas())),
		Subscriptions: make(map[dssmodels.ID]*subscriptionRecord, len(snap.GetSubscriptions())),
	}
	for key, pb := range snap.GetIsas() {
		id := dssmodels.ID(key)
		restored.ISAs[id] = &isaRecord{
			ID:         id,
			URL:        pb.GetUrl(),
			Owner:      dssmodels.Owner(pb.GetOwner()),
			Cells:      dssmodels.CellUnionFromProto(pb.GetCells()),
			StartTime:  pb.GetStartTime().AsTime(),
			EndTime:    pb.GetEndTime().AsTime(),
			AltitudeHi: utils.ClonePtr(pb.AltitudeHi),
			AltitudeLo: utils.ClonePtr(pb.AltitudeLo),
			Writer:     pb.GetWriter(),
			UpdatedAt:  pb.GetUpdatedAt().AsTime(),
		}
	}
	for key, pb := range snap.GetSubscriptions() {
		id := dssmodels.ID(key)
		restored.Subscriptions[id] = &subscriptionRecord{
			ID:                id,
			URL:               pb.GetUrl(),
			NotificationIndex: int(pb.GetNotificationIndex()),
			Owner:             dssmodels.Owner(pb.GetOwner()),
			Cells:             dssmodels.CellUnionFromProto(pb.GetCells()),
			StartTime:         pb.GetStartTime().AsTime(),
			EndTime:           pb.GetEndTime().AsTime(),
			AltitudeHi:        utils.ClonePtr(pb.AltitudeHi),
			AltitudeLo:        utils.ClonePtr(pb.AltitudeLo),
			Writer:            pb.GetWriter(),
			UpdatedAt:         pb.GetUpdatedAt().AsTime(),
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
