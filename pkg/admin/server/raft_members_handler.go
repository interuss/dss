package server

import (
	"context"
	"maps"
	"slices"
	"strconv"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/adminv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	"github.com/interuss/dss/pkg/raftstore/consensus/node"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

func (s *Server) GetThisRaftNode(ctx context.Context, req *restapi.GetThisRaftNodeRequest) restapi.GetThisRaftNodeResponseSet {
	group, err := s.raftGroup(req.GroupName)
	if err != nil {
		return restapi.GetThisRaftNodeResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}

	node, err := group.Node()
	if err != nil {
		return restapi.GetThisRaftNodeResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Could not describe this node in the %s raft group", req.GroupName))}}
	}
	return restapi.GetThisRaftNodeResponseSet{Response200: toRaftNode(node)}
}

func (s *Server) CreateRaftGroup(ctx context.Context, req *restapi.CreateRaftGroupRequest) restapi.CreateRaftGroupResponseSet {
	if req.BodyParseError != nil {
		return restapi.CreateRaftGroupResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	group, err := s.raftGroup(req.Body.GroupName)
	if err != nil {
		return restapi.CreateRaftGroupResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}

	membership, err := group.Bootstrap(ctx)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not create the %s raft group", req.Body.GroupName)
		switch stacktrace.GetCode(err) {
		case dsserr.AlreadyExists:
			return restapi.CreateRaftGroupResponseSet{Response409: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}}
		default:
			return restapi.CreateRaftGroupResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	s.Logger.Info("created raft group", zap.String("group", string(req.Body.GroupName)))
	return restapi.CreateRaftGroupResponseSet{Response200: toRaftGroup(req.Body.GroupName, membership)}
}

func (s *Server) GetRaftGroup(ctx context.Context, req *restapi.GetRaftGroupRequest) restapi.GetRaftGroupResponseSet {
	group, err := s.raftGroup(req.GroupName)
	if err != nil {
		return restapi.GetRaftGroupResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}

	membership, err := group.Members()
	if err != nil {
		err = stacktrace.Propagate(err, "Could not get the members of the %s raft group", req.GroupName)
		switch stacktrace.GetCode(err) {
		// NotFound: this node is still waiting to create or join the group.
		case dsserr.NotFound:
			return restapi.GetRaftGroupResponseSet{Response404: &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}}
		default:
			return restapi.GetRaftGroupResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}
	return restapi.GetRaftGroupResponseSet{Response200: toRaftGroup(req.GroupName, membership)}
}

func (s *Server) CreateRaftGroupMember(ctx context.Context, req *restapi.CreateRaftGroupMemberRequest) restapi.CreateRaftGroupMemberResponseSet {
	group, err := s.raftGroup(req.GroupName)
	if err != nil {
		return restapi.CreateRaftGroupMemberResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}
	if req.BodyParseError != nil {
		return restapi.CreateRaftGroupMemberResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}

	added, err := group.AddNode(ctx, req.Body.Url)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not add node %s to the %s raft group", req.Body.Url, req.GroupName)
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.BadRequest:
			return restapi.CreateRaftGroupMemberResponseSet{Response400: errResp}
		case dsserr.AlreadyExists:
			return restapi.CreateRaftGroupMemberResponseSet{Response409: errResp}
		// Unavailable: the node cannot be reached, or this node is not a member of the group.
		case dsserr.Unavailable:
			return restapi.CreateRaftGroupMemberResponseSet{Response503: errResp}
		default:
			return restapi.CreateRaftGroupMemberResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	s.Logger.Info("added node to raft group", zap.String("group", string(req.GroupName)), zap.Uint64("nodeID", added.ID), zap.String("url", added.URL))
	return restapi.CreateRaftGroupMemberResponseSet{Response200: toRaftNode(added)}
}

func (s *Server) DeleteRaftGroupMember(ctx context.Context, req *restapi.DeleteRaftGroupMemberRequest) restapi.DeleteRaftGroupMemberResponseSet {
	group, err := s.raftGroup(req.GroupName)
	if err != nil {
		return restapi.DeleteRaftGroupMemberResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}
	nodeID, err := parseNodeID(req.NodeId)
	if err != nil {
		return restapi.DeleteRaftGroupMemberResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}

	err = group.RemoveNode(ctx, nodeID)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not remove node %d from the %s raft group", nodeID, req.GroupName)
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.BadRequest, dsserr.Unavailable:
			return restapi.DeleteRaftGroupMemberResponseSet{Response400: errResp}
		default:
			return restapi.DeleteRaftGroupMemberResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	s.Logger.Info("removed node from raft group", zap.String("group", string(req.GroupName)), zap.Uint64("nodeID", nodeID))
	return restapi.DeleteRaftGroupMemberResponseSet{Response204: &api.EmptyResponseBody{}}
}

// raftGroup returns the Raft group or a dsserr.NotFound error.
func (s *Server) raftGroup(name restapi.RaftGroupName) (*consensus.Consensus, error) {
	group, ok := s.raftGroups[string(name)]
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Unknown raft group %q, expected one of %v", name, slices.Sorted(maps.Keys(s.raftGroups)))
	}
	return group, nil
}

// parseNodeID parses a node ID formatted as a decimal string, or returns a dsserr.BadRequest error.
func parseNodeID(s string) (uint64, error) {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil || id == 0 {
		return 0, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid node ID %q, expected a non-zero decimal integer", s)
	}
	return id, nil
}

func toRaftNode(info node.Info) *restapi.RaftNode {
	return &restapi.RaftNode{Id: strconv.FormatUint(info.ID, 10), Url: info.URL, State: string(info.State)}
}

// toRaftGroup returns the group with its nodes: its members first, then the nodes removed from it, each by ID.
func toRaftGroup(name restapi.RaftGroupName, membership consensus.Membership) *restapi.RaftGroup {
	result := &restapi.RaftGroup{
		GroupName: name,
		ClusterId: strconv.FormatUint(membership.ClusterID, 10),
		Nodes:     []restapi.RaftNode{},
	}
	// The leader is unknown while an election is in progress.
	if membership.LeaderID != 0 {
		result.LeaderId = strconv.FormatUint(membership.LeaderID, 10)
	}
	ids := slices.Sorted(maps.Keys(membership.Nodes))
	for _, state := range []node.State{node.Member, node.Removed} {
		for _, id := range ids {
			if info := membership.Nodes[id]; info.State == state {
				result.Nodes = append(result.Nodes, *toRaftNode(info))
			}
		}
	}
	return result
}
