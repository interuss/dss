package sql

import (
	"github.com/interuss/dss/pkg/geo"
	"github.com/interuss/stacktrace"

	"github.com/golang/geo/s2"
)

// Converts a cell union to a pgtype.Array[pgtype.Int8] ready for use in database queries
func CellUnionToCellIds(cu s2.CellUnion) []int64 {
	pgCids := make([]int64, len(cu))
	for i, cell := range cu {
		// TODO consider validating the cell here: it is/was done in many similar conversion loops
		pgCids[i] = int64(cell)
	}
	return pgCids
}

func CellUnionToCellIdsWithValidation(cu s2.CellUnion) ([]int64, error) {
	pgCids := make([]int64, len(cu))
	for i, cell := range cu {
		if err := geo.ValidateCell(cell); err != nil {
			return nil, stacktrace.Propagate(err, "Error validating cell")
		}
		pgCids[i] = int64(cell)
	}
	return pgCids, nil
}
