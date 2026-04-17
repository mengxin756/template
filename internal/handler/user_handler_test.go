package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/classic/internal/domain"
	"example.com/classic/internal/handler/request"
	"example.com/classic/internal/service"
	"example.com/classic/internal/service/dto"
	"example.com/classic/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService mock user service
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, params *dto.RegisterParams) (*domain.User, error) {
	args := m.Called(ctx, params)
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

func (m *MockUserService) Update(ctx context.Context, id int, params *dto.UpdateParams) (*domain.User, error) {
	args := m.Called(ctx, id, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) List(ctx context.Context, query *dto.UserQueryParams) ([]*domain.User, int64, error) {
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

// Verify MockUserService implements service.UserService
var _ service.UserService = (*MockUserService)(nil)

func TestUserHandler_Register(t *testing.T) {
	// Set Gin test mode
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockUserService)
	log := logger.New("test", "debug", true)

	// Create handler
	handler := NewUserHandler(mockService, log)

	// Create test request
	reqBody := request.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	reqBytes, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Setup mock behavior
	mockUser := createTestUser(1, "Test User", "test@example.com")
	mockService.On("Register", mock.Anything, mock.AnythingOfType("*dto.RegisterParams")).Return(mockUser, nil)

	// Execute handler
	handler.Register(c)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["code"])
	assert.Equal(t, "user registered successfully", response["msg"])

	// Verify mock calls
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetByID(t *testing.T) {
	// Set Gin test mode
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockUserService)
	log := logger.New("test", "debug", true)

	// Create handler
	handler := NewUserHandler(mockService, log)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/api/v1/users/1", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	// Setup mock behavior
	mockUser := createTestUser(1, "Test User", "test@example.com")
	mockService.On("GetByID", mock.Anything, 1).Return(mockUser, nil)

	// Execute handler
	handler.GetByID(c)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["code"])

	// Verify mock calls
	mockService.AssertExpectations(t)
}

func TestUserHandler_Register_InvalidRequest(t *testing.T) {
	// Set Gin test mode
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockUserService)
	log := logger.New("test", "debug", true)

	// Create handler
	handler := NewUserHandler(mockService, log)

	// Create invalid test request (missing required fields)
	reqBody := map[string]string{
		"name": "Test User",
		// Missing required email and password fields
	}
	reqBytes, _ := json.Marshal(reqBody)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler.Register(c)

	// Verify response - should return 400 error
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(400), response["code"])
}

func TestUserHandler_List(t *testing.T) {
	// Set Gin test mode
	gin.SetMode(gin.TestMode)

	// Create mock service
	mockService := new(MockUserService)
	log := logger.New("test", "debug", true)

	// Create handler
	handler := NewUserHandler(mockService, log)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/api/v1/users?page=1&page_size=10", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Setup mock behavior
	mockUsers := []*domain.User{
		createTestUser(1, "User1", "user1@example.com"),
		createTestUser(2, "User2", "user2@example.com"),
	}
	mockService.On("List", mock.Anything, mock.AnythingOfType("*dto.UserQueryParams")).Return(mockUsers, int64(2), nil)

	// Execute handler
	handler.List(c)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["code"])

	// Verify mock calls
	mockService.AssertExpectations(t)
}

// createTestUser creates a test user entity
func createTestUser(id int, name, email string) *domain.User {
	nameVO, _ := domain.NewName(name)
	emailVO, _ := domain.NewEmail(email)
	passwordVO, _ := domain.NewHashedPassword("hashed_password")
	user, _ := domain.NewUser(id, *nameVO, *emailVO, *passwordVO, domain.StatusActive, time.Now(), time.Now())
	return user
}
