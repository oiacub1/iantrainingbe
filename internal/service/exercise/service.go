package exercise

import (
	"context"
	"errors"

	"iantraining/internal/domain/exercise"
	"iantraining/internal/repository/dynamodb"
)

type Service struct {
	repo *dynamodb.ExerciseRepository
}

func NewService(repo *dynamodb.ExerciseRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, ex *exercise.Exercise) error {
	if ex == nil {
		return errors.New("exercise cannot be nil")
	}

	if ex.Name == "" {
		return errors.New("exercise name is required")
	}

	if ex.CreatedBy == "" {
		return errors.New("trainer ID is required")
	}

	return s.repo.Create(ctx, ex)
}

func (s *Service) GetByID(ctx context.Context, id string) (*exercise.Exercise, error) {
	if id == "" {
		return nil, errors.New("exercise ID is required")
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByTrainer(ctx context.Context, trainerID string, limit int, lastKey string) ([]*exercise.Exercise, string, error) {
	if trainerID == "" {
		return nil, "", errors.New("trainer ID is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return s.repo.ListByTrainer(ctx, trainerID, limit, lastKey)
}

func (s *Service) Update(ctx context.Context, ex *exercise.Exercise) error {
	if ex == nil {
		return errors.New("exercise cannot be nil")
	}

	if ex.ID == "" {
		return errors.New("exercise ID is required")
	}

	return s.repo.Update(ctx, ex)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("exercise ID is required")
	}

	return s.repo.Delete(ctx, id)
}
