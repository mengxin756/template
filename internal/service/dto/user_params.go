package dto

import "example.com/classic/internal/domain"

// RegisterParams 用户注册参数（service层入参，与传输层解耦）
type RegisterParams struct {
	Name     string
	Email    string
	Password string
}

// UpdateParams 用户更新参数（service层入参，与传输层解耦）
type UpdateParams struct {
	Name   *string
	Email  *string
	Status *domain.Status
}

// UserQueryParams 用户列表查询参数（service层入参，与传输层解耦）
type UserQueryParams struct {
	ID       *int
	Name     *string
	Email    *string
	Status   *domain.Status
	Page     int
	PageSize int
}
