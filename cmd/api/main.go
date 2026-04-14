package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/classic/internal/wire"
	"example.com/classic/pkg/logger"
)

func main() {
	ctx := context.Background()

	// Initialize HTTP server
	httpServer, err := wire.InitHTTPServer(ctx)
	if err != nil {
		panic(fmt.Errorf("init http server: %w", err))
	}

	// Initialize gRPC server
	grpcServer, err := wire.InitGRPCServer(ctx)
	if err != nil {
		panic(fmt.Errorf("init grpc server: %w", err))
	}

	// Start HTTP server
	go func() {
		logger.Info(ctx, "HTTP server starting", logger.F("addr", ":8080"))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(ctx, "HTTP server error", logger.F("error", err))
		}
	}()

	// Start gRPC server
	if err := grpcServer.Start(ctx); err != nil {
		logger.Error(ctx, "gRPC server error", logger.F("error", err))
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(ctx, "shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "HTTP server shutdown error", logger.F("error", err))
	}

	// Shutdown gRPC server
	if err := grpcServer.Stop(shutdownCtx); err != nil {
		logger.Error(ctx, "gRPC server shutdown error", logger.F("error", err))
	}

	logger.Info(ctx, "servers exited")
}
