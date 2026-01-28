//go:build wireinject

package wire

import (
	"context"
	"net/http"

	"github.com/google/wire"

	"example.com/classic/internal/config"
	"example.com/classic/internal/data/ent"
	"example.com/classic/internal/data/redis"
	"example.com/classic/internal/data/store/entstore"
	"example.com/classic/internal/handler"
	"example.com/classic/internal/job/asynq"
	"example.com/classic/internal/repository"
	"example.com/classic/internal/server/http"
	httpserver "example.com/classic/internal/server/http"
	"example.com/classic/internal/service"
	"example.com/classic/pkg/logger"
)

// InitHTTPServer 返回完整的 HTTP Server 和配置
func InitHTTPServer(ctx context.Context) (*http.Server, *config.Config, logger.Logger, error) {
	wire.Build(
		// 配置和日志
		config.Load,
		provideLogger,

		// 数据层
		entstore.New,
		provideEntClient,
		redis.New,
		asynq.New,

		// 仓储层
		repository.NewUserRepository,

		// 服务层
		service.NewUserService,

		// 处理器层
		handler.NewUserHandler,

		// HTTP 服务器
		http.NewServer,
		provideHTTPServer,
	)
	return nil, nil, nil, nil
}

// provideLogger 提供日志实例
func provideLogger(cfg *config.Config) logger.Logger {
	log := logger.New(cfg.Service, cfg.Log.Level, cfg.IsDevelopment())
	logger.SetGlobalLogger(log)
	return log
}

// provideEntClient 提供 Ent 客户端
func provideEntClient(store *entstore.Store) *ent.Client {
	return store.Client
}

// provideHTTPServer 提供 HTTP 服务器
func provideHTTPServer(server *httpserver.Server) *http.Server {
	return server.GetHTTPServer()
}
