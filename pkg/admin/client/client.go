package client

import (
	"context"
	"flag"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/interuss/dss/pkg/admin"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/adminv1"
	"github.com/interuss/stacktrace"
)

var (
	adminURL = flag.String("admin_url", "http://localhost:8078", "URL of the admin API of core-service, matching its --admin_addr")
	timeout  = flag.Duration("admin_timeout", 1*time.Minute, "Timeout of the requests to the admin API")
)

// group is an admin.RaftGroup managed through the admin API.
type group struct {
	client *restapi.Client
	name   string
}

// InitRaftGroup returns the Raft group with this name of the DSS node whose admin API is configured by the --admin_url flag.
func InitRaftGroup(name string) (admin.RaftGroup, error) {
	if _, err := url.ParseRequestURI(*adminURL); err != nil {
		return nil, stacktrace.Propagate(err, "invalid --admin_url %q", *adminURL)
	}

	return &group{
		client: &restapi.Client{BaseURL: *adminURL, HTTPClient: &http.Client{Timeout: *timeout}},
		name:   name,
	}, nil
}

func (g *group) Members(ctx context.Context) (admin.Membership, error) {
	resp, err := g.client.GetRaftMembers(ctx, g.name)
	if err != nil {
		return admin.Membership{}, stacktrace.Propagate(err, "failed to call the admin API at %s", g.client.BaseURL)
	}
	return toMembership(resp.Response200, nil, resp.Response404, resp.Response500)
}

func (g *group) AddNode(ctx context.Context, nodeID uint64, peerURL string) (admin.Membership, error) {
	resp, err := g.client.AddRaftMember(ctx, g.name, &restapi.RaftMember{Id: int64(nodeID), Url: peerURL})
	if err != nil {
		return admin.Membership{}, stacktrace.Propagate(err, "failed to call the admin API at %s", g.client.BaseURL)
	}
	return toMembership(resp.Response200, resp.Response400, resp.Response404, resp.Response500)
}

func (g *group) RemoveNode(ctx context.Context, nodeID uint64) (admin.Membership, error) {
	resp, err := g.client.RemoveRaftMember(ctx, g.name, strconv.FormatUint(nodeID, 10))
	if err != nil {
		return admin.Membership{}, stacktrace.Propagate(err, "failed to call the admin API at %s", g.client.BaseURL)
	}
	return toMembership(resp.Response200, resp.Response400, resp.Response404, resp.Response500)
}

// toMembership returns the membership of a successful response, or an error describing the error response received.
func toMembership(members *restapi.RaftGroupMembers, badRequest, notFound *restapi.ErrorResponse, internalError *api.InternalServerErrorBody) (admin.Membership, error) {
	switch {
	case members != nil:
		membership := admin.Membership{
			NodeID:   uint64(members.NodeId),
			LeaderID: uint64(members.LeaderId),
			Members:  map[uint64]string{},
		}
		for _, member := range members.Members {
			membership.Members[uint64(member.Id)] = member.Url
		}
		for _, id := range members.RemovedIds {
			membership.RemovedIDs = append(membership.RemovedIDs, uint64(id))
		}
		return membership, nil
	case badRequest != nil:
		return admin.Membership{}, responseError(http.StatusBadRequest, badRequest.Message)
	case notFound != nil:
		return admin.Membership{}, responseError(http.StatusNotFound, notFound.Message)
	case internalError != nil:
		return admin.Membership{}, responseError(http.StatusInternalServerError, &internalError.ErrorMessage)
	default:
		return admin.Membership{}, stacktrace.NewError("no response received from the admin API")
	}
}

func responseError(status int, message *string) error {
	if message == nil || *message == "" {
		return stacktrace.NewError("%d %s", status, http.StatusText(status))
	}
	return stacktrace.NewError("%d %s: %s", status, http.StatusText(status), *message)
}
