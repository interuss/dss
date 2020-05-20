package scd

import (
	"context"

	"github.com/interuss/dss/pkg/api/v1/scdpb"
	dsserr "github.com/interuss/dss/pkg/errors"
)

// DeleteConstraintReference deletes a single constraint ref for a given ID at
// the specified version.
func (a *Server) DeleteConstraintReference(ctx context.Context, req *scdpb.DeleteConstraintReferenceRequest) (*scdpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// GetConstraintReference returns a single constraint ref for the given ID.
func (a *Server) GetConstraintReference(ctx context.Context, req *scdpb.GetConstraintReferenceRequest) (*scdpb.GetConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// PutConstraintReference creates a single contraint ref.
func (a *Server) PutConstraintReference(ctx context.Context, req *scdpb.PutConstraintReferenceRequest) (*scdpb.ChangeConstraintReferenceResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}

// QueryConstraintReferences queries existing contraint refs in the given
// bounds.
func (a *Server) QueryConstraintReferences(ctx context.Context, req *scdpb.QueryConstraintReferencesRequest) (*scdpb.SearchConstraintReferencesResponse, error) {
	return nil, dsserr.BadRequest("not yet implemented")
}
