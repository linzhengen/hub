package develop

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"

	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/linzhengen/hub/v1/server/db/migrations"

	"github.com/linzhengen/hub/v1/server/pkg/logger"
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
	logger.Info("start postgres migrate up")
	mi, err := migrateInstance(m.db)
	if err != nil {
		logger.Severef("failed create migrate instance")
		return err
	}
	if err := mi.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Severef("failed postgres migrate up, err: %s", err)
		return err
	}
	logger.Info("postgres migrate up successfully")
	return nil
}

func (m migrateUseCase) Down(ctx context.Context) error {
	logger.Info("start postgres migrate down")
	mi, err := migrateInstance(m.db)
	if err != nil {
		logger.Severef("failed create migrate instance")
		return err
	}
	if err := mi.Down(); err != nil {
		logger.Severef("failed postgres migrate down, err: %s", err)
		return err
	}
	logger.Info("postgres migrate down successfully")
	return nil
}

func migrateInstance(db *sql.DB) (*migrate.Migrate, error) {
	d, err := iofs.New(migrations.PostgresMigrationsFs, "postgres")
	if err != nil {
		logger.Severef("failed get postgres migrations", err)
		return nil, err
	}
	driver, err := migratePostgres.WithInstance(db, &migratePostgres.Config{})
	if err != nil {
		logger.Severef("failed create postgres instance driver, err: %s", err)
		return nil, err
	}
	m, err := migrate.NewWithInstance(
		"iofs",
		d,
		"postgres",
		driver,
	)
	if err != nil {
		logger.Severef("failed create postgres migration instance, err: %s", err)
		return nil, err
	}
	return m, err
}
