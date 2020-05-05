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

// DeleteConstraintReference deletes a single constraint ref for a given ID at
// the specified version.
func (a *Server) DeleteConstraintReference(ctx context.Context, req *utmpb.DeleteConstraintReferenceRequest) (*utmpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// DeleteOperationReference deletes a single operation ref for a given ID at
// the specified version.
func (a *Server) DeleteOperationReference(ctx context.Context, req *utmpb.DeleteOperationReferenceRequest) (*utmpb.ChangeOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// DeleteSubscription deletes a single subscription for a given ID at the
// specified version.
func (a *Server) DeleteSubscription(ctx context.Context, req *utmpb.DeleteSubscriptionRequest) (*utmpb.DeleteSubscriptionResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// GetConstraintReference returns a single constraint ref for the given ID.
func (a *Server) GetConstraintReference(ctx context.Context, req *utmpb.GetConstraintReferenceRequest) (*utmpb.GetConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// GetOperationReference returns a single operation ref for the given ID.
func (a *Server) GetOperationReference(ctx context.Context, req *utmpb.GetOperationReferenceRequest) (*utmpb.GetOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// GetSubscription returns a single subscription for the given ID.
func (a *Server) GetSubscription(ctx context.Context, req *utmpb.GetSubscriptionRequest) (*utmpb.GetSubscriptionResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// MakeDssReport creates an error report about a DSS.
func (a *Server) MakeDssReport(ctx context.Context, req *utmpb.MakeDssReportRequest) (*utmpb.ErrorReport, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// PutConstraintReference creates a single contraint ref.
func (a *Server) PutConstraintReference(ctx context.Context, req *utmpb.PutConstraintReferenceRequest) (*utmpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// PutOperationReference creates a single operation ref.
func (a *Server) PutOperationReference(ctx context.Context, req *utmpb.PutOperationReferenceRequest) (*utmpb.ChangeOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// PutSubscription creates a single subscription.
func (a *Server) PutSubscription(ctx context.Context, req *utmpb.PutSubscriptionRequest) (*utmpb.PutSubscriptionResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// QueryConstraintReferences queries existing contraint refs in the given
// bounds.
func (a *Server) QueryConstraintReferences(ctx context.Context, req *utmpb.QueryConstraintReferencesRequest) (*utmpb.SearchConstraintReferencesResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// QuerySubscriptions queries existing subscriptions in the given bounds.
func (a *Server) QuerySubscriptions(ctx context.Context, req *utmpb.QuerySubscriptionsRequest) (*utmpb.SearchSubscriptionsResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// SearchOperationReferences queries existing operation refs in the given
// bounds.
func (a *Server) SearchOperationReferences(ctx context.Context, req *utmpb.SearchOperationReferencesRequest) (*utmpb.SearchOperationReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}
