package params

import (
	"flag"
	"fmt"
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

	// Options carries the configuration flags shared by all datastore backends.
	Options struct {
		GlobalLock                 bool
		TimeBasedNotificationIndex bool
	}
)

var (
	storeParameters StoreParameters
	storeOptions    Options
)

func init() {
	flag.StringVar(&storeParameters.StoreType, "store_type", SQLStoreType, fmt.Sprintf("Store type. Use '%s' for CockroachDB/YugabyteDB and '%s' for Raft-based store.", SQLStoreType, RaftStoreType))
	flag.BoolVar(&storeOptions.GlobalLock, "enable_scd_global_lock", false, "Experimental: Use a global lock when working with SCD subscriptions. Reduce global throughput but improve throughput with lot of subscriptions in the same areas.")
	flag.BoolVar(&storeOptions.TimeBasedNotificationIndex, "enable_time_based_notification_index", false, "Use a time-based notification index when working with RID and SCD subscriptions.")
}

// GetStoreParameters returns a StoreParameters instance that gets populated from well-known CLI flags.
func GetStoreParameters() StoreParameters {
	return storeParameters
}

// GetStoreOptions returns the datastore Options populated from well-known CLI flags.
func GetStoreOptions() Options {
	return storeOptions
}
