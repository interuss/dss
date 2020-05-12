package cockroach

import (
	"context"
	"errors"
	"time"

	"github.com/golang/geo/s2"
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

func (s *Store) UpsertOperation(ctx context.Context, operation *scdmodels.Operation, key []scdmodels.OVN, notifySubscriptionForConstraints bool, subscriptionBaseURL string) (*scdmodels.Operation, []*scdmodels.Subscription, error) {
	return nil, nil, errNotImplemented
}

func (s *Store) SearchOperations(ctx context.Context, cells s2.CellUnion, altitudeLower *float64, altitudeUpper *float64, earliest *time.Time, latest *time.Time) ([]*scdmodels.Operation, error) {
	return nil, errNotImplemented
}
