package utm

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/api/v1/utmpb"
	dsserr "github.com/interuss/dss/pkg/errors"
)

// Server implements utmpb.DiscoveryAndSynchronizationService.
type Server struct {
	Store   Store
	Timeout time.Duration
}

func (a *Server) DeleteConstraintReference(ctx context.Context, req *utmpb.DeleteConstraintReferenceRequest) (*utmpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) DeleteOperationReference(ctx context.Context, req *utmpb.DeleteOperationReferenceRequest) (*utmpb.ChangeOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) DeleteSubscription(ctx context.Context, req *utmpb.DeleteSubscriptionRequest) (*utmpb.DeleteSubscriptionResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) GetConstraintReference(ctx context.Context, req *utmpb.GetConstraintReferenceRequest) (*utmpb.GetConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) GetOperationReference(ctx context.Context, req *utmpb.GetOperationReferenceRequest) (*utmpb.GetOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) GetSubscription(ctx context.Context, req *utmpb.GetSubscriptionRequest) (*utmpb.GetSubscriptionResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) MakeDssReport(ctx context.Context, req *utmpb.MakeDssReportRequest) (*utmpb.ErrorReport, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) PutConstraintReference(ctx context.Context, req *utmpb.PutConstraintReferenceRequest) (*utmpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) PutOperationReference(ctx context.Context, req *utmpb.PutOperationReferenceRequest) (*utmpb.ChangeOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) PutSubscription(ctx context.Context, req *utmpb.PutSubscriptionRequest) (*utmpb.PutSubscriptionResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) QueryConstraintReferences(ctx context.Context, req *utmpb.QueryConstraintReferencesRequest) (*utmpb.SearchConstraintReferencesResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) QuerySubscriptions(ctx context.Context, req *utmpb.QuerySubscriptionsRequest) (*utmpb.SearchSubscriptionsResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *Server) SearchOperationReferences(ctx context.Context, req *utmpb.SearchOperationReferencesRequest) (*utmpb.SearchOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}
