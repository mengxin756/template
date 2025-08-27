//go:build !entgen

package entstore

import (
    "context"
    "fmt"

    "entgo.io/ent/dialect/sql"
    _ "modernc.org/sqlite"

    "example.com/classic/internal/config"
    "example.com/classic/internal/logger"
    "example.com/classic/internal/data/ent"
)

type Store struct {
    Client *ent.Client
}

func New(ctx context.Context, cfg *config.Config, log *logger.Logger) (*Store, error) {
    drv, err := sql.Open(cfg.DB.Driver, cfg.DB.DSN)
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }
    client := ent.NewClient(ent.Driver(drv))
    if cfg.DB.AutoMigrate {
        if err := client.Schema.Create(ctx); err != nil {
            return nil, fmt.Errorf("ent migrate: %w", err)
        }
    }
    return &Store{Client: client}, nil
}


