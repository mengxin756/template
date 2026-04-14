package handler

import (
	"context"

	"example.com/classic/api/grpc/pb"
	"example.com/classic/internal/domain"
	"example.com/classic/internal/handler/request"
	"example.com/classic/internal/service"
	"example.com/classic/pkg/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserGRPCHandler implements pb.UserServiceServer
type UserGRPCHandler struct {
	pb.UnimplementedUserServiceServer
	userSvc service.UserService
	log     logger.Logger
}

// NewUserGRPCHandler creates a new gRPC user handler
func NewUserGRPCHandler(userSvc service.UserService, log logger.Logger) pb.UserServiceServer {
	return &UserGRPCHandler{
		userSvc: userSvc,
		log:     log,
	}
}

// Register creates a new user
func (h *UserGRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.UserResponse, error) {
	h.log.Debug(ctx, "gRPC register request", logger.F("email", req.Email))

	user, err := h.userSvc.Register(ctx, &request.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return h.toUserResponse(user), nil
}

// GetByID retrieves a user by ID
func (h *UserGRPCHandler) GetByID(ctx context.Context, req *pb.GetByIDRequest) (*pb.UserResponse, error) {
	h.log.Debug(ctx, "gRPC get by id request", logger.F("id", req.Id))

	user, err := h.userSvc.GetByID(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	return h.toUserResponse(user), nil
}

// Update updates a user
func (h *UserGRPCHandler) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UserResponse, error) {
	h.log.Debug(ctx, "gRPC update request", logger.F("id", req.Id))

	updateReq := &request.UpdateUserRequest{}
	if req.Name != nil {
		updateReq.Name = req.Name
	}
	if req.Email != nil {
		updateReq.Email = req.Email
	}
	if req.Status != nil {
		status := domain.Status(req.Status.String())
		updateReq.Status = &status
	}

	user, err := h.userSvc.Update(ctx, int(req.Id), updateReq)
	if err != nil {
		return nil, err
	}

	return h.toUserResponse(user), nil
}

// Delete deletes a user
func (h *UserGRPCHandler) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	h.log.Debug(ctx, "gRPC delete request", logger.F("id", req.Id))

	if err := h.userSvc.Delete(ctx, int(req.Id)); err != nil {
		return nil, err
	}

	return &pb.DeleteResponse{Success: true}, nil
}

// List lists users
func (h *UserGRPCHandler) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	h.log.Debug(ctx, "gRPC list request", logger.F("page", req.Page))

	query := &request.UserQuery{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}
	if req.Id != nil {
		id := int(*req.Id)
		query.ID = &id
	}
	if req.Name != nil {
		query.Name = req.Name
	}
	if req.Email != nil {
		query.Email = req.Email
	}
	if req.Status != nil {
		status := domain.Status(req.Status.String())
		query.Status = &status
	}

	users, total, err := h.userSvc.List(ctx, query)
	if err != nil {
		return nil, err
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = h.toUser(user)
	}

	return &pb.ListResponse{
		Users: pbUsers,
		Total: total,
	}, nil
}

// ChangeStatus changes user status
func (h *UserGRPCHandler) ChangeStatus(ctx context.Context, req *pb.ChangeStatusRequest) (*pb.UserResponse, error) {
	h.log.Debug(ctx, "gRPC change status request",
		logger.F("id", req.Id),
		logger.F("status", req.Status))

	status := domain.Status(req.Status.String())
	if err := h.userSvc.ChangeStatus(ctx, int(req.Id), status); err != nil {
		return nil, err
	}

	user, err := h.userSvc.GetByID(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	return h.toUserResponse(user), nil
}

// toPBStatus converts domain.Status to pb.Status
func toPBStatus(s domain.Status) pb.Status {
	switch s {
	case domain.StatusActive:
		return pb.Status_ACTIVE
	case domain.StatusInactive:
		return pb.Status_INACTIVE
	default:
		return pb.Status_UNSPECIFIED
	}
}

// toUserResponse converts domain user to protobuf response
func (h *UserGRPCHandler) toUserResponse(user *domain.User) *pb.UserResponse {
	return &pb.UserResponse{
		Id:        int32(user.ID()),
		Name:      user.Name().String(),
		Email:     user.Email().String(),
		Status:    toPBStatus(user.Status()),
		CreatedAt: timestamppb.New(user.CreatedAt()),
		UpdatedAt: timestamppb.New(user.UpdatedAt()),
	}
}

// toUser converts domain user to protobuf user
func (h *UserGRPCHandler) toUser(user *domain.User) *pb.User {
	return &pb.User{
		Id:        int32(user.ID()),
		Name:      user.Name().String(),
		Email:     user.Email().String(),
		Status:    toPBStatus(user.Status()),
		CreatedAt: timestamppb.New(user.CreatedAt()),
		UpdatedAt: timestamppb.New(user.UpdatedAt()),
	}
}
