package cleanup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/interuss/dss/pkg/datastore"
	crdbflags "github.com/interuss/dss/pkg/datastore/flags"
	"github.com/interuss/dss/pkg/logging"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	ridrepos "github.com/interuss/dss/pkg/rid/repos"
	ridc "github.com/interuss/dss/pkg/rid/store/cockroach"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	scdrepos "github.com/interuss/dss/pkg/scd/repos"
	scdc "github.com/interuss/dss/pkg/scd/store/cockroach"
	"github.com/interuss/stacktrace"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	EvictCmd = &cobra.Command{
		Use:   "evict",
		Short: "List and evict expired entities",
		RunE:  evict,
	}
	flags         = pflag.NewFlagSet("evict", pflag.ExitOnError)
	checkScdOirs  = flags.Bool("scd_oir", true, "set this flag to true to check for expired SCD operational intents")
	checkScdSubs  = flags.Bool("scd_sub", true, "set this flag to true to check for expired SCD subscriptions")
	checkRidISAs  = flags.Bool("rid_isa", true, "set this flag to true to check for expired RID ISAs")
	checkRidSubs  = flags.Bool("rid_sub", true, "set this flag to true to check for expired RID subscriptions")
	ttl           = flags.Duration("ttl", time.Hour*24*112, "time-to-live duration used for determining SCD entries expiration, defaults to 2*56 days which should be a safe value in most cases")
	deleteExpired = flags.Bool("delete", false, "set this flag to true to delete the expired entities")
	locality      = flags.String("locality", "", "self-identification string of this DSS instance")
)

func init() {
	EvictCmd.Flags().AddFlagSet(flags)
}

func evict(cmd *cobra.Command, _ []string) error {
	var (
		ctx       = cmd.Context()
		threshold = time.Now().Add(-*ttl)
	)
	log.Printf("WARNING: The usage of this tool may have an impact on performance when deleting entities. Read more in the README.")

	scdStore, err := getSCDStore(ctx)
	if err != nil {
		return err
	}

	ridStore, err := getRIDStore(ctx)
	if err != nil {
		return err
	}

	var (
		expiredOpIntents []*scdmodels.OperationalIntent
		scdExpiredSub    []*scdmodels.Subscription
		expiredISAs      []*ridmodels.IdentificationServiceArea
		ridExpiredSub    []*ridmodels.Subscription
	)
	scdAction := func(ctx context.Context, r scdrepos.Repository) (err error) {
		if *checkScdOirs {
			expiredOpIntents, err = r.ListExpiredOperationalIntents(ctx, threshold)
			if err != nil {
				return fmt.Errorf("listing expired operational intents: %w", err)
			}
			if *deleteExpired {
				for _, opIntent := range expiredOpIntents {
					if err = r.DeleteOperationalIntent(ctx, opIntent.ID); err != nil {
						return fmt.Errorf("deleting expired operational intents: %w", err)
					}
				}
			}
		}

		if *checkScdSubs {
			scdExpiredSub, err = r.ListExpiredSubscriptions(ctx, threshold)
			if err != nil {
				return fmt.Errorf("SCD listing expired subscriptions: %w", err)
			}
			if *deleteExpired {
				for _, sub := range scdExpiredSub {
					if err = r.DeleteSubscription(ctx, sub.ID); err != nil {
						return fmt.Errorf("SCD deleting expired subscriptions: %w", err)
					}
				}
			}
		}
		return nil
	}
	if err = scdStore.Transact(ctx, scdAction); err != nil {
		return fmt.Errorf("failed to execute SCD transaction: %w", err)
	}

	ridAction := func(r ridrepos.Repository) (err error) {
		if *checkRidISAs {

			expiredISAs, err = r.ListExpiredISAs(ctx, *locality)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to list expired ISAs")
			}

			if *deleteExpired {
				for _, isa := range expiredISAs {
					_, err := r.DeleteISA(ctx, isa)
					if err != nil {
						return stacktrace.Propagate(err, "Failed to delete ISAs")
					}
				}
			}

		}

		if *checkRidSubs {

			ridExpiredSub, err = r.ListExpiredSubscriptions(ctx, *locality)
			if err != nil {
				return stacktrace.Propagate(err,
					"Failed to list RID expired Subscriptions")
			}

			if *deleteExpired {
				for _, sub := range ridExpiredSub {
					_, err := r.DeleteSubscription(ctx, sub)
					if err != nil {
						return stacktrace.Propagate(err,
							"Failed to delete RID Subscription")
					}
				}
			}

		}

		return nil
	}
	if err = ridStore.Transact(ctx, ridAction); err != nil {
		return fmt.Errorf("failed to execute RID transaction: %w", err)
	}

	for _, opIntent := range expiredOpIntents {
		logExpiredEntity("operational intent", opIntent.ID, threshold, *deleteExpired, opIntent.EndTime != nil)
	}
	for _, sub := range scdExpiredSub {
		logExpiredEntity("SCD subscription", sub.ID, threshold, *deleteExpired, sub.EndTime != nil)
	}
	for _, isa := range expiredISAs {
		logExpiredEntity("ISA", isa.ID, time.Now().Add(-time.Duration(ridc.ExpiredDurationInMin)*time.Minute), *deleteExpired, isa.EndTime != nil)
	}
	for _, sub := range ridExpiredSub {
		logExpiredEntity("RID subscription", sub.ID, time.Now().Add(-time.Duration(ridc.ExpiredDurationInMin)*time.Minute), *deleteExpired, sub.EndTime != nil)
	}
	if len(expiredOpIntents) == 0 && len(scdExpiredSub) == 0 && len(expiredISAs) == 0 && len(ridExpiredSub) == 0 {
		log.Printf("no entity older than %s found", threshold.String())
	} else if !*deleteExpired {
		log.Printf("no entity was deleted, run the command again with the `--delete` flag to do so")
	}
	return nil
}

func getSCDStore(ctx context.Context) (*scdc.Store, error) {
	connectParameters := crdbflags.ConnectParameters()
	connectParameters.ApplicationName = "db-manager"
	connectParameters.DBName = scdc.DatabaseName
	scdCrdb, err := datastore.Dial(ctx, connectParameters)
	if err != nil {
		logParams := connectParameters
		logParams.Credentials.Password = "[REDACTED]"
		return nil, fmt.Errorf("failed to connect to strategic conflict detection database with %+v: %w", logParams, err)
	}

	scdStore, err := scdc.NewStore(ctx, scdCrdb)
	if err != nil {
		return nil, fmt.Errorf("failed to create strategic conflict detection store with %+v: %w", connectParameters, err)
	}
	return scdStore, nil
}

func getRIDStore(ctx context.Context) (*ridc.Store, error) {

	logger := logging.WithValuesFromContext(ctx, logging.Logger)

	connectParameters := crdbflags.ConnectParameters()
	connectParameters.ApplicationName = "db-manager"
	connectParameters.DBName = "rid"
	ridCrdb, err := datastore.Dial(ctx, connectParameters)
	if err != nil {
		logParams := connectParameters
		logParams.Credentials.Password = "[REDACTED]"
		return nil, fmt.Errorf("failed to connect to remote ID database with %+v: %w", logParams, err)
	}

	ridStore, err := ridc.NewStore(ctx, ridCrdb, connectParameters.DBName, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote ID store with %+v: %w", connectParameters, err)
	}
	return ridStore, nil
}

func logExpiredEntity(entity string, entityID dssmodels.ID, threshold time.Time, deleted, hasEndTime bool) {
	logMsg := "found"
	if deleted {
		logMsg = "deleted"
	}

	expMsg := "last update before %s (missing end time)"
	if hasEndTime {
		expMsg = "end time before %s"
	}
	log.Printf("%s %s %s; expired due to %s", logMsg, entity, entityID.String(), fmt.Sprintf(expMsg, threshold.String()))
}
