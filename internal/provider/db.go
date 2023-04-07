package provider

import (
	"errors"
	"fmt"

	"github.com/adystag/jobs-search/internal"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type DB struct{}

func (DB) Provide(module *internal.Module) error {
	db, err := sqlx.Connect("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		module.Configuration.DB.User,
		module.Configuration.DB.Password,
		module.Configuration.DB.Host,
		module.Configuration.DB.Port,
		module.Configuration.DB.Name,
	))
	if err != nil {
		return fmt.Errorf("connecting to mysql db: %w", err)
	}

	if module.Configuration.DB.AutoMigrate {
		instance, err := mysql.WithInstance(db.DB, &mysql.Config{})
		if err != nil {
			return fmt.Errorf("initializing mysql db migrate instance: %w", err)
		}

		migrator, err := migrate.NewWithDatabaseInstance("file://database/migrations", "mysql", instance)
		if err != nil {
			return fmt.Errorf("initializing mysql db migrator: %w", err)
		}

		err = migrator.Up()
		if err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				return fmt.Errorf("migrating up mysql db: %w", err)
			}
		}
	}

	module.DB = db

	return nil
}
