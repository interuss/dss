package client

import (
	"context"
	"net/http"

	restapi "github.com/interuss/dss/pkg/api/adminv1"
	"github.com/interuss/stacktrace"
)

// EvictionConfig returns the eviction settings of the group.
func (g *RaftGroup) EvictionConfig(ctx context.Context) (*restapi.EvictionConfig, error) {
	resp, err := g.client.GetEvictionConfig(ctx, g.name)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to call the admin API at %s", g.socketPath)
	}
	switch {
	case resp.Response200 != nil:
		return resp.Response200, nil
	case resp.Response404 != nil:
		return nil, responseError(http.StatusNotFound, resp.Response404.Message)
	case resp.Response500 != nil:
		return nil, responseError(http.StatusInternalServerError, &resp.Response500.ErrorMessage)
	default:
		return nil, stacktrace.NewError("no response received from the admin API")
	}
}

// SetEvictionConfig replaces the eviction settings of the group and returns the stored settings.
func (g *RaftGroup) SetEvictionConfig(ctx context.Context, config *restapi.EvictionConfig) (*restapi.EvictionConfig, error) {
	resp, err := g.client.PutEvictionConfig(ctx, g.name, config)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to call the admin API at %s", g.socketPath)
	}
	switch {
	case resp.Response200 != nil:
		return resp.Response200, nil
	case resp.Response400 != nil:
		return nil, responseError(http.StatusBadRequest, resp.Response400.Message)
	case resp.Response404 != nil:
		return nil, responseError(http.StatusNotFound, resp.Response404.Message)
	case resp.Response500 != nil:
		return nil, responseError(http.StatusInternalServerError, &resp.Response500.ErrorMessage)
	default:
		return nil, stacktrace.NewError("no response received from the admin API")
	}
}

// RunEviction runs a cleanup job on the group and returns the number of entities per entity type.
func (g *RaftGroup) RunEviction(ctx context.Context, params *restapi.EvictionRunParameters) (*restapi.EvictionRunResult, error) {
	resp, err := g.client.CreateEvictionRun(ctx, g.name, params)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to call the admin API at %s", g.socketPath)
	}
	switch {
	case resp.Response200 != nil:
		return resp.Response200, nil
	case resp.Response400 != nil:
		return nil, responseError(http.StatusBadRequest, resp.Response400.Message)
	case resp.Response404 != nil:
		return nil, responseError(http.StatusNotFound, resp.Response404.Message)
	case resp.Response500 != nil:
		return nil, responseError(http.StatusInternalServerError, &resp.Response500.ErrorMessage)
	default:
		return nil, stacktrace.NewError("no response received from the admin API")
	}
}
