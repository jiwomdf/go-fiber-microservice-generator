package helper

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/spf13/viper"
)

func RunMigrations(dsn string) {
	if err := applyMigrations(dsn); err != nil {
		slog.Warn("Migration failed; continuing without applying migrations", "error", err)
	}
}

func applyMigrations(dsn string) error {
	sourceURL := viper.GetString(POSTGRES_PATH_MIGRATE)
	if sourceURL == "" {
		slog.Warn("No migration path configured; skipping migration")
		return nil
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open DB for migration: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{
		DatabaseName:          viper.GetString(POSTGES_DATABASE),
		SchemaName:            viper.GetString(POSTGRES_SCHEMA),
		MultiStatementEnabled: true,
	})
	if err != nil {
		return fmt.Errorf("create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, viper.GetString(POSTGES_DATABASE), driver)
	if err != nil {
		return fmt.Errorf("init migrate instance: %w", err)
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			slog.Warn("Failed to close migration source", "error", sourceErr)
		}
		if dbErr != nil {
			slog.Warn("Failed to close migration database", "error", dbErr)
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("apply migrations: %w", err)
	}

	slog.Info("Database migrations applied successfully")
	return nil
}
