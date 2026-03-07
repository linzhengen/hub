package develop

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"

	migrateMysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/linzhengen/hub/server/db/migrations"

	"github.com/linzhengen/hub/server/pkg/logger"
)

type MigrateUseCase interface {
	Up(ctx context.Context) error
	Down(ctx context.Context) error
}

func NewMigrateUseCase(
	db *sql.DB,
) MigrateUseCase {
	return &migrateUseCase{
		db: db,
	}
}

type migrateUseCase struct {
	db *sql.DB
}

func (m migrateUseCase) Up(ctx context.Context) error {
	logger.Info("start mysql migrate up")
	mi, err := migrateInstance(m.db)
	if err != nil {
		logger.Severef("failed create migrate instance")
		return err
	}
	if err := mi.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Severef("failed mysql migrate up, err: %s", err)
		return err
	}
	logger.Info("mysql migrate up successfully")
	return nil
}

func (m migrateUseCase) Down(ctx context.Context) error {
	logger.Info("start mysql migrate down")
	mi, err := migrateInstance(m.db)
	if err != nil {
		logger.Severef("failed create migrate instance")
		return err
	}
	if err := mi.Down(); err != nil {
		logger.Severef("failed mysql migrate down, err: %s", err)
		return err
	}
	logger.Info("mysql migrate down successfully")
	return nil
}

func migrateInstance(db *sql.DB) (*migrate.Migrate, error) {
	d, err := iofs.New(migrations.MySqlMigrationsFs, "mysql")
	if err != nil {
		logger.Severef("failed get mysql migrations", err)
		return nil, err
	}
	driver, err := migrateMysql.WithInstance(db, &migrateMysql.Config{})
	if err != nil {
		logger.Severef("failed create mysql instance driver, err: %s", err)
		return nil, err
	}
	m, err := migrate.NewWithInstance(
		"iofs",
		d,
		"mysql",
		driver,
	)
	if err != nil {
		logger.Severef("failed create mysql migration instance, err: %s", err)
		return nil, err
	}
	return m, err
}
