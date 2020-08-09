package cockroach

import (
	"context"
	"time"

	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"

	"github.com/golang/geo/s2"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
	dssql "github.com/interuss/dss/pkg/sql"

)

const (
	isaFields       = "id, owner, url, cells, starts_at, ends_at, updated_at"
	updateISAFields = "id, url, cells, starts_at, ends_at, updated_at"

	isaFieldsV3       = "id, owner, url, cells, starts_at, ends_at, updated_at"
	updateISAFieldsV3 = "id, url, cells, starts_at, ends_at, updated_at"
)

type IISARepo interface {
	GetISA(ctx context.Context, id dssmodels.ID) (*ridmodels.IdentificationServiceArea, error)
	InsertISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error)
	UpdateISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error)
	DeleteISA(ctx context.Context, isa *ridmodels.IdentificationServiceArea) (*ridmodels.IdentificationServiceArea, error)
	SearchISAs(ctx context.Context, cells s2.CellUnion, earliest *time.Time, latest *time.Time) ([]*ridmodels.IdentificationServiceArea, error)
}

func NewISARepo(ctx context.Context, db dssql.Queryable, dbVersion string, logger *zap.Logger) IISARepo {
	if (semver.Compare(dbVersion, "v3.1.0") >= 0) {
		return &isaRepoLatest{
			Queryable: db,
			logger:    logger,
		}
	} else {
		return &isaRepoV3{
			Queryable: db,
			logger:    logger,
		}
	}
}
