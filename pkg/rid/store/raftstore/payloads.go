package raftstore

import (
	"time"

	"github.com/golang/geo/s2"
)

// expiredPayload carries the arguments common to ListExpiredISAs/ListExpiredSubscriptions.
type expiredPayload struct {
	Writer    string    `json:"writer"`
	Threshold time.Time `json:"threshold"`
}

// searchISAsPayload carries the arguments of SearchISAs.
type searchISAsPayload struct {
	Cells    s2.CellUnion `json:"cells"`
	Earliest *time.Time   `json:"earliest,omitempty"`
	Latest   *time.Time   `json:"latest,omitempty"`
}
