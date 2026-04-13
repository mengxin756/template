package service

import (
	"context"
	"testing"
	"time"

	"example.com/classic/internal/domain"
	"example.com/classic/internal/handler/request"
	"example.com/classic/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository 模拟用户仓储
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, params domain.UserListParams) ([]*domain.User, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Save(ctx context.Context, aggregate *domain.UserAggregate) error {
	args := m.Called(ctx, aggregate)
	return args.Error(0)
}

func (m *MockUserRepository) GetAggregateByID(ctx context.Context, id int) (*domain.UserAggregate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserAggregate), args.Error(1)
}

func (m *MockUserRepository) GetAggregateByEmail(ctx context.Context, email string) (*domain.UserAggregate, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserAggregate), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func TestUserService_Register(t *testing.T) {
	// Test cases
	tests := []struct {
		name    string
		req     *request.CreateUserRequest
		setup   func(*MockUserRepository, *MockUserFactory)
		wantErr bool
	}{
		{
			name: "successfully register user",
			req: &request.CreateUserRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func(mockRepo *MockUserRepository, mockFactory *MockUserFactory) {
				mockRepo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
				mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.UserAggregate")).Return(nil)
				mockFactory.On("CreateNewUser", "Test User", "test@example.com", "password123").Return(createTestAggregate(1, "Test User", "test@example.com"), nil)
			},
			wantErr: false,
		},
		{
			name: "email already exists",
			req: &request.CreateUserRequest{
				Name:     "Test User",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setup: func(mockRepo *MockUserRepository, mockFactory *MockUserFactory) {
				mockRepo.On("ExistsByEmail", mock.Anything, "existing@example.com").Return(true, nil)
				mockFactory.On("CreateNewUser", "Test User", "existing@example.com", "password123").Return(createTestAggregate(0, "Test User", "existing@example.com"), nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository and factory for each test case
			mockRepo := new(MockUserRepository)
			mockFactory := new(MockUserFactory)
			log := logger.New("test", "debug", true)

			// Create service instance (pass nil as task queue to avoid complexity)
			svc := NewUserService(mockRepo, mockFactory, nil, nil, log)

			// Setup mock behavior
			tt.setup(mockRepo, mockFactory)

			// Execute test
			user, err := svc.Register(context.Background(), tt.req)

			// Verify results
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}

			// Verify mock calls
			mockRepo.AssertExpectations(t)
			mockFactory.AssertExpectations(t)
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	log := logger.New("test", "debug", true)
	svc := NewUserService(mockRepo, nil, nil, nil, log)

	// Setup mock behavior
	mockRepo.On("GetByID", mock.Anything, 1).Return(createTestUser(1, "Test User", "test@example.com"), nil)

	// Execute test
	user, err := svc.GetByID(context.Background(), 1)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, user)

	mockRepo.AssertExpectations(t)
}

func TestUserService_Update(t *testing.T) {
	mockRepo := new(MockUserRepository)
	log := logger.New("test", "debug", true)
	svc := NewUserService(mockRepo, nil, nil, nil, log)

	// Update request
	newName := "New Name"
	updateReq := &request.UpdateUserRequest{
		Name: &newName,
	}

	// Setup mock behavior
	mockRepo.On("GetAggregateByID", mock.Anything, 1).Return(createTestAggregate(1, "Old Name", "old@example.com"), nil)
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.UserAggregate")).Return(nil)

	// Execute test
	user, err := svc.Update(context.Background(), 1, updateReq)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, user)

	mockRepo.AssertExpectations(t)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

// MockUserFactory mock user factory
func NewMockUserFactory() *MockUserFactory {
	return &MockUserFactory{}
}

type MockUserFactory struct {
	mock.Mock
}

func (m *MockUserFactory) CreateNewUser(name, email, password string) (*domain.UserAggregate, error) {
	args := m.Called(name, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserAggregate), args.Error(1)
}

// createTestUser creates a test user entity
func createTestUser(id int, name, email string) *domain.User {
	nameVO, _ := domain.NewName(name)
	emailVO, _ := domain.NewEmail(email)
	passwordVO, _ := domain.NewHashedPassword("hashed_password")
	user, _ := domain.NewUser(id, *nameVO, *emailVO, *passwordVO, domain.StatusActive, time.Now(), time.Now())
	return user
}

// createTestAggregate creates a test user aggregate
func createTestAggregate(id int, name, email string) *domain.UserAggregate {
	user := createTestUser(id, name, email)
	return domain.RebuildUserAggregate(user)
}
