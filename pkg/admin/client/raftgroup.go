package client

import (
	"context"
	"net"
	"net/http"
	"time"

	restapi "github.com/interuss/dss/pkg/api/adminv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
)

// statusCodes maps the HTTP status of the error responses of the admin API to the code of the error of the DSS node.
var statusCodes = map[int]stacktrace.ErrorCode{
	http.StatusBadRequest:         dsserr.BadRequest,
	http.StatusNotFound:           dsserr.NotFound,
	http.StatusConflict:           dsserr.AlreadyExists,
	http.StatusServiceUnavailable: dsserr.Unavailable,
}

type RaftGroup struct {
	client     *restapi.Client
	socketPath string
	name       restapi.RaftGroupName
}

func NewRaftGroup(socketPath string, name string, timeout time.Duration) (*RaftGroup, error) {
	if socketPath == "" {
		return nil, stacktrace.NewError("the path of the Unix socket of the admin API is required")
	}

	// The requests are sent over the Unix socket whatever their URL, whose host is only used in the Host header.
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socketPath)
		},
	}
	return &RaftGroup{
		client:     &restapi.Client{BaseURL: "http://localhost", HTTPClient: &http.Client{Transport: transport, Timeout: timeout}},
		socketPath: socketPath,
		name:       restapi.RaftGroupName(name),
	}, nil
}

func (g *RaftGroup) Node(ctx context.Context) (*restapi.RaftNode, error) {
	resp, err := g.client.GetThisRaftNode(ctx, g.name)
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

func (g *RaftGroup) Bootstrap(ctx context.Context) (*restapi.RaftGroup, error) {
	resp, err := g.client.CreateRaftGroup(ctx, &restapi.CreateRaftGroupParameters{GroupName: g.name})
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
	case resp.Response409 != nil:
		return nil, responseError(http.StatusConflict, resp.Response409.Message)
	case resp.Response500 != nil:
		return nil, responseError(http.StatusInternalServerError, &resp.Response500.ErrorMessage)
	default:
		return nil, stacktrace.NewError("no response received from the admin API")
	}
}

func (g *RaftGroup) Members(ctx context.Context) (*restapi.RaftGroup, error) {
	resp, err := g.client.GetRaftGroup(ctx, g.name)
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

func (g *RaftGroup) AddNode(ctx context.Context, peerURL string) (*restapi.RaftNode, error) {
	resp, err := g.client.CreateRaftGroupMember(ctx, g.name, &restapi.CreateRaftGroupMemberParameters{Url: peerURL})
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
	case resp.Response409 != nil:
		return nil, responseError(http.StatusConflict, resp.Response409.Message)
	case resp.Response503 != nil:
		return nil, responseError(http.StatusServiceUnavailable, resp.Response503.Message)
	case resp.Response500 != nil:
		return nil, responseError(http.StatusInternalServerError, &resp.Response500.ErrorMessage)
	default:
		return nil, stacktrace.NewError("no response received from the admin API")
	}
}

func (g *RaftGroup) RemoveNode(ctx context.Context, nodeID string) error {
	resp, err := g.client.DeleteRaftGroupMember(ctx, g.name, nodeID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to call the admin API at %s", g.socketPath)
	}
	switch {
	case resp.Response204 != nil:
		return nil
	case resp.Response400 != nil:
		return responseError(http.StatusBadRequest, resp.Response400.Message)
	case resp.Response404 != nil:
		return responseError(http.StatusNotFound, resp.Response404.Message)
	case resp.Response500 != nil:
		return responseError(http.StatusInternalServerError, &resp.Response500.ErrorMessage)
	default:
		return stacktrace.NewError("no response received from the admin API")
	}
}

// responseError returns the error described by an error response of the admin API, with the code of statusCodes.
func responseError(status int, message *string) error {
	text := http.StatusText(status)
	if message != nil && *message != "" {
		text += ": " + *message
	}
	if code, ok := statusCodes[status]; ok {
		return stacktrace.NewErrorWithCode(code, "%d %s", status, text)
	}
	return stacktrace.NewError("%d %s", status, text)
}
