package cockroach

import (
	"context"

	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"log"
)

func (u *repo) UpsertUssAvailability(ctx context.Context, s *scdmodels.UssAvailabilityStatus) (*scdmodels.UssAvailabilityStatus, error) {
	// todo: yet to implement
	log.Println("Implement set uss availability!")
	return nil, nil
}

// GetUssAvailability returns the Availability status identified by "id".
func (u *repo) GetUssAvailability(ctx context.Context, id dssmodels.Manager) (*scdmodels.UssAvailabilityStatus, error) {
	// todo: yet to implement
	return &scdmodels.UssAvailabilityStatus{Availability: "Unknown"}, nil
}
