package postgres

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/merdernoty/job-hunter/config"
	"github.com/merdernoty/job-hunter/pkg/logger"
)

const (
	migrationsPath = "file://migrations"
)

type Migrator interface {
	Up() error
	Down() error
	Force(version int) error
	Drop() error
	Version() (version uint, dirty bool, err error)
	Close() error
	Steps(n int) error
}

type migrator struct {
	migrate *migrate.Migrate
	logger  logger.Logger
}

func NewMigrator(db *sqlx.DB, cfg *config.Config, logger logger.Logger) (Migrator, error) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}
	if err := createMigrationsDir(); err != nil {
		logger.Warnf("Failed to create migrations directory: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &migrator{
		migrate: m,
		logger:  logger,
	}, nil
}

func (m *migrator) Up() error {
	m.logger.Info("Starting database migrations...")

	if err := m.migrate.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	m.logger.Info("Database migrations completed successfully")
	return nil
}

func (m *migrator) Down() error {
	m.logger.Info("Rolling back all migrations...")

	if err := m.migrate.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	m.logger.Info("All migrations rolled back successfully")
	return nil
}

func (m *migrator) Steps(n int) error {
	direction := "forward"
	if n < 0 {
		direction = "backward"
	}

	m.logger.Infof("Running %d migration steps %s...", abs(n), direction)

	if err := m.migrate.Steps(n); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No migration steps to apply")
			return nil
		}
		return fmt.Errorf("failed to run migration steps: %w", err)
	}

	m.logger.Infof("Successfully completed %d migration steps %s", abs(n), direction)
	return nil
}

func (m *migrator) Force(version int) error {
	m.logger.Infof("Forcing migration version to %d...", version)

	if err := m.migrate.Force(version); err != nil {
		return fmt.Errorf("failed to force migration version: %w", err)
	}

	m.logger.Infof("Successfully forced migration version to %d", version)
	return nil
}

func (m *migrator) Drop() error {
	m.logger.Warn("Dropping all database objects...")

	if err := m.migrate.Drop(); err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	m.logger.Info("All database objects dropped successfully")
	return nil
}

func (m *migrator) Version() (version uint, dirty bool, err error) {
	version, dirty, err = m.migrate.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			m.logger.Info("No migrations have been applied yet")
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	status := "clean"
	if dirty {
		status = "dirty"
	}

	m.logger.Infof("Current migration version: %d (%s)", version, status)
	return version, dirty, nil
}

func (m *migrator) Close() error {
	if sourceErr, dbErr := m.migrate.Close(); sourceErr != nil || dbErr != nil {
		return fmt.Errorf("failed to close migrator: source error: %v, db error: %v", sourceErr, dbErr)
	}
	return nil
}

func AutoMigrate(db *sqlx.DB, cfg *config.Config, logger logger.Logger) error {
	migrationDB, err := NewPsqlDB(cfg, logger)
	if err != nil {
		return err
	}

	migrator, err := NewMigrator(migrationDB, cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	version, dirty, err := migrator.Version()
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state at version %d, please fix manually", version)
	}

	if err := migrator.Up(); err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}

	return nil
}

func createMigrationsDir() error {
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		return os.MkdirAll("migrations", 0755)
	}
	return nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
