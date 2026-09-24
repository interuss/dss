package cleanup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/interuss/dss/pkg/logging"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	rids "github.com/interuss/dss/pkg/rid/store"
	scdmodels "github.com/interuss/dss/pkg/scd/models"
	scds "github.com/interuss/dss/pkg/scd/store"
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
	scdTtl        = flags.Duration("scd_ttl", time.Hour*24*112, "time-to-live duration used for determining SCD entries expiration, defaults to 2*56 days")
	ridTtl        = flags.Duration("rid_ttl", time.Minute*30, "time-to-live duration used for determining RID entries expiration, defaults to 30 minutes")
	deleteExpired = flags.Bool("delete", false, "set this flag to true to delete the expired entities")
	locality      = flags.String("locality", "", "self-identification string of this DSS instance")
	timeout       = flags.Duration("timeout", 5*time.Minute, "Timeout for the command")
	scdLimit      = flags.Int("scd_limit", 0, "maximum number of SCD entities deleted, defaults to unlimited")
	ridLimit      = flags.Int("rid_limit", 0, "maximum number of RID entities deleted, defaults to unlimited")
)

func init() {
	EvictCmd.Flags().AddFlagSet(flags)
}

func evict(cmd *cobra.Command, _ []string) error {
	var (
		ctx          = cmd.Context()
		scdThreshold = time.Now().Add(-*scdTtl)
		ridThreshold = time.Now().Add(-*ridTtl)
		scdLimit     = *scdLimit
		ridLimit     = *ridLimit
	)
	if scdLimit < 0 {
		return fmt.Errorf("scd_limit must be equal to or greater than 0, got %d", scdLimit)
	}
	if ridLimit < 0 {
		return fmt.Errorf("rid_limit must be equal to or greater than 0, got %d", ridLimit)
	}
	log.Printf("WARNING: The usage of this tool may have an impact on performance when deleting entities. Read more in the README.")

	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	logger := logging.WithValuesFromContext(ctx, logging.Logger)

	scdStore, err := scds.Init(ctx, logger, false, *locality)
	if err != nil {
		return err
	}

	ridStore, err := rids.Init(ctx, logger, false, *locality)
	if err != nil {
		return err
	}

	var (
		expiredOpIntents []*scdmodels.OperationalIntent
		scdExpiredSub    []*scdmodels.Subscription
		expiredISAs      []*ridmodels.IdentificationServiceArea
		ridExpiredSub    []*ridmodels.Subscription
	)
	scdRepo, err := scdStore.Interact(ctx)
	if err != nil {
		return fmt.Errorf("failed to interact with SCD store: %w", err)
	}

	if *checkScdOirs {
		if *deleteExpired {
			expiredOpIntents, err = scdRepo.DeleteExpiredOperationalIntents(ctx, scdThreshold, scdLimit)
			if err != nil {
				return fmt.Errorf("failed to delete expired operational intents: %w", err)
			}
		} else {
			expiredOpIntents, err = scdRepo.ListExpiredOperationalIntents(ctx, scdThreshold, scdLimit)
			if err != nil {
				return fmt.Errorf("failed to list expired operational intents: %w", err)
			}
		}
	}

	if *checkScdSubs {
		if *deleteExpired {
			scdExpiredSub, err = scdRepo.DeleteExpiredSubscriptions(ctx, scdThreshold, scdLimit)
			if err != nil {
				return fmt.Errorf("failed to delete expired SCD subscriptions: %w", err)
			}
		} else {
			scdExpiredSub, err = scdRepo.ListExpiredSubscriptions(ctx, scdThreshold, scdLimit)
			if err != nil {
				return fmt.Errorf("failed to list expired SCD subscriptions: %w", err)
			}
		}
	}

	ridRepo, err := ridStore.Interact(ctx)
	if err != nil {
		return fmt.Errorf("failed to interact with RID store: %w", err)
	}

	if *checkRidISAs {
		if *deleteExpired {
			expiredISAs, err = ridRepo.DeleteExpiredISAs(ctx, *locality, ridThreshold, ridLimit)
			if err != nil {
				return fmt.Errorf("failed to delete expired ISAs: %w", err)
			}
		} else {
			expiredISAs, err = ridRepo.ListExpiredISAs(ctx, *locality, ridThreshold, ridLimit)
			if err != nil {
				return fmt.Errorf("failed to list expired ISAs: %w", err)
			}
		}
	}

	if *checkRidSubs {
		if *deleteExpired {
			ridExpiredSub, err = ridRepo.DeleteExpiredSubscriptions(ctx, *locality, ridThreshold, ridLimit)
			if err != nil {
				return fmt.Errorf("failed to delete expired RID subscriptions: %w", err)
			}
		} else {
			ridExpiredSub, err = ridRepo.ListExpiredSubscriptions(ctx, *locality, ridThreshold, ridLimit)
			if err != nil {
				return fmt.Errorf("failed to list RID expired subscriptions: %w", err)
			}
		}
	}

	for _, opIntent := range expiredOpIntents {
		logExpiredEntity("operational intent", opIntent.ID, scdThreshold, *deleteExpired, opIntent.EndTime != nil)
	}
	for _, sub := range scdExpiredSub {
		logExpiredEntity("SCD subscription", sub.ID, scdThreshold, *deleteExpired, sub.EndTime != nil)
	}
	for _, isa := range expiredISAs {
		logExpiredEntity("ISA", isa.ID, ridThreshold, *deleteExpired, isa.EndTime != nil)
	}
	for _, sub := range ridExpiredSub {
		logExpiredEntity("RID subscription", sub.ID, ridThreshold, *deleteExpired, sub.EndTime != nil)
	}
	if len(expiredOpIntents) == 0 && len(scdExpiredSub) == 0 && len(expiredISAs) == 0 && len(ridExpiredSub) == 0 {
		log.Printf("no SCD entity older than %s and no RID entity older than %s found", scdThreshold.String(), ridThreshold.String())
	} else if !*deleteExpired {
		log.Printf("no entity was deleted, run the command again with the `--delete` flag to do so")
	}
	return nil
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
