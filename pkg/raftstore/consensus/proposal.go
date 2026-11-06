package consensus

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
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

	SnapshotData []byte
}

type Proposal struct {
	ID     string
	NodeID uint64
	// Timestamp is always in UTC so that it is identical on all the nodes.
	// It is encoded as a protobuf Timestamp and decoding always yields a UTC time.
	Timestamp time.Time
	// Exactly one of RequestType (with Value) or Payload is set.
	// TODO: remove RequestType and Value once all the requests are carried by Payload.
	RequestType string
	Value       []byte
	Payload     *raftpb.Payload
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
		return p.RequestType
	}
}

func (p Proposal) encode(compressionThreshold uint64) ([]byte, error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid proposal ID %q", p.ID)
	}
	if (p.RequestType == "") == (p.Payload.GetPayload() == nil) {
		return nil, stacktrace.NewError("exactly one of request type and payload must be set")
	}
	ts := timestamppb.New(p.Timestamp)
	if err := ts.CheckValid(); err != nil {
		return nil, stacktrace.Propagate(err, "invalid timestamp %s", p.Timestamp)
	}

	pb := &raftpb.Proposal{
		Id:          id[:],
		NodeId:      p.NodeID,
		Timestamp:   ts,
		RequestType: p.RequestType,
		Value:       p.Value,
	}
	compressedPayload, err := compressPayload(p.Payload, compressionThreshold)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to compress proposal payload")
	}
	if compressedPayload != nil {
		pb.CompressedPayload = compressedPayload
	} else {
		pb.Payload = p.Payload
	}

	buf, err := proto.Marshal(pb)
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
	payload := pb.GetPayload()
	if compressedPayload := pb.GetCompressedPayload(); compressedPayload != nil {
		if payload != nil {
			return Proposal{}, stacktrace.NewError("payload and compressed payload are both set")
		}
		var err error
		payload, err = decompressPayload(compressedPayload)
		if err != nil {
			return Proposal{}, stacktrace.Propagate(err, "failed to decompress proposal payload")
		}
	}
	if (pb.GetRequestType() == "") == (payload.GetPayload() == nil) {
		return Proposal{}, stacktrace.NewError("exactly one of request type and payload must be set")
	}
	if err := pb.GetTimestamp().CheckValid(); err != nil {
		return Proposal{}, stacktrace.Propagate(err, "invalid proposal timestamp")
	}

	return Proposal{
		ID:          uuid.UUID(pb.GetId()).String(),
		NodeID:      pb.GetNodeId(),
		Timestamp:   pb.GetTimestamp().AsTime(),
		RequestType: pb.GetRequestType(),
		Value:       pb.GetValue(),
		Payload:     payload,
	}, nil
}

// compressPayload returns the gzip-compressed encoding of payload
// or nil if:
// - there is no payload
// - the payload is smaller than compressionThreshold
// - compression is disabled (compressionThreshold is 0)
// - it would not be smaller.
func compressPayload(payload *raftpb.Payload, compressionThreshold uint64) ([]byte, error) {
	if compressionThreshold == 0 || payload.GetPayload() == nil {
		return nil, nil
	}
	buf, err := proto.Marshal(payload)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to marshal payload")
	}
	if uint64(len(buf)) < compressionThreshold {
		return nil, nil
	}
	compressed, err := compress(buf)
	if err != nil {
		return nil, err
	}
	if len(compressed) >= len(buf) {
		return nil, nil
	}
	return compressed, nil
}

func decompressPayload(compressedPayload []byte) (*raftpb.Payload, error) {
	buf, err := decompress(compressedPayload)
	if err != nil {
		return nil, err
	}
	var pb raftpb.Payload
	if err := proto.Unmarshal(buf, &pb); err != nil {
		return nil, stacktrace.Propagate(err, "failed to unmarshal payload")
	}
	if pb.GetPayload() == nil {
		return nil, stacktrace.NewError("missing payload in compressed payload")
	}
	return &pb, nil
}

var gzipWriters = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

var gzipReaders = sync.Pool{
	New: func() any {
		return new(gzip.Reader)
	},
}

func compress(src []byte) ([]byte, error) {
	w := gzipWriters.Get().(*gzip.Writer)
	defer gzipWriters.Put(w)

	var buf bytes.Buffer
	w.Reset(&buf)
	if _, err := w.Write(src); err != nil {
		return nil, stacktrace.Propagate(err, "failed to write compressed data")
	}
	if err := w.Close(); err != nil {
		return nil, stacktrace.Propagate(err, "failed to flush compressed data")
	}
	return buf.Bytes(), nil
}

func decompress(src []byte) ([]byte, error) {
	r := gzipReaders.Get().(*gzip.Reader)
	defer gzipReaders.Put(r)

	if err := r.Reset(bytes.NewReader(src)); err != nil {
		return nil, stacktrace.Propagate(err, "failed to reset decompressor")
	}
	value, err := io.ReadAll(r)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to read decompressed data")
	}
	return value, nil
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
