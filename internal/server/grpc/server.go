package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"example.com/classic/api/grpc/pb"
	"example.com/classic/internal/config"
	"example.com/classic/pkg/contextx"
	"example.com/classic/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

// Server gRPC server
type Server struct {
	cfg       *config.Config
	log       logger.Logger
	grpcSrv   *grpc.Server
	userSvc   pb.UserServiceServer
}

// NewServer creates a new gRPC server
func NewServer(
	cfg *config.Config,
	log logger.Logger,
	userSvc pb.UserServiceServer,
) *Server {
	return &Server{
		cfg:     cfg,
		log:     log,
		userSvc: userSvc,
	}
}

// Start starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	// Create gRPC server with interceptors
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(s.unaryInterceptor),
		grpc.StreamInterceptor(s.streamInterceptor),
	}
	s.grpcSrv = grpc.NewServer(opts...)

	// Register services
	pb.RegisterUserServiceServer(s.grpcSrv, s.userSvc)

	// Enable reflection for development
	if s.cfg.IsDevelopment() {
		reflection.Register(s.grpcSrv)
	}

	s.log.Info(ctx, "gRPC server starting", logger.F("addr", addr))

	go func() {
		if err := s.grpcSrv.Serve(lis); err != nil {
			s.log.Error(ctx, "gRPC server error", logger.F("error", err))
		}
	}()

	return nil
}

// Stop stops the gRPC server
func (s *Server) Stop(ctx context.Context) error {
	if s.grpcSrv != nil {
		s.log.Info(ctx, "stopping gRPC server")
		s.grpcSrv.GracefulStop()
	}
	return nil
}

// unaryInterceptor 链路追踪拦截器 (增强版)
func (s *Server) unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	// 从 metadata 提取追踪信息
	ctx = s.extractTraceContext(ctx)

	// 设置操作名称
	ctx = contextx.WithOperationName(ctx, info.FullMethod)
	ctx = contextx.WithServiceName(ctx, s.cfg.Service)

	// 创建子 span
	ctx = contextx.ChildSpan(ctx)

	s.log.Debug(ctx, "gRPC request started",
		logger.String("method", info.FullMethod))

	resp, err := handler(ctx, req)

	latency := time.Since(start)
	if err != nil {
		s.log.Error(ctx, "gRPC request failed",
			logger.String("method", info.FullMethod),
			logger.Err(err),
			logger.Duration("latency", latency))
	} else {
		s.log.Info(ctx, "gRPC request completed",
			logger.String("method", info.FullMethod),
			logger.Duration("latency", latency))
	}

	return resp, err
}

// streamInterceptor 链路追踪流式拦截器
func (s *Server) streamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()

	// 从 metadata 提取追踪信息
	ctx := s.extractTraceContext(ss.Context())
	ctx = contextx.WithOperationName(ctx, info.FullMethod)
	ctx = contextx.WithServiceName(ctx, s.cfg.Service)
	ctx = contextx.ChildSpan(ctx)

	// 包装 ServerStream 以传递新的 context
	wrapped := &wrappedServerStream{
		ServerStream: ss,
		ctx:          ctx,
	}

	s.log.Debug(ctx, "gRPC stream started",
		logger.String("method", info.FullMethod))

	err := handler(srv, wrapped)

	latency := time.Since(start)
	if err != nil {
		s.log.Error(ctx, "gRPC stream failed",
			logger.String("method", info.FullMethod),
			logger.Err(err),
			logger.Duration("latency", latency))
	} else {
		s.log.Info(ctx, "gRPC stream completed",
			logger.String("method", info.FullMethod),
			logger.Duration("latency", latency))
	}

	return err
}

// extractTraceContext extracts trace context from gRPC metadata
func (s *Server) extractTraceContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return contextx.WithTraceID(ctx, contextx.GenerateTraceID())
	}

	if values := md.Get("x-trace-id"); len(values) > 0 {
		ctx = contextx.WithTraceID(ctx, values[0])
	} else {
		ctx = contextx.WithTraceID(ctx, contextx.GenerateTraceID())
	}

	if values := md.Get("x-span-id"); len(values) > 0 {
		ctx = contextx.WithSpanID(ctx, values[0])
	}
	if values := md.Get("x-parent-span-id"); len(values) > 0 {
		ctx = contextx.WithParentSpanID(ctx, values[0])
	}
	if values := md.Get("x-user-id"); len(values) > 0 {
		ctx = contextx.WithUserID(ctx, values[0])
	}
	if values := md.Get("x-client-ip"); len(values) > 0 {
		ctx = contextx.WithClientIP(ctx, values[0])
	}

	return ctx
}

// wrappedServerStream wraps ServerStream to pass context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
