package memstore

import (
	"bytes"
	"encoding/gob"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/stacktrace"
)

const snapshotVersion = 1

type snapshotEnvelope struct {
	Version int
	State   state
}

func (r *repo) GetSnapshot() ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snapshotEnvelope{Version: snapshotVersion, State: r.state}); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to encode memstore snapshot")
	}
	return buf.Bytes(), nil
}

func (r *repo) RestoreFromSnapshot(data []byte) error {
	var env snapshotEnvelope
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&env); err != nil {
		return stacktrace.Propagate(err, "Failed to decode memstore snapshot")
	}
	if env.Version != snapshotVersion {
		return stacktrace.NewError("Unsupported memstore snapshot version %d, expected %d", env.Version, snapshotVersion)
	}
	r.state = env.State
	// gob decodes an empty map as nil; re-initialize to keep the repo writable.
	if r.state.Constraints == nil {
		r.state.Constraints = map[dssmodels.ID]*constraintRecord{}
	}
	if r.state.Subscriptions == nil {
		r.state.Subscriptions = map[dssmodels.ID]*subscriptionRecord{}
	}
	if r.state.OperationalIntents == nil {
		r.state.OperationalIntents = map[dssmodels.ID]*operationalIntentRecord{}
	}
	if r.state.Availabilities == nil {
		r.state.Availabilities = map[dssmodels.Manager]*availabilityRecord{}
	}
	return nil
}
