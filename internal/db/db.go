package db

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type DB struct {
	Conn *sqlx.DB
}

func NewDB(postgresURL string, logger *zap.Logger) (*DB, error) {
	conn, err := sqlx.Connect("postgres", postgresURL)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", zap.Error(err), zap.String("url", postgresURL))
		return nil, err
	}

	
	if err := conn.Ping(); err != nil {
		logger.Error("Failed to ping PostgreSQL", zap.Error(err))
		conn.Close()
		return nil, err
	}

	migrationDriver, err := postgres.WithInstance(conn.DB, &postgres.Config{})
	if err != nil {
		logger.Error("Failed to initialize migration driver", zap.Error(err))
		conn.Close()
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations",
		"postgres",
		migrationDriver,
	)
	if err != nil {
		logger.Error("Failed to initialize migrations", zap.String("path", "file://internal/db/migrations"), zap.Error(err))
		conn.Close()
		return nil, err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("Failed to apply migrations", zap.Error(err))
		conn.Close()
		return nil, err
	}

	logger.Info("Database initialized successfully")
	return &DB{Conn: conn}, nil
}
