package exercise

import "context"

type Repository interface {
	Create(ctx context.Context, exercise *Exercise) error
	GetByID(ctx context.Context, exerciseID string) (*Exercise, error)
	ListByTrainer(ctx context.Context, trainerID string, limit int, lastKey string) ([]*Exercise, string, error)
	Update(ctx context.Context, exercise *Exercise) error
	Delete(ctx context.Context, exerciseID string) error
}
