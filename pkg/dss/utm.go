package dss

import (
	"context"
	"time"

	"github.com/interuss/dss/pkg/api/v1/utmpb"
	dsserr "github.com/interuss/dss/pkg/errors"
)

// UtmServer implements utmpb.DiscoveryAndSynchronizationService.
type UtmServer struct {
	Store   Store
	Timeout time.Duration
}

func (a *UtmServer) DeleteConstraintReference(ctx context.Context, req *utmpb.DeleteConstraintReferenceRequest) (*utmpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) DeleteOperationReference(ctx context.Context, req *utmpb.DeleteOperationReferenceRequest) (*utmpb.ChangeOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) DeleteSubscription(ctx context.Context, req *utmpb.DeleteSubscriptionRequest) (*utmpb.DeleteSubscriptionResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) GetConstraintReference(ctx context.Context, req *utmpb.GetConstraintReferenceRequest) (*utmpb.GetConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) GetOperationReference(ctx context.Context, req *utmpb.GetOperationReferenceRequest) (*utmpb.GetOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) GetSubscription(ctx context.Context, req *utmpb.GetSubscriptionRequest) (*utmpb.GetSubscriptionResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) MakeDssReport(ctx context.Context, req *utmpb.MakeDssReportRequest) (*utmpb.ErrorReport, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) PutConstraintReference(ctx context.Context, req *utmpb.PutConstraintReferenceRequest) (*utmpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) PutOperationReference(ctx context.Context, req *utmpb.PutOperationReferenceRequest) (*utmpb.ChangeOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) PutSubscription(ctx context.Context, req *utmpb.PutSubscriptionRequest) (*utmpb.PutSubscriptionResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) QueryConstraintReferences(ctx context.Context, req *utmpb.QueryConstraintReferencesRequest) (*utmpb.SearchConstraintReferencesResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) QuerySubscriptions(ctx context.Context, req *utmpb.QuerySubscriptionsRequest) (*utmpb.SearchSubscriptionsResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

func (a *UtmServer) SearchOperationReferences(ctx context.Context, req *utmpb.SearchOperationReferencesRequest) (*utmpb.SearchOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}
