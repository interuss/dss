package consensus

import (
	"context"
	"fmt"
	"sync"
	"time"
	"uuid"

	"github.com/interuss/dss/pkg/pb/raftpb"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EntryCommit struct {
	Prop Proposal
	// ReadOnly proposals are served ReadIndex. They are not replicated in the Raft log.
	// This flag must not be set for proposals that modify state, as it would break linearizability.
	ReadOnly bool
	Done     chan ProposalResult

	Snapshot *raftpb.Snapshot
}

type Proposal struct {
	ID     string
	NodeID uint64
	// Timestamp is always in UTC so that it is identical on all the nodes.
	// It is encoded as a protobuf Timestamp and decoding always yields a UTC time.
	Timestamp time.Time
	Payload   *raftpb.Payload
}

func (c *Consensus) newProposal(ctx context.Context, payload *raftpb.Payload) Proposal {
	timestamp := timestamp.MustFromContext(ctx)

	return Proposal{
		ID:        uuid.New().String(),
		NodeID:    c.nodeID,
		Timestamp: timestamp.UTC(),
		Payload:   payload,
	}
}

// Name identifies the kind of request carried by the proposal, for logging purposes.
func (p Proposal) Name() string {
	switch payload := p.Payload.GetPayload().(type) {
	case *raftpb.Payload_AuxPayload:
		return fmt.Sprintf("%T", payload.AuxPayload.GetPayload())
	case *raftpb.Payload_RidPayload:
		return fmt.Sprintf("%T", payload.RidPayload.GetPayload())
	case *raftpb.Payload_ScdPayload:
		return fmt.Sprintf("%T", payload.ScdPayload.GetPayload())
	default:
		return fmt.Sprintf("%T", payload)
	}
}

func (p Proposal) encode() ([]byte, error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid proposal ID %q", p.ID)
	}
	if p.Payload.GetPayload() == nil {
		return nil, stacktrace.NewError("missing payload")
	}
	ts := timestamppb.New(p.Timestamp)
	if err := ts.CheckValid(); err != nil {
		return nil, stacktrace.Propagate(err, "invalid timestamp %s", p.Timestamp)
	}

	buf, err := proto.Marshal(&raftpb.Proposal{
		Id:        id[:],
		NodeId:    p.NodeID,
		Timestamp: ts,
		Payload:   p.Payload,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal proposal")
	}
	return buf, nil
}

func decodeProposal(data []byte) (Proposal, error) {
	var pb raftpb.Proposal
	if err := proto.Unmarshal(data, &pb); err != nil {
		return Proposal{}, stacktrace.Propagate(err, "failed to unmarshal proposal")
	}

	if len(pb.GetId()) != len(uuid.UUID{}) {
		return Proposal{}, stacktrace.NewError("invalid proposal ID length %d", len(pb.GetId()))
	}
	if pb.GetPayload().GetPayload() == nil {
		return Proposal{}, stacktrace.NewError("missing payload")
	}
	if err := pb.GetTimestamp().CheckValid(); err != nil {
		return Proposal{}, stacktrace.Propagate(err, "invalid proposal timestamp")
	}

	return Proposal{
		ID:        uuid.UUID(pb.GetId()).String(),
		NodeID:    pb.GetNodeId(),
		Timestamp: pb.GetTimestamp().AsTime(),
		Payload:   pb.GetPayload(),
	}, nil
}

type ProposalResult struct {
	Result any
	Error  error
}

type proposalsTracker struct {
	sync.Mutex
	pending map[string]chan ProposalResult
}

func newProposalsTracker() *proposalsTracker {
	return &proposalsTracker{
		pending: make(map[string]chan ProposalResult),
	}
}

func (p *proposalsTracker) track(id string) chan ProposalResult {
	p.Lock()
	defer p.Unlock()

	applied := make(chan ProposalResult, 1)
	p.pending[id] = applied
	return applied
}

func (p *proposalsTracker) untrack(id string, result ProposalResult) {
	p.Lock()
	defer p.Unlock()

	applied, ok := p.pending[id]
	if !ok {
		return
	}

	applied <- result
	delete(p.pending, id)
}
