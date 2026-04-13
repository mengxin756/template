package handler

import (
	"strconv"

	"example.com/classic/internal/domain"
	"example.com/classic/internal/handler/request"
	"example.com/classic/internal/service"
	"example.com/classic/pkg/errors"
	"example.com/classic/pkg/logger"
	"example.com/classic/pkg/response"
	"github.com/gin-gonic/gin"
)

// UserHandler HTTP user handler
type UserHandler struct {
	userService service.UserService
	log         logger.Logger
}

// NewUserHandler creates user handler instance
func NewUserHandler(userService service.UserService, log logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         log,
	}
}

// Register user registration
// @Summary User registration
// @Description Create new user account
// @Tags User Management
// @Accept json
// @Produce json
// @Param user body request.CreateUserRequest true "user registration info"
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users [post]
func (h *UserHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()
	h.log.Info(ctx, "user registration request received")

	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn(ctx, "invalid request body", logger.F("error", err))
		response.InvalidParam(c, "invalid request body: "+err.Error())
		return
	}

	user, err := h.userService.Register(ctx, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.log.Info(ctx, "user registration successful", logger.F("user_id", user.ID))
	response.SuccessWithMsg(c, "user registered successfully", user)
}

// GetByID 根据ID获取用户
// @Summary 获取用户信息
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Warn(ctx, "invalid user id", logger.F("id", idStr), logger.F("error", err))
		response.InvalidParam(c, "invalid user id")
		return
	}

	h.log.Debug(ctx, "getting user by id", logger.F("user_id", id))

	user, err := h.userService.GetByID(ctx, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.log.Debug(ctx, "user retrieved successfully", logger.F("user_id", id))
	response.Success(c, user)
}

// Update updates user info
// @Summary Update user info
// @Description Update specified user's info
// @Tags User Management
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body request.UpdateUserRequest true "user update info"
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Warn(ctx, "invalid user id", logger.F("id", idStr), logger.F("error", err))
		response.InvalidParam(c, "invalid user id")
		return
	}

	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn(ctx, "invalid request body", logger.F("error", err))
		response.InvalidParam(c, "invalid request body: "+err.Error())
		return
	}

	h.log.Info(ctx, "updating user", logger.F("user_id", id))

	user, err := h.userService.Update(ctx, id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.log.Info(ctx, "user updated successfully", logger.F("user_id", id))
	response.SuccessWithMsg(c, "user updated successfully", user)
}

// Delete 删除用户
// @Summary 删除用户
// @Description 删除指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Warn(ctx, "invalid user id", logger.F("id", idStr), logger.F("error", err))
		response.InvalidParam(c, "invalid user id")
		return
	}

	h.log.Info(ctx, "deleting user", logger.F("user_id", id))

	if err := h.userService.Delete(ctx, id); err != nil {
		h.handleError(c, err)
		return
	}

	h.log.Info(ctx, "user deleted successfully", logger.F("user_id", id))
	response.SuccessWithMsg(c, "user deleted successfully", nil)
}

// List queries user list
// @Summary Query user list
// @Description Paginated query of user list with filtering
// @Tags User Management
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param name query string false "User name"
// @Param email query string false "User email"
// @Param status query string false "User status"
// @Success 200 {object} response.Response{data=response.PageResponse{data=[]domain.User}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	query := &request.UserQuery{
		Page:     1,
		PageSize: 20,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			query.PageSize = pageSize
		}
	}

	if name := c.Query("name"); name != "" {
		query.Name = &name
	}

	if email := c.Query("email"); email != "" {
		query.Email = &email
	}

	if status := c.Query("status"); status != "" {
		statusEnum := domain.Status(status)
		if statusEnum.IsValid() {
			query.Status = &statusEnum
		}
	}

	h.log.Debug(ctx, "listing users", logger.F("query", query))

	users, total, err := h.userService.List(ctx, query)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.log.Debug(ctx, "users listed successfully", logger.F("total", total), logger.F("count", len(users)))
	response.SuccessWithPage(c, users, total, query.Page, query.PageSize)
}

// ChangeStatus changes user status
// @Summary Change user status
// @Description Change specified user's status
// @Tags User Management
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param status body request.ChangeStatusRequest true "status info"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users/{id}/status [patch]
func (h *UserHandler) ChangeStatus(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Warn(ctx, "invalid user id", logger.F("id", idStr), logger.F("error", err))
		response.InvalidParam(c, "invalid user id")
		return
	}

	var req request.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn(ctx, "invalid request body", logger.F("error", err))
		response.InvalidParam(c, "invalid request body: "+err.Error())
		return
	}

	if !req.Status.IsValid() {
		h.log.Warn(ctx, "invalid status", logger.F("status", req.Status))
		response.InvalidParam(c, "invalid status")
		return
	}

	status := req.Status

	h.log.Info(ctx, "changing user status", logger.F("user_id", id), logger.F("status", status))

	if err := h.userService.ChangeStatus(ctx, id, status); err != nil {
		h.handleError(c, err)
		return
	}

	h.log.Info(ctx, "user status changed successfully", logger.F("user_id", id), logger.F("status", status))
	response.SuccessWithMsg(c, "user status changed successfully", nil)
}

// handleError handles errors uniformly
func (h *UserHandler) handleError(c *gin.Context, err error) {
	ctx := c.Request.Context()

	// Log error
	h.log.Error(ctx, "handler error", logger.F("error", err))

	// Return appropriate response based on error type
	if domainErr, ok := err.(*errors.Error); ok {
		switch domainErr.Code {
		case errors.ErrCodeInvalidParam:
			response.BadRequest(c, domainErr)
		case errors.ErrCodeNotFound:
			response.NotFound(c, domainErr)
		case errors.ErrCodeConflict:
			response.Conflict(c, domainErr)
		case errors.ErrCodeUnauthorized:
			response.Unauthorized(c, domainErr)
		case errors.ErrCodeForbidden:
			response.Forbidden(c, domainErr)
		case errors.ErrCodeTooManyRequest:
			response.TooManyRequests(c, domainErr)
		default:
			response.InternalServerError(c, domainErr)
		}
		return
	}

	// Unknown error type
	response.InternalServerError(c, errors.WrapInternalError(err, "unknown error"))
}
