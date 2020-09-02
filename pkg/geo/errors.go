package geo

import (
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
)

var (
	// ErrMissingSpatialVolume indicates that a spatial volume is required but
	// missing to complete an operation.
	ErrMissingSpatialVolume = stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing spatial volume")

	// ErrMissingFootprint indicates that a geometry footprint is required but
	// missing to complete an operation.
	ErrMissingFootprint = stacktrace.NewErrorWithCode(dsserr.BadRequest, "Missing footprint")

	// ErrNotEnoughPointsInPolygon indicates that a polygon did not contain enough
	// vertices to define a valid shape.
	ErrNotEnoughPointsInPolygon = stacktrace.NewErrorWithCode(dsserr.BadRequest, "Not enough points in polygon")

	// ErrBadCoordSet indicates that a polygon's coordinates did not form a valid
	// singular enclosed area.
	ErrBadCoordSet = stacktrace.NewErrorWithCode(dsserr.BadRequest, "Coordinates did not create a well-formed area")

	// ErrRadiusMustBeLargerThan0 indicates that a circle with non-positive radius
	// was specified.
	ErrRadiusMustBeLargerThan0 = stacktrace.NewErrorWithCode(dsserr.BadRequest, "Radius must be larger than 0")

	// ErrAreaTooLarge is the error passed back when the requested Area is larger
	// than maxAllowedAreaKm2
	ErrAreaTooLarge = stacktrace.NewErrorWithCode(dsserr.AreaTooLarge, "Area too large")

	// ErrOddNumberOfCoordinatesInAreaString indicates that an area string that
	// was supposed to contain lat,lng,lat,lng,... contained only lat for its last
	// coordinate pair.
	ErrOddNumberOfCoordinatesInAreaString = stacktrace.NewErrorWithCode(dsserr.BadRequest, "Odd number of coordinates in area string")
)
