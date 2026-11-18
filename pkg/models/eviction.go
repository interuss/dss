package models

import (
	"time"

	"github.com/interuss/dss/pkg/pb/evictionpb"
	"github.com/interuss/stacktrace"
	"google.golang.org/protobuf/types/known/durationpb"
)

// EvictionParameters contains the cleanup parameters of a Raft group, applied to each of its entity types.
type EvictionParameters struct {
	// TTL is how long after their end entities are expired.
	TTL time.Duration
	// Limit is the maximum number of entities evicted per entity type, 0 means unlimited.
	Limit int
}

func (p *EvictionParameters) ToProto() *evictionpb.Parameters {
	return &evictionpb.Parameters{
		Ttl:   durationpb.New(p.TTL),
		Limit: int64(p.Limit),
	}
}

// FromProto converts and validates parameters.
func (p *EvictionParameters) FromProto(pb *evictionpb.Parameters) error {
	if pb == nil {
		return stacktrace.NewError("missing eviction parameters")
	}
	if err := pb.GetTtl().CheckValid(); err != nil {
		return stacktrace.Propagate(err, "invalid TTL")
	}
	params := EvictionParameters{
		TTL:   pb.GetTtl().AsDuration(),
		Limit: int(pb.GetLimit()),
	}
	if params.TTL <= 0 {
		return stacktrace.NewError("TTL must be positive, got %s", params.TTL)
	}
	if params.Limit < 0 {
		return stacktrace.NewError("limit must not be negative, got %d", params.Limit)
	}
	*p = params
	return nil
}

// EvictionConfig contains the eviction settings of a Raft group.
type EvictionConfig struct {
	// Interval is the time between two scheduled cleanup jobs, 0 disables them.
	Interval   time.Duration
	Parameters EvictionParameters
}

// EvictionState contains the eviction settings of a Raft group and the last run of its cleanup jobs.
type EvictionState struct {
	Config EvictionConfig
	// LastRun is the proposal timestamp of the last successful cleanup job, zero if none ran.
	LastRun time.Time
}

func (c *EvictionConfig) ToProto() *evictionpb.Config {
	return &evictionpb.Config{
		Interval:   durationpb.New(c.Interval),
		Parameters: c.Parameters.ToProto(),
	}
}

// FromProto converts and validates the settings.
func (c *EvictionConfig) FromProto(pb *evictionpb.Config) error {
	if err := pb.GetInterval().CheckValid(); err != nil {
		return stacktrace.Propagate(err, "invalid interval")
	}
	interval := pb.GetInterval().AsDuration()
	if interval < 0 {
		return stacktrace.NewError("interval must not be negative, got %s", interval)
	}
	var params EvictionParameters
	if err := params.FromProto(pb.GetParameters()); err != nil {
		return err
	}
	*c = EvictionConfig{Interval: interval, Parameters: params}
	return nil
}
