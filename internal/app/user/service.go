package user

import (
    "context"
)

type Service interface {
    Get(ctx context.Context, id int) (*DTO, error)
    Create(ctx context.Context, in CreateInput) (*DTO, error)
}

type CreateInput struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type DTO struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type service struct {
    repo Repository
}

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Get(ctx context.Context, id int) (*DTO, error) {
    u, err := s.repo.Get(ctx, id)
    if err != nil {
        return nil, err
    }
    return &DTO{ID: u.ID, Name: u.Name, Email: u.Email}, nil
}

func (s *service) Create(ctx context.Context, in CreateInput) (*DTO, error) {
    u, err := s.repo.Create(ctx, in.Name, in.Email)
    if err != nil {
        return nil, err
    }
    return &DTO{ID: u.ID, Name: u.Name, Email: u.Email}, nil
}


