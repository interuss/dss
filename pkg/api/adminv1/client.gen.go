// This file is auto-generated; do not change as any changes will be overwritten
package adminv1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/interuss/dss/pkg/api"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client calls the operations of the API.
type Client struct {
	// BaseURL is the URL the paths of the API are relative to, e.g. http://localhost:8080.
	BaseURL string

	// HTTPClient sends the requests, http.DefaultClient if nil.
	HTTPClient *http.Client
}

// do sends a request to the API, with body encoded in JSON if not nil.
func (c *Client) do(ctx context.Context, method string, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(c.BaseURL, "/")+path, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return httpClient.Do(req)
}

// Describes this DSS node in a Raft group.
func (c *Client) GetThisRaftNode(ctx context.Context, groupName RaftGroupName) (GetThisRaftNodeResponseSet, error) {
	var response GetThisRaftNodeResponseSet

	path := "/admin/v1/raft/groups/" + url.PathEscape(fmt.Sprint(groupName)) + "/this_node"

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case 200:
		response.Response200 = new(RaftNode)
		err = json.NewDecoder(resp.Body).Decode(response.Response200)
	case 404:
		response.Response404 = new(ErrorResponse)
		err = json.NewDecoder(resp.Body).Decode(response.Response404)
	case 500:
		response.Response500 = new(api.InternalServerErrorBody)
		err = json.NewDecoder(resp.Body).Decode(response.Response500)
	default:
		err = fmt.Errorf("unexpected response status %s", resp.Status)
	}
	return response, err
}

// Creates a new Raft group with this DSS node as its only member.
func (c *Client) CreateRaftGroup(ctx context.Context, body *CreateRaftGroupParameters) (CreateRaftGroupResponseSet, error) {
	var response CreateRaftGroupResponseSet

	path := "/admin/v1/raft/groups"

	var requestBody any
	if body != nil {
		requestBody = body
	}
	resp, err := c.do(ctx, http.MethodPost, path, requestBody)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case 200:
		response.Response200 = new(RaftGroup)
		err = json.NewDecoder(resp.Body).Decode(response.Response200)
	case 400:
		response.Response400 = new(ErrorResponse)
		err = json.NewDecoder(resp.Body).Decode(response.Response400)
	case 404:
		response.Response404 = new(ErrorResponse)
		err = json.NewDecoder(resp.Body).Decode(response.Response404)
	case 409:
		response.Response409 = new(ErrorResponse)
		err = json.NewDecoder(resp.Body).Decode(response.Response409)
	case 500:
		response.Response500 = new(api.InternalServerErrorBody)
		err = json.NewDecoder(resp.Body).Decode(response.Response500)
	default:
		err = fmt.Errorf("unexpected response status %s", resp.Status)
	}
	return response, err
}

// Describes a Raft group, as known by this DSS node.
func (c *Client) GetRaftGroup(ctx context.Context, groupName RaftGroupName) (GetRaftGroupResponseSet, error) {
	var response GetRaftGroupResponseSet

	path := "/admin/v1/raft/groups/" + url.PathEscape(fmt.Sprint(groupName))

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case 200:
		response.Response200 = new(RaftGroup)
		err = json.NewDecoder(resp.Body).Decode(response.Response200)
	case 404:
		response.Response404 = new(ErrorResponse)
		err = json.NewDecoder(resp.Body).Decode(response.Response404)
	case 500:
		response.Response500 = new(api.InternalServerErrorBody)
		err = json.NewDecoder(resp.Body).Decode(response.Response500)
	default:
		err = fmt.Errorf("unexpected response status %s", resp.Status)
	}
	return response, err
}

// Adds a node to a Raft group, as a voter.
func (c *Client) CreateRaftGroupMember(ctx context.Context, groupName RaftGroupName, body *CreateRaftGroupMemberParameters) (CreateRaftGroupMemberResponseSet, error) {
	var response CreateRaftGroupMemberResponseSet

	path := "/admin/v1/raft/groups/" + url.PathEscape(fmt.Sprint(groupName)) + "/members"

	var requestBody any
	if body != nil {
		requestBody = body
	}
	resp, err := c.do(ctx, http.MethodPost, path, requestBody)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case 200:
		response.Response200 = new(RaftNode)
		err = json.NewDecoder(resp.Body).Decode(response.Response200)
	case 400:
		response.Response400 = new(ErrorResponse)
		err = json.NewDecoder(resp.Body).Decode(response.Response400)
	case 404:
		response.Response404 = new(ErrorResponse)
		err = json.NewDecoder(resp.Body).Decode(response.Response404)
	case 500:
		response.Response500 = new(api.InternalServerErrorBody)
		err = json.NewDecoder(resp.Body).Decode(response.Response500)
	default:
		err = fmt.Errorf("unexpected response status %s", resp.Status)
	}
	return response, err
}

// Removes a node from a Raft group.
func (c *Client) DeleteRaftGroupMember(ctx context.Context, groupName RaftGroupName, nodeId string) (DeleteRaftGroupMemberResponseSet, error) {
	var response DeleteRaftGroupMemberResponseSet

	path := "/admin/v1/raft/groups/" + url.PathEscape(fmt.Sprint(groupName)) + "/members/" + url.PathEscape(nodeId)

	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return response, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case 204:
		response.Response204 = &api.EmptyResponseBody{}
	case 400:
		response.Response400 = new(ErrorResponse)
		err = json.NewDecoder(resp.Body).Decode(response.Response400)
	case 404:
		response.Response404 = new(ErrorResponse)
		err = json.NewDecoder(resp.Body).Decode(response.Response404)
	case 500:
		response.Response500 = new(api.InternalServerErrorBody)
		err = json.NewDecoder(resp.Body).Decode(response.Response500)
	default:
		err = fmt.Errorf("unexpected response status %s", resp.Status)
	}
	return response, err
}
