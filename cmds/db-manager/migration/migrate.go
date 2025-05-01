package migration

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/interuss/dss/pkg/datastore"
	crdbflags "github.com/interuss/dss/pkg/datastore/flags"

	"github.com/interuss/stacktrace"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type MigrationStep struct {
	version      semver.Version
	upToFile     string
	downFromFile string
}

var (
	// Pattern to match files describing migration steps
	migrationStepRegexp = "(upto|downfrom)-v(\\d+\\.\\d+\\.\\d+)-(.*)\\.sql"
)

var (
	MigrationCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Database bootstrap deployment and migration",
		RunE:  migrate,
	}
	flags     = pflag.NewFlagSet("migrate", pflag.ExitOnError)
	path      = flags.String("schemas_dir", "", "path to db migration files directory. the migrations found there will be applied to the database whose name matches the folder name.")
	dbVersion = flags.String("db_version", "", "the db version to migrate to (ex: 1.0.0) or use \"latest\" to automatically upgrade to the latest version or leave blank to print the current version")
)

func init() {
	MigrationCmd.Flags().AddFlagSet(flags)
	_ = MigrationCmd.MarkFlagRequired("schemas_dir")
}

func migrate(cmd *cobra.Command, _ []string) error {
	var (
		ctx    = cmd.Context()
		dbName = filepath.Base(*path)
	)

	// Enumerate schema versions
	steps, err := enumerateMigrationSteps(path)
	if err != nil {
		return fmt.Errorf("failed to read schema version migration definitions: %w", err)
	}
	if len(steps) == 0 {
		return fmt.Errorf("no migration definitions found in schemas_dir=%s", *path)
	}

	// Determine target version
	var targetVersion *semver.Version
	if strings.ToLower(*dbVersion) == "latest" {
		targetVersion = &steps[len(steps)-1].version
	} else if strings.TrimSpace(*dbVersion) == "" {
		// User just wants to print the current version
		targetVersion = nil
	} else {
		targetVersion, err = semver.NewVersion(*dbVersion)
		if err != nil {
			return fmt.Errorf("failed to parse desired db_version: %w", err)
		}
	}

	sysDbName := "postgres" // Use an initial database that is known to always be present
	ds, err := connectTo(ctx, sysDbName)
	if err != nil {
		return fmt.Errorf("failed to connect to database %s: %w", sysDbName, err)
	}
	defer func() {
		ds.Pool.Close()
	}()

	log.Printf("Datastore server type and version: %s@%s", ds.Version.Type, ds.Version.SemVer.String())

	var (
		isCockroach = ds.Version.Type == datastore.CockroachDB
		isYugabyte  = ds.Version.Type == datastore.Yugabyte
	)

	// Make sure specified database exists
	exists, err := ds.DatabaseExists(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to check whether database %s exists: %w", dbName, err)
	}
	if isCockroach && !exists && dbName == "rid" {
		// In the special case of rid, the database was previously named defaultdb
		log.Printf("Database %s does not exist; checking for older \"defaultdb\" database", dbName)
		dbName = "defaultdb"
		exists, err = ds.DatabaseExists(ctx, dbName)
		if err != nil {
			return fmt.Errorf("failed to check whether old defaultdb database exists: %w", err)
		}
	}
	if !exists {
		log.Printf("Database %s does not exist; creating now", dbName)
		createDB := fmt.Sprintf("CREATE DATABASE %s", dbName)
		if _, err := ds.Pool.Exec(ctx, createDB); err != nil {
			return fmt.Errorf("failed to create new database %s: %v", dbName, err)
		}
	} else {
		log.Printf("Database %s already exists; reading current state", dbName)
	}

	// Reconnect to target database
	ds2, err := connectTo(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to reconnect to database %s: %w", dbName, err)
	}
	defer func() {
		ds2.Pool.Close()
	}()

	// Read current schema version of database
	currentVersion, err := ds2.GetSchemaVersion(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to get current database version for %s: %w", dbName, err)
	}
	log.Printf("Initial %s database schema version is %v, target is %v", dbName, currentVersion, targetVersion)
	if targetVersion == nil {
		return nil
	}

	// Compute index of current version
	var currentStepIndex = -1
	for i, version := range steps {
		if version.version == *currentVersion {
			currentStepIndex = i
		}
	}

	// Perform migration steps until current version matches target version
	for !currentVersion.Equal(*targetVersion) {
		// Compute which migration step to run next and how it will change the schema version
		var newCurrentStepIndex int
		var sqlFile string
		var newVersion *semver.Version
		if currentVersion.LessThan(*targetVersion) {
			// Migrate up to next version
			sqlFile = steps[currentStepIndex+1].upToFile
			newVersion = &steps[currentStepIndex+1].version
			newCurrentStepIndex = currentStepIndex + 1
		} else {
			// Migrate down from current version
			sqlFile = steps[currentStepIndex].downFromFile
			newCurrentStepIndex = currentStepIndex - 1
			newVersion = &steps[newCurrentStepIndex].version
		}
		log.Printf("Running %s to migrate %v to %v", sqlFile, currentVersion, newVersion)

		// Read migration SQL into string
		fullFilePath := filepath.Join(*path, sqlFile)
		rawMigrationSQL, err := os.ReadFile(fullFilePath)
		if err != nil {
			return fmt.Errorf("failed to load SQL content from %s: %e", fullFilePath, err)
		}

		migrationSQL := ""
		if isCockroach {
			// Ensure SQL session has implicit transactions disabled for CRDB versions 22.2+
			sessionConfigurationSQL := ""
			if ds2.Version.SemVer.Compare(*semver.New("22.2.0")) >= 0 {
				sessionConfigurationSQL = "SET enable_implicit_transaction_for_batch_statements = false;\n"
			}

			migrationSQL = sessionConfigurationSQL + fmt.Sprintf("USE %s;\n", dbName) + string(rawMigrationSQL)
		}
		if isYugabyte {
			// Migrations do not require database switch in opposite to CRDB.
			migrationSQL = string(rawMigrationSQL)
		}

		// Execute migration step
		if _, err := ds2.Pool.Exec(ctx, migrationSQL); err != nil {
			return fmt.Errorf("failed to execute %s migration step %s: %w", dbName, fullFilePath, err)
		}

		// Update current state
		if isCockroach {
			// Update current state for CRDB
			if dbName == "defaultdb" && newVersion.String() == "4.0.0" && newCurrentStepIndex > currentStepIndex {
				// RID database changes from `defaultdb` to `rid` when moving up to 4.0.0
				dbName = "rid"
			}
			if dbName == "rid" && currentVersion.String() == "4.0.0" && newCurrentStepIndex < currentStepIndex {
				// RID database changes from `rid` to `defaultdb` when moving down from 4.0.0
				dbName = "defaultdb"
			}
		}
		actualVersion, err := ds2.GetSchemaVersion(ctx, dbName)
		if err != nil {
			return fmt.Errorf("failed to get current database version for %s: %w", dbName, err)
		}
		if !actualVersion.Equal(*newVersion) {
			return fmt.Errorf("migration %s should have migrated %s schema version %v to %v, but instead resulted in %v", fullFilePath, dbName, currentVersion, newVersion, currentVersion)
		}
		currentVersion = actualVersion
		currentStepIndex = newCurrentStepIndex
	}

	log.Printf("Final %s version: %v", dbName, currentVersion)
	return nil
}

func connectTo(ctx context.Context, dbName string) (*datastore.Datastore, error) {
	// Connect to database server
	connectParameters := crdbflags.ConnectParameters()
	connectParameters.ApplicationName = "db-manager"
	connectParameters.DBName = dbName
	return datastore.Dial(ctx, connectParameters)
}

func enumerateMigrationSteps(path *string) ([]MigrationStep, error) {
	steps := make(map[semver.Version]MigrationStep)

	// Identify files defining version migration steps
	files, err := os.ReadDir(*path)
	if err != nil {
		return make([]MigrationStep, 0), stacktrace.Propagate(err, "Failed to read schema files directory")
	}
	r := regexp.MustCompile(migrationStepRegexp)
	for _, file := range files {
		if !file.IsDir() {
			match := r.FindStringSubmatch(file.Name())
			if len(match) > 0 {
				v := *semver.New(match[2])
				step := steps[v]
				step.version = v
				switch match[1] {
				case "upto":
					step.upToFile = file.Name()
				case "downfrom":
					step.downFromFile = file.Name()
				default:
					return make([]MigrationStep, 0), fmt.Errorf("unexpected migration step prefix: %s", match[1])
				}
				steps[v] = step
			}
		}
	}

	// Sort versions in ascending order
	versions := make([]*semver.Version, len(steps))
	i := 0
	for k := range steps {
		v := steps[k].version
		versions[i] = &v
		i++
	}
	semver.Sort(versions)

	// Render sorted step list
	result := make([]MigrationStep, len(versions)+1)
	result[0].version = *semver.New("0.0.0")
	for i := 0; i < len(versions); i++ {
		result[i+1] = steps[*versions[i]]
	}

	return result, nil
}
