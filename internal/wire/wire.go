//go:build wireinject

package wire

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/google/wire"

	"example.com/classic/api/grpc/pb"
	"example.com/classic/internal/config"
	"example.com/classic/internal/data"
	"example.com/classic/internal/data/db"
	"example.com/classic/internal/data/store/sqlstore"
	"example.com/classic/internal/domain"
	"example.com/classic/internal/handler"
	"example.com/classic/internal/infrastructure/messaging"
	"example.com/classic/internal/job/asynq"
	"example.com/classic/internal/repository"
	grpcserver "example.com/classic/internal/server/grpc"
	httpserver "example.com/classic/internal/server/http"
	"example.com/classic/internal/service"
	"example.com/classic/pkg/logger"
)

// ============================================================
// Provider Sets
// ============================================================

var ConfigSet = wire.NewSet(
	config.Load,
)

var LoggerSet = wire.NewSet(
	provideLogger,
)

var DataLayerSet = wire.NewSet(
	sqlstore.New,
	provideSQLDB,
	provideDBTX,
)

var TaskQueueSet = wire.NewSet(
	asynq.New,
	provideEventPublisher,
)

var DomainSet = wire.NewSet(
	providePasswordHasher,
	provideUserFactory,
	provideTransactionManager,
)

var RepositorySet = wire.NewSet(
	provideUserRepository,
)

var ServiceSet = wire.NewSet(
	service.NewUserService,
)

var HTTPHandlerSet = wire.NewSet(
	handler.NewUserHandler,
)

var GRPCHandlerSet = wire.NewSet(
	provideUserGRPCHandler,
)

var HTTPServerSet = wire.NewSet(
	httpserver.NewServer,
	provideHTTPServer,
)

var GRPCServerSet = wire.NewSet(
	grpcserver.NewServer,
)

// ============================================================
// Application Initialization Functions
// ============================================================

// InitHTTPServer initializes HTTP Server with cleanup function
// Returns: HTTP server, cleanup function, error
func InitHTTPServer(ctx context.Context) (*http.Server, func(), error) {
	wire.Build(
		ConfigSet,
		LoggerSet,
		DataLayerSet,
		TaskQueueSet,
		DomainSet,
		RepositorySet,
		ServiceSet,
		HTTPHandlerSet,
		HTTPServerSet,
	)
	return nil, nil, nil
}

// InitGRPCServer initializes gRPC Server with cleanup function
// Returns: gRPC server, cleanup function, error
func InitGRPCServer(ctx context.Context) (*grpcserver.Server, func(), error) {
	wire.Build(
		ConfigSet,
		LoggerSet,
		DataLayerSet,
		TaskQueueSet,
		DomainSet,
		RepositorySet,
		ServiceSet,
		GRPCHandlerSet,
		GRPCServerSet,
	)
	return nil, nil, nil
}

// ============================================================
// Provider Functions
// ============================================================

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

// provideTransactionManager provides transaction manager
func provideTransactionManager(sqldb *sql.DB, log logger.Logger) domain.TransactionManager {
	return data.NewTransactionManager(sqldb, log)
}

// provideUserRepository provides user repository using sqlc
func provideUserRepository(dbtx db.DBTX, log logger.Logger) domain.UserRepository {
	return repository.NewUserRepositorySQLC(dbtx, log)
}

// provideUserGRPCHandler provides user gRPC handler
func provideUserGRPCHandler(userSvc service.UserService, log logger.Logger) pb.UserServiceServer {
	return handler.NewUserGRPCHandler(userSvc, log)
}

// provideHTTPServer provides HTTP server
func provideHTTPServer(server *httpserver.Server) *http.Server {
	return server.GetHTTPServer()
}

// provideSQLDB provides sql.DB
func provideSQLDB(store *sqlstore.Store) *sql.DB {
	return store.DB
}

// provideDBTX provides DBTX interface for sqlc
func provideDBTX(sqldb *sql.DB) db.DBTX {
	return sqldb
}

// provideEventPublisher provides event publisher
func provideEventPublisher(taskQueue *asynq.Queue, log logger.Logger) domain.EventPublisher {
	return messaging.NewAsynqEventPublisher(taskQueue, log)
}