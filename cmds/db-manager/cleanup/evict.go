package cleanup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/interuss/dss/pkg/datastore"
	crdbflags "github.com/interuss/dss/pkg/datastore/flags"
	dssmodels "github.com/interuss/dss/pkg/models"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	"github.com/interuss/dss/pkg/scd/repos"
	scdc "github.com/interuss/dss/pkg/scd/store/cockroach"
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
	listScdOirs   = flags.Bool("scd_oir", true, "set this flag to true to list expired SCD operational intents")
	listScdSubs   = flags.Bool("scd_sub", true, "set this flag to true to list expired SCD subscriptions")
	ttl           = flags.Duration("ttl", time.Hour*24*112, "time-to-live duration used for determining expiration, defaults to 2*56 days which should be a safe value in most cases")
	deleteExpired = flags.Bool("delete", false, "set this flag to true to delete the expired entities")
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

	var (
		expiredOpIntents []*scdmodels.OperationalIntent
		expiredSubs      []*scdmodels.Subscription
	)
	action := func(ctx context.Context, r repos.Repository) (err error) {
		if *listScdOirs {
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

		if *listScdSubs {
			expiredSubs, err = r.ListExpiredSubscriptions(ctx, threshold)
			if err != nil {
				return fmt.Errorf("listing expired subscriptions: %w", err)
			}
			if *deleteExpired {
				for _, sub := range expiredSubs {
					if err = r.DeleteSubscription(ctx, sub.ID); err != nil {
						return fmt.Errorf("deleting expired subscriptions: %w", err)
					}
				}
			}
		}

		return nil
	}
	if err = scdStore.Transact(ctx, action); err != nil {
		return fmt.Errorf("failed to execute CRDB transaction: %w", err)
	}

	for _, opIntent := range expiredOpIntents {
		logExpiredEntity("operational intent", opIntent.ID, threshold, *deleteExpired, opIntent.EndTime != nil)
	}
	for _, sub := range expiredSubs {
		logExpiredEntity("subscription", sub.ID, threshold, *deleteExpired, sub.EndTime != nil)
	}
	if len(expiredOpIntents) == 0 && len(expiredSubs) == 0 {
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
		return nil, fmt.Errorf("failed to connect to database with %+v: %w", logParams, err)
	}

	scdStore, err := scdc.NewStore(ctx, scdCrdb)
	if err != nil {
		return nil, fmt.Errorf("failed to create strategic conflict detection store with %+v: %w", connectParameters, err)
	}
	return scdStore, nil
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
