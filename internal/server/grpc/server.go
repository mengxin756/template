package grpc

import (
	"context"
	"fmt"
	"net"

	"example.com/classic/api/grpc/pb"
	"example.com/classic/internal/config"
	"example.com/classic/pkg/logger"
	"google.golang.org/grpc"
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

// unaryInterceptor logs requests
func (s *Server) unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	s.log.Debug(ctx, "gRPC request",
		logger.F("method", info.FullMethod),
		logger.F("request", req))

	resp, err := handler(ctx, req)
	if err != nil {
		s.log.Warn(ctx, "gRPC error",
			logger.F("method", info.FullMethod),
			logger.F("error", err))
	}
	return resp, err
}

// streamInterceptor logs stream requests
func (s *Server) streamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	s.log.Debug(ss.Context(), "gRPC stream request",
		logger.F("method", info.FullMethod))

	return handler(srv, ss)
}
