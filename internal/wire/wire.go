//go:build wireinject

package wire

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	appuser "example.com/classic/internal/app/user"
	"example.com/classic/internal/config"
	"example.com/classic/internal/data/ent"
	"example.com/classic/internal/data/store/entstore"
	"example.com/classic/internal/logger"
	httpserver "example.com/classic/internal/server/http"
	"example.com/classic/internal/server/http/router"
)

// InitHTTPServer 返回完整的 HTTP Server 和配置
func InitHTTPServer(ctx context.Context) (*http.Server, *config.Config, *logger.Logger, error) {
	wire.Build(
		config.Load,
		logger.New,
		entstore.New,
		provideEntClient,
		appuser.NewRepository,
		appuser.NewService,
		router.NewUserHandler,
		httpserver.BuildEngine,
		provideHTTPServer,
	)
	return nil, nil, nil, nil
}

func provideEntClient(store *entstore.Store) *ent.Client { return store.Client }

func provideHTTPServer(cfg *config.Config, engine *gin.Engine) *http.Server {
	return &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           engine,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}
