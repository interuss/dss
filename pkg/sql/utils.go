package sql

import (
	"slices"
	"time"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/geo"
	"github.com/interuss/stacktrace"
)

func CellUnionToCellIds(cu s2.CellUnion) []int64 {
	pgCids := make([]int64, len(cu))
	for i, cell := range cu {
		// TODO consider validating the cell here: it is/was done in many similar conversion loops
		pgCids[i] = int64(cell)
	}
	// Sort the cell IDs for optimisation purpose (see https://github.com/interuss/dss/issues/814)
	slices.Sort(pgCids)
	return pgCids
}

func CellUnionToCellIdsWithValidation(cu s2.CellUnion) ([]int64, error) {
	for _, cell := range cu {
		if err := geo.ValidateCell(cell); err != nil {
			return nil, stacktrace.Propagate(err, "Error validating cell")
		}
	}
	return CellUnionToCellIds(cu), nil
}

func ForUpdate(forUpdate bool) string {
	if forUpdate {
		return "FOR UPDATE"
	}
	return ""
}

func MillisSinceMidnight() int {
	now := time.Now().UTC()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return int(now.Sub(midnight).Milliseconds())
}
