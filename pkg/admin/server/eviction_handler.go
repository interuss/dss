package server

import (
	"context"
	"errors"
	"io"
	"maps"
	"slices"
	"time"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/adminv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	dssmodels "github.com/interuss/dss/pkg/models"
	storerepos "github.com/interuss/dss/pkg/store/repos"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

func (s *Server) GetEvictionConfig(ctx context.Context, req *restapi.GetEvictionConfigRequest) restapi.GetEvictionConfigResponseSet {
	eviction, err := s.eviction(req.GroupName)
	if err != nil {
		return restapi.GetEvictionConfigResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}

	config, err := eviction.GetEvictionConfig(ctx)
	if err != nil {
		return restapi.GetEvictionConfigResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Could not get the eviction config of the %s raft group", req.GroupName))}}
	}
	return restapi.GetEvictionConfigResponseSet{Response200: toEvictionConfig(config)}
}

func (s *Server) PutEvictionConfig(ctx context.Context, req *restapi.PutEvictionConfigRequest) restapi.PutEvictionConfigResponseSet {
	eviction, err := s.eviction(req.GroupName)
	if err != nil {
		return restapi.PutEvictionConfigResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}
	if req.BodyParseError != nil {
		return restapi.PutEvictionConfigResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	config, err := fromEvictionConfig(req.Body)
	if err != nil {
		return restapi.PutEvictionConfigResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}

	stored, err := eviction.SetEvictionConfig(ctx, config)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not set the eviction config of the %s raft group", req.GroupName)
		switch stacktrace.GetCode(err) {
		case dsserr.BadRequest:
			return restapi.PutEvictionConfigResponseSet{Response400: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}}
		default:
			return restapi.PutEvictionConfigResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	s.Logger.Info("set eviction config", zap.String("group", string(req.GroupName)), zap.Duration("interval", stored.Interval),
		zap.Duration("ttl", stored.Parameters.TTL), zap.Int("limit", stored.Parameters.Limit))
	return restapi.PutEvictionConfigResponseSet{Response200: toEvictionConfig(stored)}
}

func (s *Server) CreateEvictionRun(ctx context.Context, req *restapi.CreateEvictionRunRequest) restapi.CreateEvictionRunResponseSet {
	eviction, err := s.eviction(req.GroupName)
	if err != nil {
		return restapi.CreateEvictionRunResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}
	if req.BodyParseError != nil && !errors.Is(req.BodyParseError, io.EOF) {
		return restapi.CreateEvictionRunResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	// Without parameters, those of the eviction settings are proposed, so that every node applies the same ones.
	var params dssmodels.EvictionParameters
	if req.Body.Parameters != nil {
		params, err = fromEvictionParameters(req.Body.Parameters)
		if err != nil {
			return restapi.CreateEvictionRunResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, err)}}
		}
	} else {
		config, err := eviction.GetEvictionConfig(ctx)
		if err != nil {
			return restapi.CreateEvictionRunResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Could not get the eviction config of the %s raft group", req.GroupName))}}
		}
		params = config.Parameters
	}

	counts, err := eviction.DeleteExpired(ctx, params)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not run a cleanup job on the %s raft group", req.GroupName)
		switch stacktrace.GetCode(err) {
		case dsserr.BadRequest:
			return restapi.CreateEvictionRunResponseSet{Response400: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}}
		default:
			return restapi.CreateEvictionRunResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	result := &restapi.EvictionRunResult{Counts: []restapi.EntityEvictionCount{}}
	for _, entityType := range slices.Sorted(maps.Keys(counts)) {
		result.Counts = append(result.Counts, restapi.EntityEvictionCount{Type: entityType, Count: int64(counts[entityType])})
	}
	s.Logger.Info("ran cleanup job", zap.String("group", string(req.GroupName)), zap.Any("counts", counts))
	return restapi.CreateEvictionRunResponseSet{Response200: result}
}

// eviction returns the eviction of the Raft group or a dsserr.NotFound error.
func (s *Server) eviction(name restapi.RaftGroupName) (storerepos.Eviction, error) {
	eviction, ok := s.evictions[string(name)]
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "No eviction for raft group %q", name)
	}
	return eviction, nil
}

func toEvictionConfig(config dssmodels.EvictionConfig) *restapi.EvictionConfig {
	return &restapi.EvictionConfig{
		Interval: config.Interval.String(),
		Parameters: restapi.EvictionParameters{
			Ttl:   config.Parameters.TTL.String(),
			Limit: new(int64(config.Parameters.Limit)),
		},
	}
}

// fromEvictionConfig converts and validates config, or returns a dsserr.BadRequest error.
func fromEvictionConfig(config *restapi.EvictionConfig) (dssmodels.EvictionConfig, error) {
	interval, err := time.ParseDuration(config.Interval)
	if err != nil || interval < 0 {
		return dssmodels.EvictionConfig{}, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid interval %q, expected a duration of 0 or more", config.Interval)
	}
	params, err := fromEvictionParameters(&config.Parameters)
	if err != nil {
		return dssmodels.EvictionConfig{}, err
	}
	return dssmodels.EvictionConfig{Interval: interval, Parameters: params}, nil
}

// fromEvictionParameters converts and validates params, or returns a dsserr.BadRequest error.
func fromEvictionParameters(params *restapi.EvictionParameters) (dssmodels.EvictionParameters, error) {
	ttl, err := time.ParseDuration(params.Ttl)
	if err != nil || ttl <= 0 {
		return dssmodels.EvictionParameters{}, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid TTL %q, expected a positive duration", params.Ttl)
	}
	var limit int64
	if params.Limit != nil {
		limit = *params.Limit
	}
	if limit < 0 {
		return dssmodels.EvictionParameters{}, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid limit %d, expected 0 (unlimited) or more", limit)
	}
	return dssmodels.EvictionParameters{TTL: ttl, Limit: int(limit)}, nil
}
