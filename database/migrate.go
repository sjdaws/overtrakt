package database

import (
	"fmt"
	"path/filepath"
	"runtime"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const targetVersion = 2

// Create a new database
func (d *Database) create(dbname string) error {
	_, err := d.connection.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil {
		return err
	}

	_, err = d.connection.Exec(fmt.Sprintf("USE %s", dbname))
	if err != nil {
		return err
	}

	return nil
}

// Run migrations
func (d *Database) migrate(dbname string) error {
	_, baseDir, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(baseDir)

	driver, err := mysql.WithInstance(d.connection, &mysql.Config{NoLock: true})
	if err != nil {
		return err
	}

	migrations, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s/../migrate", basePath),
		dbname,
		driver,
	)
	if err != nil {
		return err
	}

	version, _, _ := migrations.Version()
	if version == targetVersion {
		return nil
	}

	return migrations.Migrate(targetVersion)
}
