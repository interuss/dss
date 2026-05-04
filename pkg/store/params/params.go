package params

import (
	"flag"
)

const (
	RaftStoreType = "raft"
	SQLStoreType  = "sql"
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
	flag.StringVar(&storeParameters.StoreType, "store_type", SQLStoreType, "Store type. Use 'sql' for CockroachDB/YugabyteDB and 'raft' for Raft-based store.")
}

// ConnectParameters returns a ConnectParameters instance that gets populated from well-known CLI flags.
func GetStoreParameters() StoreParameters {
	return storeParameters
}
