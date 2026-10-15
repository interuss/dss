package server

import (
	"context"
	"maps"
	"slices"
	"strconv"

	"github.com/interuss/dss/pkg/admin"
	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/adminv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

func (s *Server) GetRaftMembers(ctx context.Context, req *restapi.GetRaftMembersRequest) restapi.GetRaftMembersResponseSet {
	group, err := s.raftGroup(req.Group)
	if err != nil {
		return restapi.GetRaftMembersResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}

	membership, err := group.Members(ctx)
	if err != nil {
		return restapi.GetRaftMembersResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Could not get the members of the %s raft group", req.Group))}}
	}
	return restapi.GetRaftMembersResponseSet{Response200: toRaftGroupMembers(membership)}
}

func (s *Server) AddRaftMember(ctx context.Context, req *restapi.AddRaftMemberRequest) restapi.AddRaftMemberResponseSet {
	group, err := s.raftGroup(req.Group)
	if err != nil {
		return restapi.AddRaftMemberResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}
	if req.BodyParseError != nil {
		return restapi.AddRaftMemberResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Body.Id < 1 {
		return restapi.AddRaftMemberResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid node ID %d", req.Body.Id))}}
	}

	membership, err := group.AddNode(ctx, uint64(req.Body.Id), req.Body.Url)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not add node %d to the %s raft group", req.Body.Id, req.Group)
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.BadRequest:
			return restapi.AddRaftMemberResponseSet{Response400: errResp}
		default:
			return restapi.AddRaftMemberResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	s.Logger.Info("added node to raft group", zap.String("group", req.Group), zap.Int64("nodeID", req.Body.Id), zap.String("url", req.Body.Url))
	return restapi.AddRaftMemberResponseSet{Response200: toRaftGroupMembers(membership)}
}

func (s *Server) RemoveRaftMember(ctx context.Context, req *restapi.RemoveRaftMemberRequest) restapi.RemoveRaftMemberResponseSet {
	group, err := s.raftGroup(req.Group)
	if err != nil {
		return restapi.RemoveRaftMemberResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, err)}}
	}
	nodeID, err := strconv.ParseUint(req.NodeId, 10, 64)
	if err != nil {
		return restapi.RemoveRaftMemberResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid node ID %q", req.NodeId))}}
	}

	membership, err := group.RemoveNode(ctx, nodeID)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not remove node %d from the %s raft group", nodeID, req.Group)
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.BadRequest:
			return restapi.RemoveRaftMemberResponseSet{Response400: errResp}
		default:
			return restapi.RemoveRaftMemberResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	s.Logger.Info("removed node from raft group", zap.String("group", req.Group), zap.Uint64("nodeID", nodeID))
	return restapi.RemoveRaftMemberResponseSet{Response200: toRaftGroupMembers(membership)}
}

// raftGroup returns the Raft group name, or a dsserr.NotFound error.
func (s *Server) raftGroup(name string) (admin.RaftGroup, error) {
	group, ok := s.RaftGroups[name]
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "Unknown raft group %q, expected one of %v", name, slices.Sorted(maps.Keys(s.RaftGroups)))
	}
	return group, nil
}

func toRaftGroupMembers(membership admin.Membership) *restapi.RaftGroupMembers {
	result := &restapi.RaftGroupMembers{
		NodeId:     int64(membership.NodeID),
		LeaderId:   int64(membership.LeaderID),
		Members:    []restapi.RaftMember{},
		RemovedIds: []int64{},
	}
	for _, id := range slices.Sorted(maps.Keys(membership.Members)) {
		result.Members = append(result.Members, restapi.RaftMember{Id: int64(id), Url: membership.Members[id]})
	}
	for _, id := range slices.Sorted(slices.Values(membership.RemovedIDs)) {
		result.RemovedIds = append(result.RemovedIds, int64(id))
	}
	return result
}
