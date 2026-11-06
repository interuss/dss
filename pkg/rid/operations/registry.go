package operations

import (
	"github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/dss/pkg/pb/ridpb"
	"github.com/interuss/dss/pkg/rid/repos"
	dssstore "github.com/interuss/dss/pkg/store"
)

// Registry maps operation IDs to their handlers.
var Registry = map[string]dssstore.OperationHandler[repos.Repository]{}

// ProposalPayload wraps a rid payload in the payload of a proposal.
func ProposalPayload(payload *ridpb.Payload) *raftpb.Payload {
	return &raftpb.Payload{Payload: &raftpb.Payload_RidPayload{RidPayload: payload}}
}
