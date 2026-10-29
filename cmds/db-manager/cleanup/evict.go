package cleanup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/interuss/dss/pkg/logging"
	dssmodels "github.com/interuss/dss/pkg/models"
	rids "github.com/interuss/dss/pkg/rid/store"
	scds "github.com/interuss/dss/pkg/scd/store"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/multierr"
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
	_             = flags.String("locality", "", "self-identification string of this DSS instance")
	timeout       = flags.Duration("timeout", 5*time.Minute, "Timeout for the command")
	scdLimit      = flags.Int("scd_limit", 0, "maximum number of SCD entities deleted per entity type (may be exceeded by up to scd_batch_size - 1 when scd_batch_size is set), defaults to unlimited")
	ridLimit      = flags.Int("rid_limit", 0, "maximum number of RID entities deleted per entity type (may be exceeded by up to rid_batch_size - 1 when rid_batch_size is set), defaults to unlimited")
	scdBatchSize  = flags.Int("scd_batch_size", 0, "number of SCD entities deleted per batch query (per entity type), defaults to 0 (single query for deleting entities of a type)")
	ridBatchSize  = flags.Int("rid_batch_size", 0, "number of RID entities deleted per batch query (per entity type), defaults to 0 (single query for deleting entities of a type)")
)

func init() {
	if err := flags.MarkDeprecated("locality", "it has no effect anymore, expired RID entities of all DSS instances are evicted"); err != nil {
		panic(err)
	}
	EvictCmd.Flags().AddFlagSet(flags)
}

func deleteExpiredEntities(
	ctx context.Context,
	deleteFn func(ctx context.Context, threshold time.Time, limit int) ([]dssmodels.ID, error),
	threshold time.Time,
	limit int,
	batchSize int,
) ([]dssmodels.ID, error) {
	if batchSize == 0 {
		return deleteFn(ctx, threshold, limit)
	}
	var deleted []dssmodels.ID
	for {
		batch, err := deleteFn(ctx, threshold, batchSize)
		if err != nil {
			return deleted, err
		}
		deleted = append(deleted, batch...)
		if len(batch) < batchSize || (limit > 0 && len(deleted) >= limit) {
			break
		}
	}
	return deleted, nil
}

func evict(cmd *cobra.Command, _ []string) error {
	var (
		ctx          = cmd.Context()
		scdThreshold = time.Now().Add(-*scdTtl)
		ridThreshold = time.Now().Add(-*ridTtl)
		scdLimit     = *scdLimit
		ridLimit     = *ridLimit
		scdBatchSize = *scdBatchSize
		ridBatchSize = *ridBatchSize
	)
	if scdLimit < 0 {
		return fmt.Errorf("scd_limit must be equal to or greater than 0, got %d", scdLimit)
	}
	if ridLimit < 0 {
		return fmt.Errorf("rid_limit must be equal to or greater than 0, got %d", ridLimit)
	}
	if scdBatchSize < 0 {
		return fmt.Errorf("scd_batch_size must be equal to or greater than 0, got %d", scdBatchSize)
	}
	if ridBatchSize < 0 {
		return fmt.Errorf("rid_batch_size must be equal to or greater than 0, got %d", ridBatchSize)
	}
	for _, deleteOnlyFlag := range []string{"scd_limit", "rid_limit", "scd_batch_size", "rid_batch_size"} {
		if cmd.Flags().Changed(deleteOnlyFlag) && !*deleteExpired {
			return fmt.Errorf("%s can only be set together with --delete", deleteOnlyFlag)
		}
	}
	log.Printf("WARNING: The usage of this tool may have an impact on performance when deleting entities. Read more in the README.")

	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	logger := logging.WithValuesFromContext(ctx, logging.Logger)

	scdStore, err := scds.Init(ctx, logger, false)
	if err != nil {
		return err
	}

	ridStore, err := rids.Init(ctx, logger, false)
	if err != nil {
		return err
	}

	var (
		expiredOpIntents []dssmodels.ID
		scdExpiredSub    []dssmodels.ID
		expiredISAs      []dssmodels.ID
		ridExpiredSub    []dssmodels.ID
		delErr           error
	)
	scdRepo, err := scdStore.Interact(ctx)
	if err != nil {
		return fmt.Errorf("failed to interact with SCD store: %w", err)
	}

	if *checkScdOirs {
		if *deleteExpired {
			expiredOpIntents, err = deleteExpiredEntities(ctx, scdRepo.DeleteExpiredOperationalIntents, scdThreshold, scdLimit, scdBatchSize)
			if err != nil {
				delErr = multierr.Append(delErr, fmt.Errorf("failed to delete expired operational intents: %w", err))
			}
		} else {
			expiredOpIntents, err = scdRepo.ListExpiredOperationalIntents(ctx, scdThreshold)
			if err != nil {
				return fmt.Errorf("failed to list expired operational intents: %w", err)
			}
		}
	}

	if *checkScdSubs {
		if *deleteExpired {
			scdExpiredSub, err = deleteExpiredEntities(ctx, scdRepo.DeleteExpiredSubscriptions, scdThreshold, scdLimit, scdBatchSize)
			if err != nil {
				delErr = multierr.Append(delErr, fmt.Errorf("failed to delete expired SCD subscriptions: %w", err))
			}
		} else {
			scdExpiredSub, err = scdRepo.ListExpiredSubscriptions(ctx, scdThreshold)
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
			expiredISAs, err = deleteExpiredEntities(ctx, ridRepo.DeleteExpiredISAs, ridThreshold, ridLimit, ridBatchSize)
			if err != nil {
				delErr = multierr.Append(delErr, fmt.Errorf("failed to delete expired ISAs: %w", err))
			}
		} else {
			expiredISAs, err = ridRepo.ListExpiredISAs(ctx, ridThreshold)
			if err != nil {
				return fmt.Errorf("failed to list expired ISAs: %w", err)
			}
		}
	}

	if *checkRidSubs {
		if *deleteExpired {
			ridExpiredSub, err = deleteExpiredEntities(ctx, ridRepo.DeleteExpiredSubscriptions, ridThreshold, ridLimit, ridBatchSize)
			if err != nil {
				delErr = multierr.Append(delErr, fmt.Errorf("failed to delete expired RID subscriptions: %w", err))
			}
		} else {
			ridExpiredSub, err = ridRepo.ListExpiredSubscriptions(ctx, ridThreshold)
			if err != nil {
				return fmt.Errorf("failed to list RID expired subscriptions: %w", err)
			}
		}
	}

	action := "found"
	if *deleteExpired {
		action = "deleted"
	}
	for _, id := range expiredOpIntents {
		log.Printf("%s expired operational intent %s", action, id)
	}
	for _, id := range scdExpiredSub {
		log.Printf("%s expired SCD subscription %s", action, id)
	}
	for _, id := range expiredISAs {
		log.Printf("%s expired ISA %s", action, id)
	}
	for _, id := range ridExpiredSub {
		log.Printf("%s expired RID subscription %s", action, id)
	}
	if len(expiredOpIntents) == 0 && len(scdExpiredSub) == 0 && len(expiredISAs) == 0 && len(ridExpiredSub) == 0 {
		log.Printf("no SCD entity older than %s and no RID entity older than %s found", scdThreshold.String(), ridThreshold.String())
	} else if !*deleteExpired {
		log.Printf("no entity was deleted, run the command again with the `--delete` flag to do so")
	}

	if delErr != nil {
		log.Printf("error(s) while deleting entities: %v", delErr)
		return delErr
	}
	return nil
}
