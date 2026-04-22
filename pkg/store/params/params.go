package params

import (
	"flag"
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
	flag.StringVar(&storeParameters.StoreType, "store_type", "sql", "Store type. Use 'sql' for CockroachDB/YugabyteDB, or 'raft' for raft implementation")
}

// ConnectParameters returns a ConnectParameters instance that gets populated from well-known CLI flags.
func GetStoreParameters() StoreParameters {
	return storeParameters
}
