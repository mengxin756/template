package dto

import (
	"time"

	"example.com/classic/internal/domain"
)

// UserDTO user data transfer object for API responses
type UserDTO struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	Email     string       `json:"email"`
	Status    domain.Status `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// FromUser creates UserDTO from domain User entity
func UserDTOFromUser(user *domain.User) *UserDTO {
	if user == nil {
		return nil
	}
	return &UserDTO{
		ID:        user.ID(),
		Name:      user.Name(),
		Email:     user.Email(),
		Status:    user.Status(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	}
}

// FromUsers creates UserDTO slice from domain User slice
func UserDTOFromUsers(users []*domain.User) []*UserDTO {
	if users == nil {
		return nil
	}
	dtos := make([]*UserDTO, len(users))
	for i, user := range users {
		dtos[i] = UserDTOFromUser(user)
	}
	return dtos
}

// UserListDTO user list response with pagination
type UserListDTO struct {
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Users    []*UserDTO `json:"users"`
}
