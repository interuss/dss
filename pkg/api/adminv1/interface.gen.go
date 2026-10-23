// This file is auto-generated; do not change as any changes will be overwritten
package adminv1

import (
	"context"
	"github.com/interuss/dss/pkg/api"
)

var (
	GetThisRaftNodeSecurity  = []api.AuthorizationOption{}
	CreateRaftGroupSecurity  = []api.AuthorizationOption{}
	GetRaftGroupSecurity     = []api.AuthorizationOption{}
	CreateRaftMemberSecurity = []api.AuthorizationOption{}
	DeleteRaftMemberSecurity = []api.AuthorizationOption{}
)

type GetThisRaftNodeRequest struct {
	// Name of the Raft group.
	GroupName RaftGroupName

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const GetThisRaftNodeOperationID = "adminv1.GetThisRaftNode"

func (request *GetThisRaftNodeRequest) OperationID() string { return GetThisRaftNodeOperationID }

type GetThisRaftNodeResponseSet struct {
	// This DSS node is successfully described.
	Response200 *RaftNode

	// This DSS node has no Raft group with this name, e.g. `scd` when it runs without `--enable_scd`.
	Response404 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type CreateRaftGroupRequest struct {
	// The data contained in the body of this request, if it parsed correctly
	Body *CreateRaftGroupRequestBody

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const CreateRaftGroupOperationID = "adminv1.CreateRaftGroup"

func (request *CreateRaftGroupRequest) OperationID() string { return CreateRaftGroupOperationID }

type CreateRaftGroupResponseSet struct {
	// The Raft group was created. It is returned with this DSS node as its only member.
	Response200 *RaftGroup

	// The request was not properly formed.
	Response400 *ErrorResponse

	// This DSS node has no Raft group with this name, e.g. `scd` when it runs without `--enable_scd`.
	Response404 *ErrorResponse

	// This DSS node is not waiting, it is or was a member of a Raft group with this name.
	Response409 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type GetRaftGroupRequest struct {
	// Name of the Raft group.
	GroupName RaftGroupName

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const GetRaftGroupOperationID = "adminv1.GetRaftGroup"

func (request *GetRaftGroupRequest) OperationID() string { return GetRaftGroupOperationID }

type GetRaftGroupResponseSet struct {
	// The Raft group is successfully described.
	Response200 *RaftGroup

	// This DSS node has no Raft group with this name, e.g. `scd` when it runs without `--enable_scd`, or it is still waiting to create or join one.
	Response404 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type CreateRaftMemberRequest struct {
	// Name of the Raft group.
	GroupName RaftGroupName

	// The data contained in the body of this request, if it parsed correctly
	Body *CreateRaftMemberRequestBody

	// The error encountered when attempting to parse the body of this request
	BodyParseError error

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const CreateRaftMemberOperationID = "adminv1.CreateRaftMember"

func (request *CreateRaftMemberRequest) OperationID() string { return CreateRaftMemberOperationID }

type CreateRaftMemberResponseSet struct {
	// The node was added. It is returned as a member of the Raft group.
	Response200 *RaftGroupMember

	// The request was not properly formed, or adding the node would lead to an invalid membership.
	Response400 *ErrorResponse

	// This DSS node has no Raft group with this name, e.g. `scd` when it runs without `--enable_scd`.
	Response404 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type DeleteRaftMemberRequest struct {
	// Name of the Raft group.
	GroupName RaftGroupName

	// ID of the node to remove.
	NodeId string

	// The result of attempting to authorize this request
	Auth api.AuthorizationResult
}

const DeleteRaftMemberOperationID = "adminv1.DeleteRaftMember"

func (request *DeleteRaftMemberRequest) OperationID() string { return DeleteRaftMemberOperationID }

type DeleteRaftMemberResponseSet struct {
	// The node was removed.
	Response204 *api.EmptyResponseBody

	// The node ID is invalid, or removing the node would lead to an invalid membership.
	Response400 *ErrorResponse

	// This DSS node has no Raft group with this name, e.g. `scd` when it runs without `--enable_scd`.
	Response404 *ErrorResponse

	// Auto-generated internal server error response
	Response500 *api.InternalServerErrorBody
}

type Implementation interface {
	// Describes this DSS node in a Raft group.
	// ---
	// Representation of this node within the specified group. It is available for every group that the node runs, including while the node waits to create or join it.
	GetThisRaftNode(ctx context.Context, req *GetThisRaftNodeRequest) GetThisRaftNodeResponseSet

	// Creates a new Raft group with this DSS node as its only member.
	// ---
	// This DSS node must be waiting, i.e. not a member of a group with this name yet. The other nodes are then added to the group one at a time. The request returns once this DSS node is the leader of the new group.
	CreateRaftGroup(ctx context.Context, req *CreateRaftGroupRequest) CreateRaftGroupResponseSet

	// Describes a Raft group, as known by this DSS node.
	GetRaftGroup(ctx context.Context, req *GetRaftGroupRequest) GetRaftGroupResponseSet

	// Adds a node to a Raft group, as a voter.
	// ---
	// The request must be sent to a member of the group, and the node to add is another DSS node, which must be running and waiting: as soon as it is added, it counts in the quorum of the group. A node started with an empty data directory waits to be added. The request returns once the change is applied by this DSS node.
	CreateRaftMember(ctx context.Context, req *CreateRaftMemberRequest) CreateRaftMemberResponseSet

	// Removes a node from a Raft group.
	// ---
	// The ID of the removed node is retired permanently: it can never be added again to the group. A voter cannot be removed if the group would be left with fewer than 3 voters. The request returns once the change is applied by this DSS node.
	DeleteRaftMember(ctx context.Context, req *DeleteRaftMemberRequest) DeleteRaftMemberResponseSet
}
