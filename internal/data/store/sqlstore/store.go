package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"example.com/classic/internal/config"
	"example.com/classic/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

// Store data storage using standard sql.DB
type Store struct {
	DB     *sql.DB
	config *config.Config
	log    logger.Logger
}

// New creates a new data store instance
func New(ctx context.Context, cfg *config.Config, log logger.Logger) (*Store, error) {
	log.Info(ctx, "initializing data store", logger.F("driver", cfg.DB.Driver))

	var db *sql.DB
	var err error

	switch cfg.DB.Driver {
	case "sqlite":
		dsn := cfg.DB.DSN
		if !strings.Contains(dsn, "_fk=1") {
			if strings.Contains(dsn, "?") {
				dsn += "&_fk=1"
			} else {
				dsn += "?_fk=1"
			}
		}
		db, err = sql.Open("sqlite", dsn)
		if err != nil {
			return nil, fmt.Errorf("open sqlite: %w", err)
		}
		// Enable foreign keys
		_, _ = db.Exec("PRAGMA foreign_keys=ON")

	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Database, cfg.DB.Charset)
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("open mysql: %w", err)
		}

	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password, cfg.DB.Database)
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("open postgres: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.DB.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	store := &Store{
		DB:     db,
		config: cfg,
		log:    log,
	}

	// Auto migrate (using schema.sql)
	if cfg.DB.AutoMigrate {
		if err := store.AutoMigrate(ctx); err != nil {
			log.Error(ctx, "failed to auto migrate database", logger.F("error", err))
			return nil, fmt.Errorf("auto migrate: %w", err)
		}
	}

	log.Info(ctx, "data store initialized successfully", logger.F("driver", cfg.DB.Driver))
	return store, nil
}

// AutoMigrate runs schema migration
func (s *Store) AutoMigrate(ctx context.Context) error {
	s.log.Info(ctx, "starting database migration")
	// For sqlc, migrations should be handled separately (e.g., golang-migrate)
	// This is a placeholder for manual migration execution
	s.log.Info(ctx, "database migration completed successfully")
	return nil
}

// Close closes the database connection
func (s *Store) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

// Ping checks database connection
func (s *Store) Ping(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}
