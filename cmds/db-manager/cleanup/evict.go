package cleanup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/interuss/dss/pkg/logging"
	dssmodels "github.com/interuss/dss/pkg/models"
	rids "github.com/interuss/dss/pkg/rid/store"
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
	outputFile    = flags.String("output", "", "file to write the IDs of the expired entities to, as JSON, can only be set without --delete")
	inputFile     = flags.String("input", "", "file with the IDs of the entities to delete. All the listed entities are deleted, whether they are expired or not, can only be set together with --delete")
)

// expiredEntities is the content of the --output and --input files.
type expiredEntities struct {
	OperationalIntents []dssmodels.ID `json:"operational_intents"`
	SCDSubscriptions   []dssmodels.ID `json:"scd_subscriptions"`
	ISAs               []dssmodels.ID `json:"rid_isas"`
	RIDSubscriptions   []dssmodels.ID `json:"rid_subscriptions"`
}

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
	for _, limitFlag := range []string{"scd_limit", "rid_limit"} {
		if cmd.Flags().Changed(limitFlag) && !*deleteExpired {
			return fmt.Errorf("%s can only be set together with --delete", limitFlag)
		}
		if cmd.Flags().Changed(limitFlag) && *inputFile != "" {
			return fmt.Errorf("%s cannot be set together with --input", limitFlag)
		}
	}
	for _, ttlFlag := range []string{"scd_ttl", "rid_ttl"} {
		if cmd.Flags().Changed(ttlFlag) && *inputFile != "" {
			return fmt.Errorf("%s cannot be set together with --input", ttlFlag)
		}
	}
	if *outputFile != "" && *deleteExpired {
		return fmt.Errorf("output can only be set without --delete")
	}
	if *inputFile != "" && !*deleteExpired {
		return fmt.Errorf("input can only be set together with --delete")
	}

	// When an input file is given, only the entities it lists are deleted.
	var input *expiredEntities
	if *inputFile != "" {
		data, err := os.ReadFile(*inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
		input = &expiredEntities{}
		if err := json.Unmarshal(data, input); err != nil {
			return fmt.Errorf("failed to parse input file: %w", err)
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
	)
	scdRepo, err := scdStore.Interact(ctx)
	if err != nil {
		return fmt.Errorf("failed to interact with SCD store: %w", err)
	}

	if *checkScdOirs {
		if *deleteExpired {
			if input != nil {
				expiredOpIntents, err = scdRepo.DeleteOperationalIntentsByIDs(ctx, input.OperationalIntents)
			} else {
				expiredOpIntents, err = scdRepo.DeleteExpiredOperationalIntents(ctx, scdThreshold, scdLimit)
			}
			if err != nil {
				return fmt.Errorf("failed to delete expired operational intents: %w", err)
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
			if input != nil {
				scdExpiredSub, err = scdRepo.DeleteSubscriptionsByIDs(ctx, input.SCDSubscriptions)
			} else {
				scdExpiredSub, err = scdRepo.DeleteExpiredSubscriptions(ctx, scdThreshold, scdLimit)
			}
			if err != nil {
				return fmt.Errorf("failed to delete expired SCD subscriptions: %w", err)
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
			if input != nil {
				expiredISAs, err = ridRepo.DeleteISAsByIDs(ctx, input.ISAs)
			} else {
				expiredISAs, err = ridRepo.DeleteExpiredISAs(ctx, *locality, ridThreshold, ridLimit)
			}
			if err != nil {
				return fmt.Errorf("failed to delete expired ISAs: %w", err)
			}
		} else {
			expiredISAs, err = ridRepo.ListExpiredISAs(ctx, *locality, ridThreshold)
			if err != nil {
				return fmt.Errorf("failed to list expired ISAs: %w", err)
			}
		}
	}

	if *checkRidSubs {
		if *deleteExpired {
			if input != nil {
				ridExpiredSub, err = ridRepo.DeleteSubscriptionsByIDs(ctx, input.RIDSubscriptions)
			} else {
				ridExpiredSub, err = ridRepo.DeleteExpiredSubscriptions(ctx, *locality, ridThreshold, ridLimit)
			}
			if err != nil {
				return fmt.Errorf("failed to delete expired RID subscriptions: %w", err)
			}
		} else {
			ridExpiredSub, err = ridRepo.ListExpiredSubscriptions(ctx, *locality, ridThreshold)
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

	if *outputFile != "" {
		data, err := json.Marshal(expiredEntities{
			OperationalIntents: expiredOpIntents,
			SCDSubscriptions:   scdExpiredSub,
			ISAs:               expiredISAs,
			RIDSubscriptions:   ridExpiredSub,
		})
		if err != nil {
			return fmt.Errorf("failed to encode output file: %w", err)
		}
		if err := os.WriteFile(*outputFile, data, 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		log.Printf("wrote the IDs of the expired entities to %s", *outputFile)
	}

	if len(expiredOpIntents) == 0 && len(scdExpiredSub) == 0 && len(expiredISAs) == 0 && len(ridExpiredSub) == 0 {
		log.Printf("no SCD entity older than %s and no RID entity older than %s found", scdThreshold.String(), ridThreshold.String())
	} else if !*deleteExpired && *outputFile != "" {
		log.Printf("no entity was deleted, run the command again with `--delete --input %s` to delete the listed entities", *outputFile)
	} else if !*deleteExpired {
		log.Printf("no entity was deleted, run the command again with the `--delete` flag to do so")
	}
	return nil
}
