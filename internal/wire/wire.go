//go:build wireinject

package wire

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/google/wire"

	"example.com/classic/internal/config"
	"example.com/classic/internal/data"
	"example.com/classic/internal/data/db"
	"example.com/classic/internal/data/store/sqlstore"
	"example.com/classic/internal/domain"
	"example.com/classic/internal/handler"
	"example.com/classic/internal/infrastructure/messaging"
	"example.com/classic/internal/job/asynq"
	"example.com/classic/internal/repository"
	httpserver "example.com/classic/internal/server/http"
	"example.com/classic/internal/service"
	"example.com/classic/pkg/logger"
)

// InitHTTPServer initializes HTTP Server
func InitHTTPServer(ctx context.Context) (*http.Server, error) {
	wire.Build(
		// Config and logger
		config.Load,
		provideLogger,

		// Data layer (sqlc)
		sqlstore.New,
		provideSQLDB,
		provideDBTX,

		// Task queue
		asynq.New,

		// Transaction manager
		provideTransactionManager,

		// Domain services
		providePasswordHasher,
		provideUserFactory,

		// Infrastructure layer
		provideEventPublisher,

		// Repository layer (sqlc)
		provideUserRepository,

		// Service layer
		service.NewUserService,

		// Handler layer
		handler.NewUserHandler,

		// HTTP server
		httpserver.NewServer,
		provideHTTPServer,
	)
	return nil, nil
}

// provideLogger provides logger instance
func provideLogger(cfg *config.Config) logger.Logger {
	log := logger.New(cfg.Service, cfg.Log.Level, cfg.IsDevelopment())
	logger.SetGlobalLogger(log)
	return log
}

// providePasswordHasher provides password hasher
func providePasswordHasher() domain.PasswordHasher {
	return domain.NewBcryptPasswordHasher()
}

// provideUserFactory provides user factory
func provideUserFactory(hasher domain.PasswordHasher) domain.UserFactory {
	return domain.NewUserFactory(hasher)
}

// provideEventPublisher provides event publisher
func provideEventPublisher(taskQueue *asynq.Queue, log logger.Logger) domain.EventPublisher {
	return messaging.NewAsynqEventPublisher(taskQueue, log)
}

// provideSQLDB provides sql.DB
func provideSQLDB(store *sqlstore.Store) *sql.DB {
	return store.DB
}

// provideDBTX provides DBTX interface for sqlc
func provideDBTX(sqldb *sql.DB) db.DBTX {
	return sqldb
}

// provideTransactionManager provides transaction manager
func provideTransactionManager(sqldb *sql.DB, log logger.Logger) domain.TransactionManager {
	return data.NewTransactionManager(sqldb, log)
}

// provideUserRepository provides user repository using sqlc
func provideUserRepository(dbtx db.DBTX, log logger.Logger) domain.UserRepository {
	return repository.NewUserRepositorySQLC(dbtx, log)
}

// provideHTTPServer provides HTTP server
func provideHTTPServer(server *httpserver.Server) *http.Server {
	return server.GetHTTPServer()
}