package cockroach

import (
	"context"
	"errors"

	dssmodels "github.com/interuss/dss/pkg/dss/models"
	scdmodels "github.com/interuss/dss/pkg/dss/scd/models"
)

var (
	errNotImplemented = errors.New("not implemented")
)

func (s *Store) GetOperation(ctx context.Context, id scdmodels.ID) (*scdmodels.Operation, error) {
	return nil, errNotImplemented
}

func (s *Store) DeleteOperation(ctx context.Context, id scdmodels.ID, owner dssmodels.Owner) (*scdmodels.Operation, []*scdmodels.Subscription, error) {
	return nil, nil, errNotImplemented
}

func (s *Store) UpsertOperation(ctx context.Context, operation *scdmodels.Operation, key []scdmodels.OVN) (*scdmodels.Operation, []*scdmodels.Subscription, error) {
	return nil, nil, errNotImplemented
}

func (s *Store) SearchOperations(ctx context.Context, v4d *dssmodels.Volume4D, owner dssmodels.Owner) ([]*scdmodels.Operation, error) {
	return nil, errNotImplemented
}
