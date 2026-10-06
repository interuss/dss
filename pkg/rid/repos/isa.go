package repos

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
)

// ISA is an interface to a storage layer for the ISA entity
type ISA interface {
	// Returns nil, nil if not found
	GetISA(ctx context.Context, id dssmodels.ID, forUpdate bool) (*ridmodels.IdentificationServiceArea, error)

	// DeleteISA deletes the IdentificationServiceArea identified by "id" and owned by "owner".
	// Returns the delete IdentificationServiceArea and all Subscriptions affected by the delete.
	// Returns nil, nil if ID, version not found
	DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error)

	// InsertISA inserts or updates an ISA.
	InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error)

	// UpdateISA
	// Returns nil, nil if ID, version not found
	UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error)

	// SearchISAs returns all subscriptions ownded by "owner" in "cells".
	SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error)

	// ListExpiredISAs lists the IDs of all expired ISAs based on writer
	ListExpiredISAs(ctx context.Context, writer string, threshold time.Time) ([]dssmodels.ID, error)

	// DeleteExpiredISAs deletes up to `limit` expired ISAs based on writer and returns the IDs of the deleted ISAs. A limit of 0 means unlimited.
	DeleteExpiredISAs(ctx context.Context, writer string, threshold time.Time, limit int) ([]dssmodels.ID, error)

	// DeleteISAsByIDs deletes the ISAs with the given IDs and returns the IDs of the deleted ISAs.
	DeleteISAsByIDs(ctx context.Context, ids []dssmodels.ID) ([]dssmodels.ID, error)

	// Count the number of existing ISA
	CountISAs(ctx context.Context) (int64, error)
}
