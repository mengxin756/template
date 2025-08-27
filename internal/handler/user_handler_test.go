package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/classic/internal/domain"
	"example.com/classic/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService 模拟用户服务
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, id int) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) Update(ctx context.Context, id int, req *domain.UpdateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) List(ctx context.Context, query *domain.UserQuery) ([]*domain.User, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserService) ChangeStatus(ctx context.Context, id int, status domain.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func TestUserHandler_Register(t *testing.T) {
	// 设置 Gin 测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockUserService)
	log := logger.New("test", "debug", true)

	// 创建处理器
	handler := NewUserHandler(mockService, log)

	// 创建测试请求
	reqBody := domain.CreateUserRequest{
		Name:     "测试用户",
		Email:    "test@example.com",
		Password: "password123",
	}
	reqBytes, _ := json.Marshal(reqBody)

	// 创建 HTTP 请求
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 设置模拟行为
	mockUser := &domain.User{
		ID:     1,
		Name:   "测试用户",
		Email:  "test@example.com",
		Status: domain.StatusActive,
	}
	mockService.On("Register", mock.Anything, &reqBody).Return(mockUser, nil)

	// 执行处理器
	handler.Register(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["code"])
	assert.Equal(t, "user registered successfully", response["msg"])

	// 验证模拟调用
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetByID(t *testing.T) {
	// 设置 Gin 测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockUserService)
	log := logger.New("test", "debug", true)

	// 创建处理器
	handler := NewUserHandler(mockService, log)

	// 创建 HTTP 请求
	req := httptest.NewRequest("GET", "/api/v1/users/1", nil)

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	// 设置模拟行为
	mockUser := &domain.User{
		ID:     1,
		Name:   "测试用户",
		Email:  "test@example.com",
		Status: domain.StatusActive,
	}
	mockService.On("GetByID", mock.Anything, 1).Return(mockUser, nil)

	// 执行处理器
	handler.GetByID(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["code"])

	// 验证模拟调用
	mockService.AssertExpectations(t)
}

func TestUserHandler_Register_InvalidRequest(t *testing.T) {
	// 设置 Gin 测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockUserService)
	log := logger.New("test", "debug", true)

	// 创建处理器
	handler := NewUserHandler(mockService, log)

	// 创建无效的测试请求
	reqBody := map[string]string{
		"name": "测试用户",
		// 缺少必需的 email 和 password 字段
	}
	reqBytes, _ := json.Marshal(reqBody)

	// 创建 HTTP 请求
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 执行处理器
	handler.Register(c)

	// 验证响应 - 应该返回 400 错误
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(400), response["code"])
}

func TestUserHandler_List(t *testing.T) {
	// 设置 Gin 测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockUserService)
	log := logger.New("test", "debug", true)

	// 创建处理器
	handler := NewUserHandler(mockService, log)

	// 创建 HTTP 请求
	req := httptest.NewRequest("GET", "/api/v1/users?page=1&page_size=10", nil)

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 设置模拟行为
	mockUsers := []*domain.User{
		{
			ID:     1,
			Name:   "用户1",
			Email:  "user1@example.com",
			Status: domain.StatusActive,
		},
		{
			ID:     2,
			Name:   "用户2",
			Email:  "user2@example.com",
			Status: domain.StatusActive,
		},
	}
	mockService.On("List", mock.Anything, mock.AnythingOfType("*domain.UserQuery")).Return(mockUsers, int64(2), nil)

	// 执行处理器
	handler.List(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["code"])

	// 验证模拟调用
	mockService.AssertExpectations(t)
}
