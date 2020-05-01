package dss

import (
	"context"
	"fmt"

	"github.com/interuss/dss/pkg/api/v1/auxpb"
	"github.com/interuss/dss/pkg/dss/auth"
	dsserr "github.com/interuss/dss/pkg/errors"
)

// AuxServer implements auxpb.DSSAuxService.
type AuxServer struct{}

// ValidateOauth will exercise validating the Oauth token
func (a *AuxServer) ValidateOauth(ctx context.Context, req *auxpb.ValidateOauthRequest) (*auxpb.ValidateOauthResponse, error) {
	owner, ok := auth.OwnerFromContext(ctx)
	if !ok {
		return nil, dsserr.PermissionDenied("missing owner from context")
	}
	if req.Owner != "" && req.Owner != owner.String() {
		return nil, dsserr.PermissionDenied(fmt.Sprintf("owner mismatch, required: %s, but oauth token has %s", req.Owner, owner))
	}
	return &auxpb.ValidateOauthResponse{}, nil
}
