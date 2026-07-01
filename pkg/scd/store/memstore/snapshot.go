package memstore

import (
	"github.com/interuss/stacktrace"
)

func (r *repo) GetSnapshot() ([]byte, error) {
	return nil, stacktrace.NewError("GetSnapshot not yet implemented for rid")
}

func (r *repo) RestoreFromSnapshot(data []byte) error {
	return stacktrace.NewError("RestoreFromSnapshot not yet implemented for rid")
}
