package handler

import (
	"strconv"

	"example.com/classic/internal/domain"
	"example.com/classic/pkg/errors"
	"example.com/classic/pkg/logger"
	"example.com/classic/pkg/response"
	"github.com/gin-gonic/gin"
)

// UserHandler HTTP 用户处理器
type UserHandler struct {
	userService domain.UserService
	log         logger.Logger
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService domain.UserService, log logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         log,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body domain.CreateUserRequest true "用户注册信息"
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users [post]
func (h *UserHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()
	h.log.Info(ctx, "user registration request received")

	var req domain.CreateUserRequest
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

// Update 更新用户信息
// @Summary 更新用户信息
// @Description 更新指定用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param user body domain.UpdateUserRequest true "用户更新信息"
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

	var req domain.UpdateUserRequest
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

// List 查询用户列表
// @Summary 查询用户列表
// @Description 分页查询用户列表，支持条件筛选
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param name query string false "用户姓名"
// @Param email query string false "用户邮箱"
// @Param status query string false "用户状态"
// @Success 200 {object} response.Response{data=response.PageResponse{data=[]domain.User}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	// 解析查询参数
	query := &domain.UserQuery{
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

// ChangeStatus 改变用户状态
// @Summary 改变用户状态
// @Description 改变指定用户的状态
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param status body map[string]string true "状态信息"
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

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn(ctx, "invalid request body", logger.F("error", err))
		response.InvalidParam(c, "invalid request body: "+err.Error())
		return
	}

	status := domain.Status(req.Status)
	if !status.IsValid() {
		h.log.Warn(ctx, "invalid status", logger.F("status", req.Status))
		response.InvalidParam(c, "invalid status")
		return
	}

	h.log.Info(ctx, "changing user status", logger.F("user_id", id), logger.F("status", status))

	if err := h.userService.ChangeStatus(ctx, id, status); err != nil {
		h.handleError(c, err)
		return
	}

	h.log.Info(ctx, "user status changed successfully", logger.F("user_id", id), logger.F("status", status))
	response.SuccessWithMsg(c, "user status changed successfully", nil)
}

// handleError 统一错误处理
func (h *UserHandler) handleError(c *gin.Context, err error) {
	ctx := c.Request.Context()

	// 记录错误日志
	h.log.Error(ctx, "handler error", logger.F("error", err))

	// 根据错误类型返回相应的响应
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

	// 未知错误类型
	response.InternalServerError(c, errors.WrapInternalError(err, "unknown error"))
}
