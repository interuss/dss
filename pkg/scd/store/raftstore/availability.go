package raftstore

import (
	"context"
	"encoding/json"

	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/stacktrace"
)

const (
	getUssAvailability    consensus.RequestType[*scdmodels.UssAvailabilityStatus] = "getUssAvailability"
	upsertUssAvailability consensus.RequestType[*scdmodels.UssAvailabilityStatus] = "upsertUssAvailability"
)

func (r *repo) GetUssAvailability(ctx context.Context, id dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error) {
	buf, err := json.Marshal(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, getUssAvailability, buf, true)
}

func (r *repo) UpsertUssAvailability(ctx context.Context, ussa *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	buf, err := json.Marshal(ussa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}

	return r.consensus.HandleClientRequest(ctx, upsertUssAvailability, buf, false)
}
