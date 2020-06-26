// Bootstrap script for Database deployment and migration

package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	dssCockroach "github.com/interuss/dss/pkg/cockroach"
	"go.uber.org/zap"
)

// MyMigrate is an alias for exstending migrate.Migrate
type MyMigrate struct {
	*migrate.Migrate
}

// Direction is an alias for int indicating the direction and steps of migration
type Direction int

func (d Direction) String() string {
	if d > 0 {
		return "Up"
	} else if d < 0 {
		return "Down"
	}
	return "No Change"
}

var (
	path      = flag.String("schemas_dir", "", "path to db migration files directory")
	dbVersion = flag.String("db_version", "", "the db version to migrate to (ex: v1.0.0)")
	step      = flag.Int("migration_step", 0, "the db migration step to go to")

	cockroachParams = struct {
		host    *string
		port    *int
		sslMode *string
		sslDir  *string
		user    *string
	}{
		host:    flag.String("cockroach_host", "", "cockroach host to connect to"),
		port:    flag.Int("cockroach_port", 26257, "cockroach port to connect to"),
		sslMode: flag.String("cockroach_ssl_mode", "disable", "cockroach sslmode"),
		user:    flag.String("cockroach_user", "root", "cockroach user to authenticate as"),
		sslDir:  flag.String("cockroach_ssl_dir", "", "directory to ssl certificates. Must contain files: ca.crt, client.<user>.crt, client.<user>.key"),
	}
)

func main() {
	flag.Parse()
	if *path == "" {
		log.Panic("Must specify schemas_dir path")
	}
	if (*dbVersion == "" && *step == 0) || (*dbVersion != "" && *step != 0) {
		log.Panic("Must specify a db_version or migration_step to goto")
	}
	uriParams := map[string]string{
		"host":             *cockroachParams.host,
		"port":             strconv.Itoa(*cockroachParams.port),
		"user":             *cockroachParams.user,
		"ssl_mode":         *cockroachParams.sslMode,
		"ssl_dir":          *cockroachParams.sslDir,
		"application_name": "SchemaManager",
		"db_name":          filepath.Base(*path),
	}
	postgresURI, err := dssCockroach.BuildURI(uriParams)
	if err != nil {
		log.Panic("Failed to build URI", zap.Error(err))
	}
	crdbURI := strings.Replace(postgresURI, "postgresql", "cockroachdb", 1)
	*path = fmt.Sprintf("file://%v", *path)
	migrater, migraterErr := migrate.New(*path, crdbURI)
	if migraterErr != nil {
		log.Panic(migraterErr)
	}
	myMigrater := &MyMigrate{migrater}
	defer func() {
		if migraterErr == nil {
			if _, err := myMigrater.Close(); err != nil {
				log.Println(err)
			}
		}
	}()
	myMigrater.PrefetchMigrations = 10
	myMigrater.LockTimeout = time.Duration(15) * time.Second

	// handle Ctrl+c
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	go func() {
		for range signals {
			log.Println("Stopping after this running migration ...")
			myMigrater.GracefulStop <- true
			return
		}
	}()
	totalMove := 0
	migrateDirection, err := myMigrater.MigrationDirection(postgresURI, *dbVersion, *step)
	if err != nil {
		log.Panic(err)
	}
	preMigrationStep, _, err := myMigrater.Version()
	if err != nil {
		log.Panic(err)
	}
	for migrateDirection != 0 {
		migraterErr = myMigrater.Steps(int(migrateDirection))
		if migraterErr != nil {
			log.Fatal(migraterErr)
		}
		totalMove += int(migrateDirection)
		log.Printf("Migrated %s by %d step", migrateDirection.String(), intAbs(int(migrateDirection)))
		migrateDirection, err = myMigrater.MigrationDirection(postgresURI, *dbVersion, *step)
		if err != nil {
			log.Panic(err)
		}
	}
	postMigrationStep, dirty, err := myMigrater.Version()
	if err != nil {
		log.Fatal("Failed to get Migration Step for confirmation")
	}
	if totalMove == 0 {
		log.Println("No Changes")
	} else {
		log.Printf("Moved %d step(s) in total from Step %d to Step %d", intAbs(totalMove), preMigrationStep, postMigrationStep)
	}

	currentDBVersion, err := getCurrentDBVersion(postgresURI)
	if err != nil {
		log.Fatal("Failed to get Current DB version for confirmation")
	}
	log.Printf("DB Version: %s, Migration Step # %d, Dirty: %v", currentDBVersion, postMigrationStep, dirty)
}

func intAbs(x int) int {
	return int(math.Abs(float64(x)))
}

func getCurrentDBVersion(crdbURI string) (string, error) {
	crdb, err := dssCockroach.Dial(crdbURI)
	defer func() {
		crdb.Close()
	}()
	if err != nil {
		return "", fmt.Errorf("Failed to dial CRDB while getting DB version: %v", err)
	}
	// check if schema_versions table exists
	const checkTableQuery = `
		SELECT EXISTS (
  		SELECT *
		  FROM information_schema.tables 
		WHERE table_name = 'schema_versions'
	)`
	var (
		version = "v0.0.0"
		exists  bool
	)

	if err := crdb.QueryRow(checkTableQuery).Scan(&exists); err != nil {
		return "", err
	}

	if !exists {
		return version, nil
	}
	// query for the schema version string
	const getVersionQuery = `
      SELECT schema_version 
      FROM schema_versions
	  WHERE onerow_enforcer = TRUE`
	if err := crdb.QueryRow(getVersionQuery).Scan(&version); err != nil {
		return "", err
	}
	// if for some reason the string returned is empty
	if version == "" {
		version = "v0.0.0"
	}
	return version, nil
}

func parseVersion(version string) [3]int {
	var (
		major, minor, patch int
	)
	if n, err := fmt.Sscanf(version, "v%d.%d.%d", &major, &minor, &patch); err != nil || n != 3 {
		log.Panic(err)
	}
	result := [3]int{major, minor, patch}
	return result
}

// MigrationDirection reads our custom DB version string as well as the Migration Steps from the framework
// and returns a signed integer value of the Direction and count to migrate the db
func (m *MyMigrate) MigrationDirection(dbURI string, desiredVersion string, desiredStep int) (Direction, error) {
	if desiredStep != 0 {
		currentStep, dirty, err := m.Version()
		if err != migrate.ErrNilVersion && err != nil {
			return 0, fmt.Errorf("Failed to get Migration Step to determine migration direction: %v", err)
		}
		if dirty {
			log.Fatal("DB in Dirty state, Please fix before migrating")
		}
		return Direction(desiredStep - int(currentStep)), nil
	}
	currentVersion, err := getCurrentDBVersion(dbURI)
	if err != nil {
		return 0, fmt.Errorf("Failed to get current DB version to determine migration direction: %v", err)
	}
	var (
		current     = parseVersion(currentVersion)
		destination = parseVersion(desiredVersion)
	)
	for i := 0; i < 3; i++ {
		diff := destination[0] - current[0]
		if diff > 0 {
			return 1, nil
		} else if diff < 0 {
			return -1, nil
		}
	}
	return 0, nil
}
