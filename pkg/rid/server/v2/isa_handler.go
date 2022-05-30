package server

import (
	"context"
	"time"

	ridpb "github.com/interuss/dss/pkg/api/v2/ridpbv2"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	geoerr "github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	apiv2 "github.com/interuss/dss/pkg/rid/models/api/v2"
	"github.com/interuss/stacktrace"
	"github.com/pkg/errors"
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
	return &ridpb.GetIdentificationServiceAreaResponse{
		ServiceArea: apiv2.ToIdentificationServiceArea(isa),
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
	if params.UssBaseUrl == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required USS base URL")
	}
	if params.Extents == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required extents")
	}
	extents, err := apiv2.FromVolume4D(params.Extents)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err))
	}
	id, err := dssmodels.IDFromString(req.Id)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	if !s.EnableHTTP {
		err = ridmodels.ValidateURL(params.GetUssBaseUrl())
		if err != nil {
			return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate base URL")
		}
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:     id,
		URL:    params.GetUssBaseUrl() + "/uss/flights",
		Owner:  owner,
		Writer: s.Locality,
	}

	if err := isa.SetExtents(extents); err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents")
	}

	insertedISA, subscribers, err := s.App.InsertISA(ctx, isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not insert ISA")
	}

	pbISA := apiv2.ToIdentificationServiceArea(insertedISA)

	pbSubscribers := []*ridpb.SubscriberToNotify{}
	for _, subscriber := range subscribers {
		pbSubscribers = append(pbSubscribers, apiv2.ToSubscriberToNotify(subscriber))
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
	if params.UssBaseUrl == "" {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required USS base URL")
	}
	if params.Extents == nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required extents")
	}
	extents, err := apiv2.FromVolume4D(params.Extents)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err))
	}
	id, err := dssmodels.IDFromString(req.Id)
	if err != nil {
		return nil, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format")
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:      dssmodels.ID(id),
		URL:     params.UssBaseUrl + "/uss/flights",
		Owner:   owner,
		Version: version,
		Writer:  s.Locality,
	}

	if err := isa.SetExtents(extents); err != nil {
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents")
	}

	insertedISA, subscribers, err := s.App.UpdateISA(ctx, isa)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not update ISA")
	}

	pbISA := apiv2.ToIdentificationServiceArea(insertedISA)

	pbSubscribers := []*ridpb.SubscriberToNotify{}
	for _, subscriber := range subscribers {
		pbSubscribers = append(pbSubscribers, apiv2.ToSubscriberToNotify(subscriber))
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

	p := apiv2.ToIdentificationServiceArea(isa)
	sp := make([]*ridpb.SubscriberToNotify, len(subscribers))
	for i := range subscribers {
		sp[i] = apiv2.ToSubscriberToNotify(subscribers[i])
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
		if errors.Is(err, geoerr.ErrAreaTooLarge) {
			return nil, stacktrace.Propagate(err, "Invalid area")
		}
		return nil, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid area")
	}

	var (
		earliest *time.Time
		latest   *time.Time
	)

	if et := req.GetEarliestTime(); et != nil {
		ts := et.AsTime()
		err := et.CheckValid()
		if err == nil {
			earliest = &ts
		} else {
			return nil, stacktrace.Propagate(err, "Unable to convert earliest timestamp to ptype")
		}
	}

	if lt := req.GetLatestTime(); lt != nil {
		ts := lt.AsTime()
		err := lt.CheckValid()
		if err == nil {
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
		areas[i] = apiv2.ToIdentificationServiceArea(isas[i])
	}

	return &ridpb.SearchIdentificationServiceAreasResponse{
		ServiceAreas: areas,
	}, nil
}
