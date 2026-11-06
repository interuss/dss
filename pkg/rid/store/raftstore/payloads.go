package raftstore

import (
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
)

// cellsByOwnerPayload carries the arguments common to SearchSubscriptionsByOwner/MaxSubscriptionCountInCellsByOwner.
type cellsByOwnerPayload struct {
	Cells s2.CellUnion    `json:"cells"`
	Owner dssmodels.Owner `json:"owner"`
}

// expiredPayload carries the arguments of ListExpiredSubscriptions.
type expiredPayload struct {
	Threshold time.Time `json:"threshold"`
}
