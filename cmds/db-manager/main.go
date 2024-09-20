package main

import (
	"flag"
	"log"
	"os"

	"github.com/interuss/dss/cmds/db-manager/migration"
	"github.com/spf13/cobra"
)

var (
	DBManagerCmd = &cobra.Command{
		Use:   "db-manager",
		Short: "DSS database management utility",
	}
)

func init() {
	DBManagerCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine) // enable support for flags not yet migrated to using pflag (e.g. crdb flags)
	DBManagerCmd.AddCommand(migration.MigrationCmd)
}

func main() {
	if err := DBManagerCmd.Execute(); err != nil {
		log.Printf("failed to execute db-manager: %v", err)
		os.Exit(1)
	}
}
