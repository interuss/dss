package consensus

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
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

const gzipThreshold = 1024

const proposalHeaderSize = 16 + 8 + 1 + 1

const (
	flagReadOnly   byte = 1 << 0
	flagCompressed byte = 1 << 1
)

type Proposal struct {
	ID          string
	Timestamp   time.Time
	RequestType string
	Value       []byte
	ReadOnly    bool
	Compressed  bool
}

type ProposalResult struct {
	Result any
	Error  error
}

func newProposal(ctx context.Context, requestType string, payload any, readOnly bool) (Proposal, error) {
	proposalTimestamp := timestamp.NowFromContext(ctx)
	if proposalTimestamp.IsZero() {
		proposalTimestamp = time.Now().UTC()
	}

	value, err := json.Marshal(payload)
	if err != nil {
		return Proposal{}, stacktrace.Propagate(err, "failed to serialize proposal payload")
	}

	p := Proposal{
		ID:          uuid.NewString(),
		Timestamp:   proposalTimestamp,
		RequestType: requestType,
		Value:       value,
		ReadOnly:    readOnly,
	}

	if len(value) > gzipThreshold {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		if _, err := zw.Write(value); err != nil {
			return Proposal{}, stacktrace.Propagate(err, "failed to compress proposal payload")
		}
		if err := zw.Close(); err != nil {
			return Proposal{}, stacktrace.Propagate(err, "failed to compress proposal payload")
		}
		p.Value = buf.Bytes()
		p.Compressed = true
	}

	return p, nil
}

func (p Proposal) Encode() ([]byte, error) {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to parse proposal ID %q", p.ID)
	}

	rt := []byte(p.RequestType)
	if len(rt) > 255 {
		return nil, stacktrace.NewError("request type %q exceeds 255 bytes", p.RequestType)
	}

	var flags byte
	if p.ReadOnly {
		flags |= flagReadOnly
	}
	if p.Compressed {
		flags |= flagCompressed
	}

	buf := make([]byte, 0, proposalHeaderSize+len(rt)+len(p.Value))
	buf = append(buf, id[:]...)
	buf = binary.LittleEndian.AppendUint64(buf, uint64(p.Timestamp.UnixNano()))
	buf = append(buf, flags, byte(len(rt)))
	buf = append(buf, rt...)
	buf = append(buf, p.Value...)
	return buf, nil
}

func (p *Proposal) Decode(data []byte) error {
	if len(data) < proposalHeaderSize {
		return stacktrace.NewError("proposal data too short: %d bytes", len(data))
	}

	p.ID = uuid.UUID(data[:16]).String()
	p.Timestamp = time.Unix(0, int64(binary.LittleEndian.Uint64(data[16:24]))).UTC()

	flags := data[24]
	p.ReadOnly = flags&flagReadOnly != 0
	p.Compressed = flags&flagCompressed != 0

	n := int(data[25])
	if len(data) < proposalHeaderSize+n {
		return stacktrace.NewError("proposal data too short for request type of %d bytes", n)
	}

	p.RequestType = string(data[proposalHeaderSize : proposalHeaderSize+n])
	p.Value = data[proposalHeaderSize+n:]
	return nil
}

func (p *Proposal) Decompress() error {
	if !p.Compressed {
		return nil
	}

	zr, err := gzip.NewReader(bytes.NewReader(p.Value))
	if err != nil {
		return stacktrace.Propagate(err, "failed to create gzip reader")
	}

	value, err := io.ReadAll(zr)
	if err != nil {
		return stacktrace.Propagate(err, "failed to decompress proposal payload")
	}

	p.Value = value
	p.Compressed = false
	return nil
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
