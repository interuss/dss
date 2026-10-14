// This file is auto-generated; do not change as any changes will be overwritten
package adminv1

import (
	"context"
	"github.com/interuss/dss/pkg/api"
)

var (
	GetRaftMembersSecurity   = []api.AuthorizationOption{}
	AddRaftMemberSecurity    = []api.AuthorizationOption{}
	RemoveRaftMemberSecurity = []api.AuthorizationOption{}
)

type GetRaftMembersRequest struct {
	// Name of the Raft group (aux, rid or scd).
	Group string

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const GetRaftMembersOperationID = "adminv1.GetRaftMembers"

func (request *GetRaftMembersRequest) OperationID() string { return GetRaftMembersOperationID }

type GetRaftMembersResponseSet struct {
	// The members of the Raft group are successfully returned.
	Response200 *RaftGroupMembers

	// This DSS node has no Raft group with this name.
	Response404 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type AddRaftMemberRequest struct {
	// Name of the Raft group (aux, rid or scd).
	Group string

	// The data contained in the body of this request, if it parsed correctly
	Body *RaftMember

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const AddRaftMemberOperationID = "adminv1.AddRaftMember"

func (request *AddRaftMemberRequest) OperationID() string { return AddRaftMemberOperationID }

type AddRaftMemberResponseSet struct {
	// The node was added. The resulting members of the Raft group are returned.
	Response200 *RaftGroupMembers

	// The request was not properly formed, or adding the node would lead to an invalid membership.
	Response400 *ErrorResponse

	// This DSS node has no Raft group with this name.
	Response404 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type RemoveRaftMemberRequest struct {
	// Name of the Raft group (aux, rid or scd).
	Group string

	// ID of the node to remove.
	NodeId string

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const RemoveRaftMemberOperationID = "adminv1.RemoveRaftMember"

func (request *RemoveRaftMemberRequest) OperationID() string { return RemoveRaftMemberOperationID }

type RemoveRaftMemberResponseSet struct {
	// The node was removed. The resulting members of the Raft group are returned.
	Response200 *RaftGroupMembers

	// The node ID is invalid, or removing the node would lead to an invalid membership.
	Response400 *ErrorResponse

	// This DSS node has no Raft group with this name.
	Response404 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type Implementation interface {
	// Lists the members of a Raft group, as known by this DSS node.
	GetRaftMembers(ctx context.Context, req *GetRaftMembersRequest) GetRaftMembersResponseSet

	// Adds a node to a Raft group, as a voter.
	// ---
	// The node must be running, started with `--raft_join`, before being added: as soon as it is added, it counts in the quorum of the group. The request returns once the change is applied by this DSS node.
	AddRaftMember(ctx context.Context, req *AddRaftMemberRequest) AddRaftMemberResponseSet

	// Removes a node from a Raft group.
	// ---
	// The ID of the removed node is retired permanently: it can never be added again to the group. A voter cannot be removed if the group would be left with fewer than 3 voters. The request returns once the change is applied by this DSS node.
	RemoveRaftMember(ctx context.Context, req *RemoveRaftMemberRequest) RemoveRaftMemberResponseSet
}
