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

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	dssCockroach "github.com/interuss/dss/pkg/cockroach"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"
)

// MyMigrate is an alias for extending migrate.Migrate
type MyMigrate struct {
	*migrate.Migrate
	postgresURI string
	database    string
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
	path      = flag.String("schemas_dir", "", "path to db migration files directory. the migrations found there will be applied to the database whose name matches the folder name.")
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
	if *dbVersion != "" && !semver.IsValid(*dbVersion) {
		log.Panic("db_version must be in a valid format ex: v1.2.3")
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
	myMigrater, err := New(*path, postgresURI, uriParams["db_name"])
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		if _, err := myMigrater.Close(); err != nil {
			log.Println(err)
		}
	}()
	preMigrationStep, _, err := myMigrater.Version()
	if err != migrate.ErrNilVersion && err != nil {
		log.Panic(err)
	}
	totalMoves, err := myMigrater.DoMigrate(*dbVersion, *step)
	if err != nil {
		log.Panic(err)
	}
	postMigrationStep, dirty, err := myMigrater.Version()
	if err != nil {
		log.Fatal("Failed to get Migration Step for confirmation")
	}
	if totalMoves == 0 {
		log.Println("No Changes")
	} else {
		log.Printf("Moved %d step(s) in total from Step %d to Step %d", intAbs(totalMoves), preMigrationStep, postMigrationStep)
	}

	currentDBVersion, err := getCurrentDBVersion(postgresURI, uriParams["db_name"])
	if err != nil {
		log.Fatal("Failed to get Current DB version for confirmation")
	}
	log.Printf("DB Version: %s, Migration Step # %d, Dirty: %v", currentDBVersion, postMigrationStep, dirty)
}

// DoMigrate performs the migration given the desired state we want to reach
func (m *MyMigrate) DoMigrate(desiredDBVersion string, desiredStep int) (int, error) {
	totalMoves := 0
	migrateDirection, err := m.MigrationDirection(desiredDBVersion, desiredStep)
	if err != nil {
		return totalMoves, err
	}
	for migrateDirection != 0 {
		err = m.Steps(int(migrateDirection))
		if err != nil {
			return totalMoves, err
		}
		totalMoves += int(migrateDirection)
		log.Printf("Migrated %s by %d step", migrateDirection.String(), intAbs(int(migrateDirection)))
		migrateDirection, err = m.MigrationDirection(desiredDBVersion, *step)
		if err != nil {
			return totalMoves, err
		}
	}
	return totalMoves, nil
}

// New instantiates a new migrate object
func New(path string, dbURI string, database string) (*MyMigrate, error) {
	noDbPostgres := strings.Replace(dbURI, fmt.Sprintf("/%s", database), "", 1)
	err := createDatabaseIfNotExists(noDbPostgres, database)
	if err != nil {
		return nil, err
	}
	path = fmt.Sprintf("file://%v", path)
	crdbURI := strings.Replace(dbURI, "postgresql", "cockroachdb", 1)
	migrater, err := migrate.New(path, crdbURI)
	if err != nil {
		return nil, err
	}
	myMigrater := &MyMigrate{migrater, dbURI, database}
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
	return myMigrater, err
}

func intAbs(x int) int {
	return int(math.Abs(float64(x)))
}

func createDatabaseIfNotExists(crdbURI string, database string) error {
	crdb, err := dssCockroach.Dial(crdbURI)
	defer func() {
		crdb.Close()
	}()
	if err != nil {
		return fmt.Errorf("Failed to dial CRDB to check DB exists: %v", err)
	}
	const checkDbQuery = `
		SELECT EXISTS (
			SELECT *
				FROM pg_database 
			WHERE datname = $1
		)
	`

	var exists bool

	if err := crdb.QueryRow(checkDbQuery, database).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		log.Printf("Database \"%s\" doesn't exists, attempt to create", database)
		createDB := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database)
		_, err := crdb.Exec(createDB)
		if err != nil {
			return fmt.Errorf("Failed to Create Database: %v", err)
		}
	}
	return nil
}

func getCurrentDBVersion(crdbURI string, database string) (string, error) {
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
		AND table_catalog = $1
		)
	`
	var (
		version = "v0.0.0"
		exists  bool
	)

	if err := crdb.QueryRow(checkTableQuery, database).Scan(&exists); err != nil {
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

// MigrationDirection reads our custom DB version string as well as the Migration Steps from the framework
// and returns a signed integer value of the Direction and count to migrate the db
func (m *MyMigrate) MigrationDirection(desiredVersion string, desiredStep int) (Direction, error) {
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
	currentVersion, err := getCurrentDBVersion(m.postgresURI, m.database)
	if err != nil {
		return 0, fmt.Errorf("Failed to get current DB version to determine migration direction: %v", err)
	}
	if !semver.IsValid(currentVersion) {
		return 0, fmt.Errorf("The current DB Version format is in valid")
	}
	return Direction(semver.Compare(desiredVersion, currentVersion)), nil
}
