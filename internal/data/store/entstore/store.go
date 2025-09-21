//go:build !entgen

package entstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"example.com/classic/internal/config"
	"example.com/classic/internal/data/ent"
	"example.com/classic/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

// Store 数据存储
type Store struct {
	Client *ent.Client
	config *config.Config
	log    logger.Logger
}

// New 创建数据存储实例
func New(ctx context.Context, cfg *config.Config, log logger.Logger) (*Store, error) {
	log.Info(ctx, "initializing data store", logger.F("driver", cfg.DB.Driver))

	var client *ent.Client
	var err error

	switch cfg.DB.Driver {
	case "sqlite":
		// Ensure _fk=1 is present for foreign keys pragma
		dsn := cfg.DB.DSN
		if !strings.Contains(dsn, "_fk=1") {
			if strings.Contains(dsn, "?") {
				dsn += "&_fk=1"
			} else {
				dsn += "?_fk=1"
			}
		}
		var sqldb *sql.DB
		sqldb, err = sql.Open("sqlite", dsn)
		if err == nil {
			// Best-effort enable foreign_keys in case DSN was ignored
			_, _ = sqldb.Exec("PRAGMA foreign_keys=ON")
			drv := entsql.OpenDB(dialect.SQLite, sqldb)
			client = ent.NewClient(ent.Driver(drv))
		}
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Database, cfg.DB.Charset)
		client, err = ent.Open(dialect.MySQL, dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password, cfg.DB.Database)
		client, err = ent.Open(dialect.Postgres, dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.DB.Driver)
	}

	if err != nil {
		log.Error(ctx, "failed to open database connection", logger.F("error", err))
		return nil, fmt.Errorf("open database: %w", err)
	}

	// 配置数据库连接池
	if cfg.DB.MaxOpen > 0 {
		client = client.Debug()
	}

	store := &Store{
		Client: client,
		config: cfg,
		log:    log,
	}

	// 自动迁移
	if cfg.DB.AutoMigrate {
		if err := store.AutoMigrate(ctx); err != nil {
			log.Error(ctx, "failed to auto migrate database", logger.F("error", err))
			return nil, fmt.Errorf("auto migrate: %w", err)
		}
	}

	log.Info(ctx, "data store initialized successfully", logger.F("driver", cfg.DB.Driver))
	return store, nil
}

// AutoMigrate 自动迁移数据库
func (s *Store) AutoMigrate(ctx context.Context) error {
	s.log.Info(ctx, "starting database migration")

	if err := s.Client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	s.log.Info(ctx, "database migration completed successfully")
	return nil
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	if s.Client != nil {
		return s.Client.Close()
	}
	return nil
}

// Ping 检查数据库连接
func (s *Store) Ping(ctx context.Context) error {
	return s.Client.Schema.Create(ctx)
}
