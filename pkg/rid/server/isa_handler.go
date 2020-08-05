package server

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/interuss/dss/pkg/api/v1/ridpb"
	"github.com/interuss/dss/pkg/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/geo"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
)

// GetIdentificationServiceArea returns a single ISA for a given ID.
func (s *Server) GetIdentificationServiceArea(
	ctx context.Context, req *ridpb.GetIdentificationServiceAreaRequest) (
	*ridpb.GetIdentificationServiceAreaResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isa, err := s.App.GetISA(ctx, dssmodels.ID(req.GetId()))
	if err != nil {
		return nil, err
	}
	if isa == nil {
		return nil, dsserr.NotFound(req.GetId())
	}
	p, err := isa.ToProto()
	if err != nil {
		return nil, err
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
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	if params == nil {
		return nil, dsserr.BadRequest("params not set")
	}
	// TODO: put the validation logic in the models layer
	if params.FlightsUrl == "" {
		return nil, dsserr.BadRequest("missing required flightsURL")
	}
	if params.Extents == nil {
		return nil, dsserr.BadRequest("missing required extents")
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:    dssmodels.ID(req.Id),
		URL:   params.GetFlightsUrl(),
		Owner: owner,
	}

	if err := isa.SetExtents(params.Extents); err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad extents: %s", err))
	}

	insertedISA, subscribers, err := s.App.InsertISA(ctx, isa)
	if err != nil {
		return nil, err
	}

	pbISA, err := insertedISA.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
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
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	// TODO: put the validation logic in the models layer
	if params == nil {
		return nil, dsserr.BadRequest("params not set")
	}
	if params.FlightsUrl == "" {
		return nil, dsserr.BadRequest("missing required flightsURL")
	}
	if params.Extents == nil {
		return nil, dsserr.BadRequest("missing required extents")
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:      dssmodels.ID(req.Id),
		URL:     params.FlightsUrl,
		Owner:   owner,
		Version: version,
		Writer:  s.Locality,
	}

	if err := isa.SetExtents(params.Extents); err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad extents: %s", err))
	}

	insertedISA, subscribers, err := s.App.UpdateISA(ctx, isa)
	if err != nil {
		return nil, err
	}

	pbISA, err := insertedISA.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
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
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	version, err := dssmodels.VersionFromString(req.GetVersion())
	if err != nil {
		return nil, dsserr.BadRequest(fmt.Sprintf("bad version: %s", err))
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isa, subscribers, err := s.App.DeleteISA(ctx, dssmodels.ID(req.GetId()), owner, version)
	if err != nil {
		return nil, err
	}

	p, err := isa.ToProto()
	if err != nil {
		return nil, dsserr.Internal(err.Error())
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
		errMsg := fmt.Sprintf("bad area: %s", err)
		switch err.(type) {
		case *geo.ErrAreaTooLarge:
			return nil, dsserr.AreaTooLarge(errMsg)
		}
		return nil, dsserr.BadRequest(errMsg)
	}

	var (
		earliest *time.Time
		latest   *time.Time
	)

	if et := req.GetEarliestTime(); et != nil {
		if ts, err := ptypes.Timestamp(et); err == nil {
			earliest = &ts
		} else {
			return nil, dsserr.Internal(err.Error())
		}
	}

	if lt := req.GetLatestTime(); lt != nil {
		if ts, err := ptypes.Timestamp(lt); err == nil {
			latest = &ts
		} else {
			return nil, dsserr.Internal(err.Error())
		}
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isas, err := s.App.SearchISAs(ctx, cu, earliest, latest)
	if err != nil {
		return nil, err
	}

	areas := make([]*ridpb.IdentificationServiceArea, len(isas))
	for i := range isas {
		a, err := isas[i].ToProto()
		if err != nil {
			return nil, err
		}
		areas[i] = a
	}

	return &ridpb.SearchIdentificationServiceAreasResponse{
		ServiceAreas: areas,
	}, nil
}
