package consensus

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/interuss/dss/pkg/timestamp"
	"github.com/interuss/stacktrace"
)

type EntryCommit struct {
	Prop Proposal
	Done chan ProposalResult

	SnapshotData []byte
}

type Proposal struct {
	ID          string    `json:"id"`
	NodeID      uint64    `json:"node_id"`
	Timestamp   time.Time `json:"timestamp"`
	RequestType string    `json:"request_type"`
	Value       []byte    `json:"value"`
	// ReadOnly proposals do not modify the state machine and,
	// therefore, do not need to be applied by nodes who did not initiate them.
	// TODO: This is a temporary solution. In the future, we will use ReadIndex
	// for read-only operations without needing to propose them to Raft.
	ReadOnly bool `json:"read_only"`
}

func (c *Consensus) newProposal(ctx context.Context, requestType string, payload any, readOnly bool) (Proposal, error) {
	timestamp, err := timestamp.RequestTimestampFromContext(ctx)
	if err != nil || timestamp.IsZero() {
		return Proposal{}, stacktrace.Propagate(err, "failed to get timestamp from context")
	}

	value, err := json.Marshal(payload)
	if err != nil {
		return Proposal{}, stacktrace.Propagate(err, "failed to serialize proposal payload")
	}

	return Proposal{
		ID:          uuid.NewString(),
		NodeID:      c.nodeID,
		Timestamp:   timestamp,
		RequestType: requestType,
		Value:       value,
		ReadOnly:    readOnly,
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

func (p *proposalsTracker) isPending(id string) bool {
	p.Lock()
	defer p.Unlock()

	_, ok := p.pending[id]
	return ok
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
