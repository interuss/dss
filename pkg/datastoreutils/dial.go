package datastoreutils

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/interuss/dss/pkg/datastore"
	"github.com/interuss/dss/pkg/datastore/flags"
	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/stacktrace"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const (
	CodeRetryable = stacktrace.ErrorCode(1)
)

var ApplicationName = flags.ConnectParameters().ApplicationName

func getDBStats(ctx context.Context, db *datastore.Datastore, databaseName string) {
	logger := logging.WithValuesFromContext(ctx, logging.Logger)
	statsPtr := db.Pool.Stat()
	stats := make(map[string]string)
	stats["DBName"] = databaseName
	stats["AcquireCount"] = strconv.Itoa(int(statsPtr.AcquireCount()))
	stats["AcquiredConns"] = strconv.Itoa(int(statsPtr.AcquiredConns()))
	stats["CanceledAcquireCount"] = strconv.Itoa(int(statsPtr.CanceledAcquireCount()))
	stats["ConstructingConns"] = strconv.Itoa(int(statsPtr.ConstructingConns()))
	stats["EmptyAcquireCount"] = strconv.Itoa(int(statsPtr.EmptyAcquireCount()))
	stats["IdleConns"] = strconv.Itoa(int(statsPtr.IdleConns()))
	stats["MaxConns"] = strconv.Itoa(int(statsPtr.MaxConns()))
	stats["TotalConns"] = strconv.Itoa(int(statsPtr.TotalConns()))
	if stats["TotalConns"] == "0" {
		logger.Warn("Failed periodic DB Ping (TotalConns=0)", zap.String("Database", databaseName))
	} else {
		logger.Info("Successful periodic DB Ping ", zap.String("Database", databaseName))
	}
}

func DialStore[S any](ctx context.Context, dbName string, newStore func(*datastore.Datastore) (S, error)) (S, error) {

	var zero S

	cp := flags.ConnectParameters()
	cp.DBName = dbName
	cp.ApplicationName = ApplicationName

	db, err := datastore.Dial(ctx, cp)

	if err != nil {
		if strings.Contains(err.Error(), "connect: connection refused") {
			return zero, stacktrace.PropagateWithCode(err, CodeRetryable, "Failed to connect to datastore server for %s", dbName)
		}
		return zero, stacktrace.Propagate(err, "Failed to connect to %s database", dbName)
	}
	s, err := newStore(db)
	if err != nil {
		db.Pool.Close()
		if strings.Contains(err.Error(), "connect: connection refused") || strings.Contains(err.Error(), fmt.Sprintf("database \"%s\" does not exist", dbName)) || strings.Contains(err.Error(), "database has not been bootstrapped with Schema Manager") {
			return zero, stacktrace.PropagateWithCode(err, CodeRetryable, "Failed to create %s store", dbName)
		}
		return zero, stacktrace.Propagate(err, "Failed to create %s store", dbName)
	}

	c := cron.New()
	if _, err := c.AddFunc("@every 1m", func() { getDBStats(ctx, db, dbName) }); err != nil {
		db.Pool.Close()
		return zero, stacktrace.Propagate(err, "Failed to schedule db stats for %s", dbName)
	}
	c.Start()

	return s, nil
}
