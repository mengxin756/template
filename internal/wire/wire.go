//go:build wireinject

package wire

import (
	"context"
	"net/http"

	"github.com/google/wire"

	"example.com/classic/internal/config"
	"example.com/classic/internal/data/ent"
	"example.com/classic/internal/data/store/entstore"
	"example.com/classic/internal/domain"
	"example.com/classic/internal/handler"
	"example.com/classic/internal/infrastructure/messaging"
	"example.com/classic/internal/job/asynq"
	"example.com/classic/internal/repository"
	httpserver "example.com/classic/internal/server/http"
	"example.com/classic/internal/service"
	"example.com/classic/pkg/logger"
)

// InitHTTPServer 返回完整的 HTTP Server
func InitHTTPServer(ctx context.Context) (*http.Server, error) {
	wire.Build(
		// 配置和日志
		config.Load,
		provideLogger,

		// 数据层
		entstore.New,
		provideEntClient,
		asynq.New,

		// 领域服务
		providePasswordHasher,
		provideUserFactory,

		// 基础设施层
		provideEventPublisher,

		// 仓储层
		repository.NewUserRepository,

		// 服务层
		service.NewUserService,

		// 处理器层
		handler.NewUserHandler,

		// HTTP 服务器
		httpserver.NewServer,
		provideHTTPServer,
	)
	return nil, nil
}

// provideLogger 提供日志实例
func provideLogger(cfg *config.Config) logger.Logger {
	log := logger.New(cfg.Service, cfg.Log.Level, cfg.IsDevelopment())
	logger.SetGlobalLogger(log)
	return log
}

// providePasswordHasher 提供密码哈希器
func providePasswordHasher() domain.PasswordHasher {
	return domain.NewBcryptPasswordHasher()
}

// provideUserFactory 提供用户工厂
func provideUserFactory(hasher domain.PasswordHasher) domain.UserFactory {
	return domain.NewUserFactory(hasher)
}

// provideEventPublisher 提供事件发布器
func provideEventPublisher(taskQueue *asynq.Queue, log logger.Logger) domain.EventPublisher {
	return messaging.NewAsynqEventPublisher(taskQueue, log)
}

// provideEntClient 提供 Ent 客户端
func provideEntClient(store *entstore.Store) *ent.Client {
	return store.Client
}

// provideHTTPServer 提供 HTTP 服务器
func provideHTTPServer(server *httpserver.Server) *http.Server {
	return server.GetHTTPServer()
}