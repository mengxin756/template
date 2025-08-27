package user

import (
	"context"

	"example.com/classic/internal/data/ent"
)

type Repository interface {
	Get(ctx context.Context, id int) (*ent.User, error)
	Create(ctx context.Context, name, email string) (*ent.User, error)
}

type EntRepository struct {
	client *ent.Client
}

func NewRepository(client *ent.Client) *EntRepository { return &EntRepository{client: client} }

func (r *EntRepository) Get(ctx context.Context, id int) (*ent.User, error) {
	return r.client.User.Get(ctx, id)
}

func (r *EntRepository) Create(ctx context.Context, name, email string) (*ent.User, error) {
	return r.client.User.Create().SetName(name).SetEmail(email).Save(ctx)
}
