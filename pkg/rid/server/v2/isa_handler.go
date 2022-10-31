package server

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/api"
	restapi "github.com/interuss/dss/pkg/api/ridv2"
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
func (s *Server) GetIdentificationServiceArea(ctx context.Context, req *restapi.GetIdentificationServiceAreaRequest,
) restapi.GetIdentificationServiceAreaResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.GetIdentificationServiceAreaResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	id, err := dssmodels.IDFromString(string(req.Id))
	if err != nil {
		return restapi.GetIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format"))}}
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isa, err := s.App.GetISA(ctx, id)
	if err != nil {
		return restapi.GetIdentificationServiceAreaResponseSet{Response500: &api.InternalServerErrorBody{
			ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Could not get ISA from application layer"))}}
	}
	if isa == nil {
		return restapi.GetIdentificationServiceAreaResponseSet{Response404: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.NotFound, "ISA %s not found", req.Id))}}
	}
	return restapi.GetIdentificationServiceAreaResponseSet{Response200: &restapi.GetIdentificationServiceAreaResponse{
		ServiceArea: *apiv2.ToIdentificationServiceArea(isa)}}
}

// CreateIdentificationServiceArea creates an ISA
func (s *Server) CreateIdentificationServiceArea(ctx context.Context, req *restapi.CreateIdentificationServiceAreaRequest,
) restapi.CreateIdentificationServiceAreaResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.CreateIdentificationServiceAreaResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	if req.Auth.ClientID == nil {
		return restapi.CreateIdentificationServiceAreaResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}
	if req.BodyParseError != nil {
		return restapi.CreateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	// TODO: put the validation logic in the models layer
	if req.Body.UssBaseUrl == "" {
		return restapi.CreateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required USS base URL"))}}
	}
	extents, err := apiv2.FromVolume4D(&req.Body.Extents)
	if err != nil {
		return restapi.CreateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err)))}}
	}
	id, err := dssmodels.IDFromString(string(req.Id))
	if err != nil {
		return restapi.CreateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format"))}}
	}

	if !s.EnableHTTP {
		err = ridmodels.ValidateURL(string(req.Body.UssBaseUrl))
		if err != nil {
			return restapi.CreateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Failed to validate base URL"))}}
		}
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:     id,
		URL:    string(req.Body.UssBaseUrl),
		Owner:  dssmodels.Owner(*req.Auth.ClientID),
		Writer: s.Locality,
	}

	if err := isa.SetExtents(extents); err != nil {
		return restapi.CreateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents"))}}
	}

	insertedISA, subscribers, err := s.App.InsertISA(ctx, isa)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not insert ISA")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.AlreadyExists:
			return restapi.CreateIdentificationServiceAreaResponseSet{Response409: errResp}
		case dsserr.BadRequest:
			return restapi.CreateIdentificationServiceAreaResponseSet{Response400: errResp}
		default:
			return restapi.CreateIdentificationServiceAreaResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	apiSubscribers := make([]restapi.SubscriberToNotify, 0, len(subscribers))
	for _, subscriber := range subscribers {
		apiSubscribers = append(apiSubscribers, *apiv2.ToSubscriberToNotify(subscriber))
	}

	return restapi.CreateIdentificationServiceAreaResponseSet{Response200: &restapi.PutIdentificationServiceAreaResponse{
		ServiceArea: *apiv2.ToIdentificationServiceArea(insertedISA),
		Subscribers: &apiSubscribers,
	}}
}

// UpdateIdentificationServiceArea updates an existing ISA.
func (s *Server) UpdateIdentificationServiceArea(ctx context.Context, req *restapi.UpdateIdentificationServiceAreaRequest,
) restapi.UpdateIdentificationServiceAreaResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.UpdateIdentificationServiceAreaResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	version, err := dssmodels.VersionFromString(req.Version)
	if err != nil {
		return restapi.UpdateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid version"))}}
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	if req.Auth.ClientID == nil {
		return restapi.UpdateIdentificationServiceAreaResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}
	// TODO: put the validation logic in the models layer
	if req.BodyParseError != nil {
		return restapi.UpdateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(req.BodyParseError, dsserr.BadRequest, "Malformed params"))}}
	}
	if req.Body.UssBaseUrl == "" {
		return restapi.UpdateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing required USS base URL"))}}
	}
	extents, err := apiv2.FromVolume4D(&req.Body.Extents)
	if err != nil {
		return restapi.UpdateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Error parsing Volume4D: %v", stacktrace.RootCause(err)))}}
	}
	id, err := dssmodels.IDFromString(string(req.Id))
	if err != nil {
		return restapi.UpdateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format"))}}
	}

	isa := &ridmodels.IdentificationServiceArea{
		ID:      id,
		URL:     string(req.Body.UssBaseUrl),
		Owner:   dssmodels.Owner(*req.Auth.ClientID),
		Version: version,
		Writer:  s.Locality,
	}

	if err := isa.SetExtents(extents); err != nil {
		return restapi.UpdateIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid extents"))}}
	}

	insertedISA, subscribers, err := s.App.UpdateISA(ctx, isa)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not update ISA")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.UpdateIdentificationServiceAreaResponseSet{Response403: errResp}
		case dsserr.VersionMismatch:
			return restapi.UpdateIdentificationServiceAreaResponseSet{Response409: errResp}
		case dsserr.BadRequest, dsserr.NotFound:
			return restapi.UpdateIdentificationServiceAreaResponseSet{Response400: errResp}
		default:
			return restapi.UpdateIdentificationServiceAreaResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	apiSubscribers := make([]restapi.SubscriberToNotify, 0, len(subscribers))
	for _, subscriber := range subscribers {
		apiSubscribers = append(apiSubscribers, *apiv2.ToSubscriberToNotify(subscriber))
	}

	return restapi.UpdateIdentificationServiceAreaResponseSet{Response200: &restapi.PutIdentificationServiceAreaResponse{
		ServiceArea: *apiv2.ToIdentificationServiceArea(insertedISA),
		Subscribers: &apiSubscribers,
	}}
}

// DeleteIdentificationServiceArea deletes an existing ISA.
func (s *Server) DeleteIdentificationServiceArea(ctx context.Context, req *restapi.DeleteIdentificationServiceAreaRequest,
) restapi.DeleteIdentificationServiceAreaResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.DeleteIdentificationServiceAreaResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.Auth.ClientID == nil {
		return restapi.DeleteIdentificationServiceAreaResponseSet{Response403: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "Missing owner"))}}
	}

	version, err := dssmodels.VersionFromString(req.Version)
	if err != nil {
		return restapi.DeleteIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid version"))}}
	}
	id, err := dssmodels.IDFromString(string(req.Id))
	if err != nil {
		return restapi.DeleteIdentificationServiceAreaResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Invalid ID format"))}}
	}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isa, subscribers, err := s.App.DeleteISA(ctx, id, dssmodels.Owner(*req.Auth.ClientID), version)
	if err != nil {
		err = stacktrace.Propagate(err, "Could not delete ISA")
		errResp := &restapi.ErrorResponse{Message: dsserr.Handle(ctx, err)}
		switch stacktrace.GetCode(err) {
		case dsserr.PermissionDenied:
			return restapi.DeleteIdentificationServiceAreaResponseSet{Response403: errResp}
		case dsserr.VersionMismatch:
			return restapi.DeleteIdentificationServiceAreaResponseSet{Response409: errResp}
		case dsserr.NotFound:
			return restapi.DeleteIdentificationServiceAreaResponseSet{Response404: errResp}
		default:
			return restapi.DeleteIdentificationServiceAreaResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	apiSubscribers := make([]restapi.SubscriberToNotify, 0, len(subscribers))
	for _, subscriber := range subscribers {
		apiSubscribers = append(apiSubscribers, *apiv2.ToSubscriberToNotify(subscriber))
	}

	return restapi.DeleteIdentificationServiceAreaResponseSet{Response200: &restapi.DeleteIdentificationServiceAreaResponse{
		ServiceArea: *apiv2.ToIdentificationServiceArea(isa),
		Subscribers: &apiSubscribers,
	}}
}

// SearchIdentificationServiceAreas queries for all ISAs in the bounds.
func (s *Server) SearchIdentificationServiceAreas(ctx context.Context, req *restapi.SearchIdentificationServiceAreasRequest,
) restapi.SearchIdentificationServiceAreasResponseSet {
	if req.Auth.Error != nil {
		resp := restapi.SearchIdentificationServiceAreasResponseSet{}
		setAuthError(ctx, stacktrace.Propagate(req.Auth.Error, "Auth failed"), &resp.Response401, &resp.Response403, &resp.Response500)
		return resp
	}

	if req.Area == nil {
		return restapi.SearchIdentificationServiceAreasResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing area"))}}
	}
	cu, err := geo.AreaToCellIDs(string(*req.Area))
	if err != nil {
		if errors.Is(err, geoerr.ErrAreaTooLarge) {
			return restapi.SearchIdentificationServiceAreasResponseSet{Response413: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, stacktrace.Propagate(err, "Invalid area"))}}
		}
		return restapi.SearchIdentificationServiceAreasResponseSet{Response400: &restapi.ErrorResponse{
			Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Invalid area"))}}
	}

	var (
		earliest *time.Time
		latest   *time.Time
	)

	if req.EarliestTime != nil {
		ts, err := time.Parse(time.RFC3339Nano, *req.EarliestTime)
		if err != nil {
			return restapi.SearchIdentificationServiceAreasResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Unable to convert earliest timestamp"))}}
		}
		earliest = &ts
	}

	if req.LatestTime != nil {
		ts, err := time.Parse(time.RFC3339Nano, *req.LatestTime)
		if err != nil {
			return restapi.SearchIdentificationServiceAreasResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, stacktrace.PropagateWithCode(err, dsserr.BadRequest, "Unable to convert latest timestamp"))}}
		}
		latest = &ts
	}

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	isas, err := s.App.SearchISAs(ctx, cu, earliest, latest)
	if err != nil {
		err = stacktrace.Propagate(err, "Unable to search ISAs")
		if stacktrace.GetCode(err) == dsserr.BadRequest {
			return restapi.SearchIdentificationServiceAreasResponseSet{Response400: &restapi.ErrorResponse{
				Message: dsserr.Handle(ctx, err)}}
		} else {
			return restapi.SearchIdentificationServiceAreasResponseSet{Response500: &api.InternalServerErrorBody{
				ErrorMessage: *dsserr.Handle(ctx, stacktrace.Propagate(err, "Got an unexpected error"))}}
		}
	}

	areas := make([]restapi.IdentificationServiceArea, 0, len(isas))
	for _, isa := range isas {
		areas = append(areas, *apiv2.ToIdentificationServiceArea(isa))
	}

	return restapi.SearchIdentificationServiceAreasResponseSet{Response200: &restapi.SearchIdentificationServiceAreasResponse{
		ServiceAreas: &areas,
	}}
}
