package consensus

import (
	proposalpb "github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/stacktrace"
	"google.golang.org/protobuf/proto"
)

// snapshot wraps the application snapshot with the membership state that the raft library does not track itself,
// so that a node catching up via a snapshot fully recovers it, including which IDs must never be reused.
type snapshot struct {
	// Members maps the ID of every member of the group to its membership properties.
	Members memberMap
	// Removed maps the ID of every node permanently removed from the group to its membership properties when it was removed.
	Removed memberMap
	// AppSnapshot holds the application snapshot.
	AppSnapshot *proposalpb.Snapshot
}

// encodeSnapshot encodes the application snapshot with the given membership state into snapshot data.
func encodeSnapshot(members, removed memberMap, getAppSnapshot func() (*proposalpb.Snapshot, error)) ([]byte, error) {
	env, err := getAppSnapshot()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get application snapshot")
	}
	if env.GetAppSnapshot() == nil {
		return nil, stacktrace.NewError("missing application snapshot")
	}

	env.Members = members.toProto()
	env.Removed = removed.toProto()

	data, err := proto.Marshal(env)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to encode snapshot")
	}
	return data, nil
}

// decodeSnapshot decodes the snapshot data produced by encodeSnapshot.
func decodeSnapshot(data []byte) (snapshot, error) {
	var env proposalpb.Snapshot
	if err := proto.Unmarshal(data, &env); err != nil {
		return snapshot{}, stacktrace.Propagate(err, "failed to decode snapshot")
	}
	if env.GetAppSnapshot() == nil {
		return snapshot{}, stacktrace.NewError("missing application snapshot")
	}
	snap := snapshot{AppSnapshot: &env}
	snap.Members.fromProto(env.GetMembers())
	snap.Removed.fromProto(env.GetRemoved())
	return snap, nil
}
