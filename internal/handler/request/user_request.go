package request

import (
	"example.com/classic/internal/domain"
)

// CreateUserRequest create user request
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// UpdateUserRequest update user request
type UpdateUserRequest struct {
	Name   *string        `json:"name,omitempty" binding:"omitempty,min=2,max=50"`
	Email  *string        `json:"email,omitempty" binding:"omitempty,email"`
	Status *domain.Status `json:"status,omitempty" binding:"omitempty,oneof=active inactive banned"`
}

// ChangeStatusRequest change status request
type ChangeStatusRequest struct {
	Status domain.Status `json:"status" binding:"required,oneof=active inactive banned"`
}

// UserQuery user query parameters
type UserQuery struct {
	ID       *int           `form:"id,omitempty"`
	Name     *string        `form:"name,omitempty"`
	Email    *string        `form:"email,omitempty"`
	Status   *domain.Status `form:"status,omitempty"`
	Page     int            `form:"page,default=1" binding:"min=1"`
	PageSize int            `form:"page_size,default=10" binding:"min=1,max=100"`
}
