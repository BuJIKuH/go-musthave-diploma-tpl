package postgres

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

var openDB = sql.Open

type DBStorage struct {
	DB     *sql.DB
	Logger *zap.Logger
}

func NewDBStorage(dns string, logger *zap.Logger) (*DBStorage, error) {

	db, err := openDB("postgres", dns)
	if err != nil {
		logger.Error("failed to open database", zap.Error(err))
	}

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", zap.Error(err))
		return nil, err
	}

	logger.Info("successfully connected to database", zap.String("dns", dns))

	return &DBStorage{
		DB:     db,
		Logger: logger,
	}, nil
}

func (s *DBStorage) PingContext(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}

func (s *DBStorage) Close() error {
	return s.DB.Close()
}
