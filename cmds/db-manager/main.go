// Bootstrap script for Database deployment and migration

package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
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

var (
	path      = flag.String("migration_files", "", "path to db migration files directory")
	dbVersion = flag.String("db_version", "", "the db version to migrate to (ex: v1.0.0)")
	step      = flag.Int("migration_step", 0, "the db migration step to go to")

	cockroachParams = struct {
		host            *string
		port            *int
		sslMode         *string
		sslDir          *string
		user            *string
		applicationName *string
	}{
		host:    flag.String("cockroach_host", "0.0.0.0", "cockroach host to connect to"),
		port:    flag.Int("cockroach_port", 26257, "cockroach port to connect to"),
		sslMode: flag.String("cockroach_ssl_mode", "disable", "cockroach sslmode"),
		user:    flag.String("cockroach_user", "root", "cockroach user to authenticate as"),
		sslDir:  flag.String("cockroach_ssl_dir", "", "directory to ssl certificates. Must contain files: ca.crt, client.<user>.crt, client.<user>.key"),
	}
)

func main() {
	flag.Parse()
	if *path == "" {
		log.Panic("Must specify migration_files path")
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
	}
	postgresURI, err := dssCockroach.BuildURI(uriParams)
	if err != nil {
		log.Panic("Failed to build URI", zap.Error(err))
	}
	crdbURI := strings.Replace(postgresURI, "postgresql", "cockroachdb", 1)
	*path = fmt.Sprintf("file://%v", *path)
	migrater, migraterErr := migrate.New(*path, crdbURI)
	myMigrater := &MyMigrate{migrater}
	defer func() {
		if migraterErr == nil {
			if _, err := myMigrater.Close(); err != nil {
				log.Println(err)
			}
		}
	}()
	if migraterErr == nil {
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
	}
	totalMove := 0
	direction := myMigrater.MigrationDirection(postgresURI, *dbVersion, *step)
	for direction != 0 {
		executeErr := myMigrater.Steps(direction)
		if executeErr != nil {
			log.Fatal(executeErr)
		}
		totalMove += direction
		log.Printf("Migrated %s by %d step", directionText(direction), int(math.Abs(float64(direction))))
		direction = myMigrater.MigrationDirection(postgresURI, *dbVersion, *step)
	}
	if totalMove == 0 {
		log.Println("No Changes")
	} else {
		log.Printf("Moved %d steps %s in total", int(math.Abs(float64(totalMove))), directionText(totalMove))
	}

	currentDBVersion, err := getCurrentDBVersion(postgresURI)
	if err != nil {
		log.Fatal("Failed to get Current DB version for confirmation")
	}
	migrationStep, dirty, err := myMigrater.Version()
	if err != nil {
		log.Fatal("Failed to get Migration Step for confirmation")
	}
	log.Printf("DB Version: %s, Migration Step # %d, Dirty: %v", currentDBVersion, migrationStep, dirty)
}

func directionText(direction int) string {
	if direction > 0 {
		return "Up"
	} else if direction < 0 {
		return "Down"
	}
	return "No Change"
}

func getCurrentDBVersion(crdbURI string) (string, error) {
	crdb, err := dssCockroach.Dial(crdbURI)
	if err != nil {
		log.Panic("Failed to dial CockroachDB")
	}
	// check if schema_versions table exists
	const checkTableQuery = `
		SELECT EXISTS (
  		SELECT *
		  FROM information_schema.tables 
   		WHERE table_name = 'schema_versions'
  )`
	row := crdb.QueryRow(checkTableQuery)
	var ret bool
	scanErr := row.Scan(&ret)
	if scanErr != nil {
		return "", scanErr
	}
	version := "v0.0.0"
	if ret {
		const getVersionQuery = `
      SELECT schema_version 
      FROM schema_versions
      WHERE onerow_enforcer = TRUE`
		row := crdb.QueryRow(getVersionQuery)
		scanErr = row.Scan(&version)
		if scanErr != nil {
			return "", scanErr
		}
		if version == "" {
			version = "v0.0.0"
		}
	}
	crdb.Close()
	return version, nil
}

func parseVersion(version string) [3]int {
	splitVersion := strings.Split(version, ".")
	first, err := strconv.Atoi(strings.Replace(splitVersion[0], "v", "", 1))
	if err != nil {
		log.Panic(err)
	}
	second, err := strconv.Atoi(strings.Replace(splitVersion[1], "v", "", 1))
	if err != nil {
		log.Panic(err)
	}
	third, err := strconv.Atoi(strings.Replace(splitVersion[2], "v", "", 1))
	if err != nil {
		log.Panic(err)
	}
	result := [3]int{first, second, third}
	return result
}

// MigrationDirection reads our custom DB version string as well as the Migration Steps from the framework
// and returns a signed integer value of the direction and count to migrate the db
func (m *MyMigrate) MigrationDirection(dbURI string, desiredVersion string, desireStep int) int {
	if desireStep != 0 {
		migrateStep, dirty, err := m.Version()
		if err != nil {
			log.Panic(err)
		}
		if dirty {
			log.Fatal("DB in Dirty state, Please fix before migrating")
		}
		return desireStep - int(migrateStep)
	}
	currentVersion, err := getCurrentDBVersion(dbURI)
	if err != nil {
		log.Panic("Failed to get current version", zap.Error(err))
	}
	var (
		current     = parseVersion(currentVersion)
		destination = parseVersion(desiredVersion)
	)
	for i := 0; i < 3; i++ {
		diff := destination[0] - current[0]
		if diff > 0 {
			return 1
		} else if diff < 0 {
			return -1
		}
	}
	return 0
}
