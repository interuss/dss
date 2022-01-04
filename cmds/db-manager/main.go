// Script for Database bootstrap deployment and migration

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/cockroach/flags"
	"github.com/interuss/stacktrace"
	_ "github.com/lib/pq"
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
	path      = flag.String("schemas_dir", "", "path to db migration files directory. the migrations found there will be applied to the database whose name matches the folder name.")
	dbVersion = flag.String("db_version", "", "the db version to migrate to (ex: 1.0.0) or use \"latest\" to automatically upgrade to the latest version or leave blank to print the current version")
)

func main() {
	// Read and validate schemas_dir input
	flag.Parse()
	if *path == "" {
		log.Panic("Must specify schemas_dir path")
	}
	dbName := filepath.Base(*path)

	// Enumerate schema versions
	steps, err := enumerateMigrationSteps(path)
	if err != nil {
		log.Panicf("Failed to read schema version migration definitions: %v", err)
	}
	if len(steps) == 0 {
		log.Panicf("No migration definitions found in schemas_dir=%s", *path)
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
			log.Panicf("Failed to parse desired db_version: %v", err)
		}
	}

	// Connect to database server
	connectParameters := flags.ConnectParameters()
	connectParameters.ApplicationName = "db-manager"
	connectParameters.DBName = "postgres" // Use an initial database that is known to always be present
	crdb, err := cockroach.ConnectTo(connectParameters)
	if err != nil {
		log.Panicf("Failed to connect to database with %+v: %v", connectParameters, err)
	}
	defer func() {
		crdb.Close()
	}()

	// Make sure specified database exists
	exists, err := doesDatabaseExist(crdb, dbName)
	if err != nil {
		log.Panicf("Failed to check whether database %s exists: %v", dbName, err)
	}
	if !exists && dbName == "rid" {
		// In the special case of rid, the database was previously named defaultdb
		log.Printf("Database %s does not exist; checking for older \"defaultdb\" database", dbName)
		dbName = "defaultdb"
		exists, err = doesDatabaseExist(crdb, dbName)
		if err != nil {
			log.Panicf("Failed to check whether old defaultdb database exists: %v", err)
		}
	}
	if !exists {
		log.Printf("Database %s does not exist; creating now", dbName)
		createDB := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)
		if _, err := crdb.Exec(createDB); err != nil {
			log.Panicf("Failed to create new database %s: %v", dbName, err)
		}
	} else {
		log.Printf("Database %s already exists; reading current state", dbName)
	}

	// Read current schema version of database
	currentVersion, err := crdb.GetVersion(context.Background(), dbName)
	if err != nil {
		log.Panicf("Failed to get current database version for %s: %v", dbName, err)
	}
	log.Printf("Initial %s database schema version is %v, target is %v", dbName, currentVersion, targetVersion)
	if targetVersion == nil {
		return
	}

	// Compute index of current version
	var currentStepIndex int = -1
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
		rawMigrationSQL, err := ioutil.ReadFile(fullFilePath)
		if err != nil {
			log.Panicf("Failed to load SQL content from %s: %v", fullFilePath, err)
		}
		migrationSQL := fmt.Sprintf("USE %s;\n", dbName) + string(rawMigrationSQL)

		// Execute migration step
		if _, err := crdb.Exec(migrationSQL); err != nil {
			log.Panicf("Failed to execute %s migration step %s: %v", dbName, fullFilePath, err)
		}

		// Update current state
		if dbName == "defaultdb" && newVersion.String() == "4.0.0" && newCurrentStepIndex > currentStepIndex {
			// RID database changes from `defaultdb` to `rid` when moving up to 4.0.0
			dbName = "rid"
		}
		if dbName == "defaultdb" && currentVersion.String() == "4.0.0" && newCurrentStepIndex < currentStepIndex {
			// RID database changes from `rid` to `defaultdb` when moving down from 4.0.0
			dbName = "defaultdb"
		}
		actualVersion, err := crdb.GetVersion(context.Background(), dbName)
		if err != nil {
			log.Panicf("Failed to get current database version for %s: %v", dbName, err)
		}
		if !actualVersion.Equal(*newVersion) {
			log.Panicf("Migration %s should have migrated %s schema version %v to %v, but instead resulted in %v", fullFilePath, dbName, currentVersion, newVersion, currentVersion)
		}
		currentVersion = actualVersion
		currentStepIndex = newCurrentStepIndex
	}

	log.Printf("Final %s version: %v", dbName, currentVersion)
}

func enumerateMigrationSteps(path *string) ([]MigrationStep, error) {
	steps := make(map[semver.Version]MigrationStep)

	// Identify files defining version migration steps
	files, err := ioutil.ReadDir(*path)
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
				if match[1] == "upto" {
					step.upToFile = file.Name()
				} else if match[1] == "downfrom" {
					step.downFromFile = file.Name()
				} else {
					return make([]MigrationStep, 0), fmt.Errorf("Unexpected migration step prefix: %s", match[1])
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

func doesDatabaseExist(crdb *cockroach.DB, database string) (bool, error) {
	const checkDbQuery = `
		SELECT EXISTS (
			SELECT * FROM pg_database WHERE datname = $1
		)`

	var exists bool
	if err := crdb.QueryRow(checkDbQuery, database).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
