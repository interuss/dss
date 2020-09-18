// Bootstrap script for Database deployment and migration

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/coreos/go-semver/semver"
	"github.com/golang-migrate/migrate/v4"
	"github.com/interuss/dss/pkg/cockroach"
	"github.com/interuss/dss/pkg/cockroach/flags"
	"go.uber.org/zap"

	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb" // Force registration of cockroachdb backend
	_ "github.com/golang-migrate/migrate/v4/source/file"          // Force registration of file source
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
	dbVersion = flag.String("db_version", "", "the db version to migrate to (ex: 1.0.0)")
	step      = flag.Int("migration_step", 0, "the db migration step to go to")
	latest    = flag.Bool("latest", false, "migrate db to the latest availiable step in path specified")
)

func main() {
	flag.Parse()
	if *path == "" {
		log.Panic("Must specify schemas_dir path")
	}
	if (*dbVersion == "" && *step == 0 && *latest == false) || (*dbVersion != "" && *step != 0 && *latest == true) {
		log.Panic("Must specify one of [db_version, migration_step, latest] to goto")
	}

	var (
		desiredVersion *semver.Version
	)

	if *dbVersion != "" {
		if v, err := semver.NewVersion(*dbVersion); err == nil {
			desiredVersion = v
		} else {
			log.Panic("db_version must be in a valid format ex: 1.2.3", err)
		}
	}

	params := flags.ConnectParameters()
	params.ApplicationName = "SchemaManager"
	params.DBName = filepath.Base(*path)
	postgresURI, err := params.BuildURI()
	if err != nil {
		log.Panic("Failed to build URI", zap.Error(err))
	}
	myMigrater, err := New(*path, postgresURI, params.DBName)
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
	if *latest == true {
		if err := myMigrater.Up(); err != nil {
			log.Panic(err)
		}
	} else {
		if err := myMigrater.DoMigrate(*desiredVersion, *step); err != nil {
			log.Panic(err)
		}
	}
	postMigrationStep, dirty, err := myMigrater.Version()
	if err != nil {
		log.Fatal("Failed to get Migration Step for confirmation")
	}
	totalMoves := int(postMigrationStep - preMigrationStep)
	if totalMoves == 0 && *latest == false {
		log.Println("No Changes")
	} else {
		log.Printf("Moved %d step(s) in total from Step %d to Step %d", intAbs(totalMoves), preMigrationStep, postMigrationStep)
	}

	currentDBVersion, err := getCurrentDBVersion(postgresURI, params.DBName)
	if err != nil {
		log.Fatal("Failed to get Current DB version for confirmation")
	}
	log.Printf("DB Version: %s, Migration Step # %d, Dirty: %v", currentDBVersion, postMigrationStep, dirty)
}

// DoMigrate performs the migration given the desired state we want to reach
func (m *MyMigrate) DoMigrate(desiredDBVersion semver.Version, desiredStep int) error {
	migrateDirection, err := m.MigrationDirection(desiredDBVersion, desiredStep)
	if err != nil {
		return err
	}
	for migrateDirection != 0 {
		err = m.Steps(int(migrateDirection))
		if err != nil {
			return err
		}
		log.Printf("Migrated %s by %d step", migrateDirection.String(), intAbs(int(migrateDirection)))
		migrateDirection, err = m.MigrationDirection(desiredDBVersion, *step)
		if err != nil {
			return err
		}
	}
	return nil
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
	crdb, err := cockroach.Dial(crdbURI)
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
		log.Printf("Database \"%s\" doesn't exist, attempting to create", database)
		createDB := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database)
		_, err := crdb.Exec(createDB)
		if err != nil {
			return fmt.Errorf("Failed to Create Database: %v", err)
		}
	}
	return nil
}

func getCurrentDBVersion(crdbURI string, database string) (*semver.Version, error) {
	crdb, err := cockroach.Dial(crdbURI)
	defer func() {
		crdb.Close()
	}()
	if err != nil {
		return nil, fmt.Errorf("Failed to dial CRDB while getting DB version: %v", err)
	}

	return crdb.GetVersion(context.Background(), database)
}

// MigrationDirection reads our custom DB version string as well as the Migration Steps from the framework
// and returns a signed integer value of the Direction and count to migrate the db
func (m *MyMigrate) MigrationDirection(desiredVersion semver.Version, desiredStep int) (Direction, error) {
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

	return Direction(desiredVersion.Compare(*currentVersion)), nil
}
