package memstore

import (
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/pb/auxpb/snapshotpb"
	"github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/stacktrace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *repo) GetSnapshot() (*raftpb.Snapshot, error) {
	snap := &snapshotpb.Snapshot{
		Participants: make([]*snapshotpb.Participant, 0, len(r.state.Participants)),
		Heartbeats:   make([]*snapshotpb.Heartbeat, 0, len(r.state.Heartbeats)),
	}
	for loc, p := range r.state.Participants {
		snap.Participants = append(snap.Participants, &snapshotpb.Participant{
			Locality:       string(loc),
			PublicEndpoint: p.PublicEndpoint,
			UpdatedAt:      timestamppb.New(p.UpdatedAt),
		})
	}
	for key, hb := range r.state.Heartbeats {
		snap.Heartbeats = append(snap.Heartbeats, &snapshotpb.Heartbeat{
			Locality:                    string(key.Locality),
			Source:                      key.Source,
			Timestamp:                   dssmodels.TimeToProto(hb.Timestamp),
			NextHeartbeatExpectedBefore: dssmodels.TimeToProto(hb.NextHeartbeatExpectedBefore),
			Reporter:                    hb.Reporter,
		})
	}

	return &raftpb.Snapshot{AppSnapshot: &raftpb.Snapshot_AuxSnapshot{AuxSnapshot: snap}}, nil
}

func (r *repo) RestoreFromSnapshot(s *raftpb.Snapshot) error {
	snap := s.GetAuxSnapshot()
	if snap == nil {
		return stacktrace.NewError("Expected an aux snapshot, got %T", s.GetAppSnapshot())
	}

	restored := state{
		Participants: make(map[locality]*participant, len(snap.GetParticipants())),
		Heartbeats:   make(map[heartbeatKey]*heartbeat, len(snap.GetHeartbeats())),
	}
	for _, pb := range snap.GetParticipants() {
		restored.Participants[locality(pb.GetLocality())] = &participant{
			PublicEndpoint: pb.GetPublicEndpoint(),
			UpdatedAt:      pb.GetUpdatedAt().AsTime(),
		}
	}
	for _, pb := range snap.GetHeartbeats() {
		key := heartbeatKey{Locality: locality(pb.GetLocality()), Source: pb.GetSource()}
		restored.Heartbeats[key] = &heartbeat{
			Timestamp:                   dssmodels.TimeFromProto(pb.GetTimestamp()),
			NextHeartbeatExpectedBefore: dssmodels.TimeFromProto(pb.GetNextHeartbeatExpectedBefore()),
			Reporter:                    pb.GetReporter(),
		}
	}

	r.state = restored
	return nil
}
