package server

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/interuss/dss/pkg/api/v1/ridpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/palantir/stacktrace"
)

// GetIdentificationServiceArea returns a single ISA for a given ID.
func (s *Server) GetIdentificationServiceArea(
	ctx context.Context, req *ridpb.GetIdentificationServiceAreaRequest) (
	*ridpb.GetIdentificationServiceAreaResponse, error) {

	id, err := dssmodels.IDFromString(req.Id)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isa, err := s.App.GetISA(ctx, id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not get ISA from application layer")
	}
	if isa == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.NotFound, "ISA %s not found", req.GetId())
	}
	p, err := isa.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not convert ISA to proto")
	}
	return &ridpb.GetIdentificationServiceAreaResponse{
		ServiceArea: p,
	}, nil
}

// CreateIdentificationServiceArea creates an ISA
func (s *Server) CreateIdentificationServiceArea(
	ctx context.Context, req *ridpb.CreateIdentificationServiceAreaRequest) (
	*ridpb.PutIdentificationServiceAreaResponse, error) {

	params := req.GetParams()
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}
	if params == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Params not set")
	}
	// TODO: put the validation logic in the models layer
	if params.FlightsUrl == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required flightsURL")
	}
	if params.Extents == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required extents")
	}
	id, err := dssmodels.IDFromString(req.Id)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:     id,
		URL:    params.GetFlightsUrl(),
		Owner:  owner,
		Writer: s.Locality,
	}

	if err := isa.SetExtents(params.Extents); err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents")
	}

	insertedISA, subscribers, err := s.App.InsertISA(ctx, isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not insert ISA")
	}

	pbISA, err := insertedISA.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not convert ISA to proto")
	}

	pbSubscribers := []*ridpb.SubscriberToNotify{}
	for _, subscriber := range subscribers {
		pbSubscribers = append(pbSubscribers, subscriber.ToNotifyProto())
	}

	return &ridpb.PutIdentificationServiceAreaResponse{
		ServiceArea: pbISA,
		Subscribers: pbSubscribers,
	}, nil
}

// UpdateIdentificationServiceArea updates an existing ISA.
func (s *Server) UpdateIdentificationServiceArea(
	ctx context.Context, req *ridpb.UpdateIdentificationServiceAreaRequest) (
	*ridpb.PutIdentificationServiceAreaResponse, error) {

	params := req.GetParams()

	version, err := dssmodels.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid version")
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}
	// TODO: put the validation logic in the models layer
	if params == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Params not set")
	}
	if params.FlightsUrl == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required flightsURL")
	}
	if params.Extents == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required extents")
	}
	id, err := dssmodels.IDFromString(req.Id)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:      dssmodels.ID(id),
		URL:     params.FlightsUrl,
		Owner:   owner,
		Version: version,
		Writer:  s.Locality,
	}

	if err := isa.SetExtents(params.Extents); err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents")
	}

	insertedISA, subscribers, err := s.App.UpdateISA(ctx, isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not update ISA")
	}

	pbISA, err := insertedISA.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not convert ISA to proto")
	}

	pbSubscribers := []*ridpb.SubscriberToNotify{}
	for _, subscriber := range subscribers {
		pbSubscribers = append(pbSubscribers, subscriber.ToNotifyProto())
	}

	return &ridpb.PutIdentificationServiceAreaResponse{
		ServiceArea: pbISA,
		Subscribers: pbSubscribers,
	}, nil
}

// DeleteIdentificationServiceArea deletes an existing ISA.
func (s *Server) DeleteIdentificationServiceArea(
	ctx context.Context, req *ridpb.DeleteIdentificationServiceAreaRequest) (
	*ridpb.DeleteIdentificationServiceAreaResponse, error) {

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner from context")
	}
	version, err := dssmodels.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid version")
	}
	id, err := dssmodels.IDFromString(req.Id)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isa, subscribers, err := s.App.DeleteISA(ctx, id, owner, version)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not delete ISA")
	}

	p, err := isa.ToProto()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not convert ISA to proto")
	}
	sp := make([]*ridpb.SubscriberToNotify, len(subscribers))
	for i := range subscribers {
		sp[i] = subscribers[i].ToNotifyProto()
	}

	return &ridpb.DeleteIdentificationServiceAreaResponse{
		ServiceArea: p,
		Subscribers: sp,
	}, nil
}

// SearchIdentificationServiceAreas queries for all ISAs in the bounds.
func (s *Server) SearchIdentificationServiceAreas(
	ctx context.Context, req *ridpb.SearchIdentificationServiceAreasRequest) (
	*ridpb.SearchIdentificationServiceAreasResponse, error) {

	cu, err := geo.AreaToCellIDs(req.GetArea())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Invalid area")
	}

	var (
		earliest *time.Time
		latest   *time.Time
	)

	if et := req.GetEarliestTime(); et != nil {
		if ts, err := ptypes.Timestamp(et); err == nil {
			earliest = &ts
		} else {
			return nil, stacktrace.Propagate(err, "Unable to convert earliest timestamp to ptype")
		}
	}

	if lt := req.GetLatestTime(); lt != nil {
		if ts, err := ptypes.Timestamp(lt); err == nil {
			latest = &ts
		} else {
			return nil, stacktrace.Propagate(err, "Unable to convert latest timestamp to ptype")
		}
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isas, err := s.App.SearchISAs(ctx, cu, earliest, latest)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to search ISAs")
	}

	areas := make([]*ridpb.IdentificationServiceArea, len(isas))
	for i := range isas {
		a, err := isas[i].ToProto()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not convert ISA to proto")
		}
		areas[i] = a
	}

	return &ridpb.SearchIdentificationServiceAreasResponse{
		ServiceAreas: areas,
	}, nil
}
