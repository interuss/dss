package params

import (
	"flag"
	"fmt"
)

const (
	RaftStoreType = "raft"
	SQLStoreType  = "sql"
	MemStoreType  = "mem"
)

type (
	// StoreParameters bundles up parameters used to configure store at a generic/top level.
	StoreParameters struct {
		StoreType string
	}
)

var (
	storeParameters StoreParameters
)

func init() {
	// NB: Memstore not available there on purpose.
	flag.StringVar(&storeParameters.StoreType, "store_type", SQLStoreType, fmt.Sprintf("Store type. Use '%s' for CockroachDB/YugabyteDB and '%s' for Raft-based store.", SQLStoreType, RaftStoreType))
}

// GetStoreParameters returns a StoreParameters instance that gets populated from well-known CLI flags.
func GetStoreParameters() StoreParameters {
	return storeParameters
}
