package consensus

import (
	"context"
	"encoding/binary"
	"math"
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

type Header struct {
	ID        uuid.UUID
	NodeID    uint64
	Timestamp time.Time
	// ReadOnly proposals are served ReadIndex. They are not replicated in the Raft log.
	// This flag must not be set for proposals that modify state, as it would break linearizability.
	ReadOnly    bool
	RequestType string
}

type Proposal struct {
	Header
	Value []byte
}

func (c *Consensus) newProposal(ctx context.Context, requestType string, value []byte, readOnly bool) Proposal {
	timestamp := timestamp.MustFromContext(ctx)

	return Proposal{
		Header: Header{
			ID:          uuid.New(),
			NodeID:      c.nodeID,
			Timestamp:   timestamp.UTC(),
			ReadOnly:    readOnly,
			RequestType: requestType,
		},
		Value: value,
	}
}

const proposalEncodingVersion byte = 1

const (
	headerFlagReadOnly byte = 1 << iota

	headerKnownFlags = headerFlagReadOnly
)

const headerFixedSize = 16 + 8 + 8 + 1 + 1

func (h Header) encodedSize() int {
	return headerFixedSize + len(h.RequestType)
}

func (h Header) appendEncoded(buf []byte) ([]byte, error) {
	if h.RequestType == "" {
		return nil, stacktrace.NewError("missing request type")
	}
	if len(h.RequestType) > math.MaxUint8 {
		return nil, stacktrace.NewError("request type %q exceeds %d bytes", h.RequestType, math.MaxUint8)
	}
	timestampNanos := h.Timestamp.UnixNano()
	if !time.Unix(0, timestampNanos).Equal(h.Timestamp) {
		return nil, stacktrace.NewError("timestamp %s cannot be represented as Unix nanoseconds", h.Timestamp)
	}

	var flags byte
	if h.ReadOnly {
		flags |= headerFlagReadOnly
	}

	buf = append(buf, h.ID[:]...)
	buf = binary.BigEndian.AppendUint64(buf, h.NodeID)
	buf = binary.BigEndian.AppendUint64(buf, uint64(timestampNanos))
	buf = append(buf, flags, byte(len(h.RequestType)))
	buf = append(buf, h.RequestType...)
	return buf, nil
}

func decodeHeader(data []byte) (Header, int, error) {
	if len(data) < headerFixedSize {
		return Header{}, 0, stacktrace.NewError("header too short: %d bytes", len(data))
	}

	flags := data[32]
	if flags&^headerKnownFlags != 0 {
		return Header{}, 0, stacktrace.NewError("unknown header flags %08b", flags)
	}

	size := headerFixedSize + int(data[33])
	if len(data) < size {
		return Header{}, 0, stacktrace.NewError("header too short for its request type: %d bytes, expected %d", len(data), size)
	}
	if size == headerFixedSize {
		return Header{}, 0, stacktrace.NewError("missing request type")
	}

	return Header{
		ID:          uuid.UUID(data[0:16]),
		NodeID:      binary.BigEndian.Uint64(data[16:24]),
		Timestamp:   time.Unix(0, int64(binary.BigEndian.Uint64(data[24:32]))).UTC(),
		ReadOnly:    flags&headerFlagReadOnly != 0,
		RequestType: string(data[headerFixedSize:size]),
	}, size, nil
}

func (p Proposal) encode() ([]byte, error) {
	buf := make([]byte, 0, 1+p.encodedSize()+len(p.Value))
	buf = append(buf, proposalEncodingVersion)
	buf, err := p.appendEncoded(buf)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to encode proposal header")
	}
	return append(buf, p.Value...), nil
}

func decodeProposal(data []byte) (Proposal, error) {
	if len(data) == 0 {
		return Proposal{}, stacktrace.NewError("empty proposal")
	}
	if data[0] != proposalEncodingVersion {
		return Proposal{}, stacktrace.NewError("unsupported proposal encoding version %d, expected %d", data[0], proposalEncodingVersion)
	}

	header, headerSize, err := decodeHeader(data[1:])
	if err != nil {
		return Proposal{}, stacktrace.Propagate(err, "failed to decode proposal header")
	}
	return Proposal{
		Header: header,
		Value:  data[1+headerSize:],
	}, nil
}

type ProposalResult struct {
	Result any
	Error  error
}

type proposalsTracker struct {
	sync.Mutex
	pending map[uuid.UUID]chan ProposalResult
}

func newProposalsTracker() *proposalsTracker {
	return &proposalsTracker{
		pending: make(map[uuid.UUID]chan ProposalResult),
	}
}

func (p *proposalsTracker) track(id uuid.UUID) chan ProposalResult {
	p.Lock()
	defer p.Unlock()

	applied := make(chan ProposalResult, 1)
	p.pending[id] = applied
	return applied
}

func (p *proposalsTracker) untrack(id uuid.UUID, result ProposalResult) {
	p.Lock()
	defer p.Unlock()

	applied, ok := p.pending[id]
	if !ok {
		return
	}

	applied <- result
	delete(p.pending, id)
}
