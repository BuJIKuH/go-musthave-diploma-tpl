package postgres

import (
	"database/sql"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func RunMigrations(dns string, logger *zap.Logger) error {
	log.Println("Running migrations...")
	db, err := sql.Open("postgres", dns)
	if err != nil {
		logger.Error("cannot open DB", zap.Error(err))
		return err
	}
	defer db.Close()

	path, err := filepath.Abs("internal/migrations")
	if err != nil {
		logger.Error("failed to get absolute path", zap.Error(err))
		return err
	}

	migrationsURL := "file://" + path
	logger.Info("Running migrations", zap.String("path", migrationsURL))

	m, err := migrate.New(migrationsURL, dns)
	if err != nil {
		logger.Error("cannot create migration", zap.Error(err))
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Info("cannot run migration", zap.Error(err))
		return err
	}
	logger.Info("migrations successfully migrated")

	return nil
}
