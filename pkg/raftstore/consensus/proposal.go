package consensus

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/raftstore/consensus/proposalpb"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
	"google.golang.org/protobuf/proto"
)

type EntryCommit struct {
	Prop Proposal
	Done chan ProposalResult

	SnapshotData []byte
}

type Proposal struct {
	ID          string
	NodeID      uint64
	Timestamp   time.Time
	RequestType string
	Value       []byte
	// ReadOnly proposals are served ReadIndex. They are not replicated in the Raft log.
	// This flag must not be set for proposals that modify state, as it would break linearizability.
	ReadOnly bool
}

func (c *Consensus) newProposal(ctx context.Context, requestType string, value []byte, readOnly bool) Proposal {
	timestamp := timestamp.MustFromContext(ctx)

	return Proposal{
		ID:          uuid.NewString(),
		NodeID:      c.nodeID,
		Timestamp:   timestamp.UTC(),
		RequestType: requestType,
		Value:       value,
		ReadOnly:    readOnly,
	}
}

func (p Proposal) encode() ([]byte, error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid proposal ID %q", p.ID)
	}
	if p.RequestType == "" {
		return nil, stacktrace.NewError("missing request type")
	}
	timestampNanos := p.Timestamp.UnixNano()
	if !time.Unix(0, timestampNanos).Equal(p.Timestamp) {
		return nil, stacktrace.NewError("timestamp %s cannot be represented as Unix nanoseconds", p.Timestamp)
	}

	buf, err := proto.Marshal(&proposalpb.Proposal{
		Id:                 id[:],
		NodeId:             p.NodeID,
		TimestampUnixNanos: timestampNanos,
		RequestType:        p.RequestType,
		Value:              p.Value,
		ReadOnly:           p.ReadOnly,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal proposal")
	}
	return buf, nil
}

func decodeProposal(data []byte) (Proposal, error) {
	var pb proposalpb.Proposal
	if err := proto.Unmarshal(data, &pb); err != nil {
		return Proposal{}, stacktrace.Propagate(err, "failed to unmarshal proposal")
	}

	id, err := uuid.FromBytes(pb.GetId())
	if err != nil {
		return Proposal{}, stacktrace.Propagate(err, "invalid proposal ID")
	}
	if pb.GetRequestType() == "" {
		return Proposal{}, stacktrace.NewError("missing request type")
	}

	return Proposal{
		ID:          id.String(),
		NodeID:      pb.GetNodeId(),
		Timestamp:   time.Unix(0, pb.GetTimestampUnixNanos()).UTC(),
		RequestType: pb.GetRequestType(),
		Value:       pb.GetValue(),
		ReadOnly:    pb.GetReadOnly(),
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
