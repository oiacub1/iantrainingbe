package user

import (
	"context"
	"regexp"

	"iantraining/internal/domain/user"
)

type Service struct {
	repo user.Repository
}

func NewService(repo user.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTrainer(ctx context.Context, trainer *user.Trainer) error {
	if err := s.validateUser(&trainer.User); err != nil {
		return err
	}

	return s.repo.CreateTrainer(ctx, trainer)
}

func (s *Service) GetTrainer(ctx context.Context, trainerID string) (*user.Trainer, error) {
	if trainerID == "" {
		return nil, user.ErrInvalidUserID
	}

	return s.repo.GetTrainer(ctx, trainerID)
}

func (s *Service) CreateStudent(ctx context.Context, student *user.Student) error {
	if err := s.validateUser(&student.User); err != nil {
		return err
	}

	if err := s.validateStudentMetadata(&student.Metadata); err != nil {
		return err
	}

	return s.repo.CreateStudent(ctx, student)
}

func (s *Service) GetStudent(ctx context.Context, studentID string) (*user.Student, error) {
	if studentID == "" {
		return nil, user.ErrInvalidUserID
	}

	return s.repo.GetStudent(ctx, studentID)
}

func (s *Service) AssignStudentToTrainer(ctx context.Context, trainerID, studentID string) error {
	if trainerID == "" || studentID == "" {
		return user.ErrInvalidUserID
	}

	if trainerID == studentID {
		return user.ErrCannotAssignSelf
	}

	_, err := s.repo.GetTrainer(ctx, trainerID)
	if err != nil {
		return err
	}

	student, err := s.repo.GetStudent(ctx, studentID)
	if err != nil {
		return err
	}

	relation := &user.TrainerStudent{
		TrainerID:    trainerID,
		StudentID:    studentID,
		StudentName:  student.Name,
		StudentEmail: student.Email,
		Status:       user.StatusActive,
	}

	if err := s.repo.AssignStudentToTrainer(ctx, relation); err != nil {
		return err
	}

	student.TrainerID = trainerID
	return s.repo.UpdateStudent(ctx, student)
}

func (s *Service) ListStudentsByTrainer(ctx context.Context, trainerID string, limit int, lastKey string) ([]*user.TrainerStudent, string, error) {
	if trainerID == "" {
		return nil, "", user.ErrInvalidUserID
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return s.repo.ListStudentsByTrainer(ctx, trainerID, limit, lastKey)
}

func (s *Service) GetTrainerByStudent(ctx context.Context, studentID string) (*user.Trainer, error) {
	if studentID == "" {
		return nil, user.ErrInvalidUserID
	}

	return s.repo.GetTrainerByStudent(ctx, studentID)
}

func (s *Service) RemoveStudentFromTrainer(ctx context.Context, trainerID, studentID string) error {
	if trainerID == "" || studentID == "" {
		return user.ErrInvalidUserID
	}

	return s.repo.RemoveStudentFromTrainer(ctx, trainerID, studentID)
}

func (s *Service) validateUser(u *user.User) error {
	if u.Name == "" {
		return user.ErrNameRequired
	}

	if !isValidEmail(u.Email) {
		return user.ErrInvalidEmail
	}

	if !u.Role.IsValid() {
		return user.ErrInvalidRole
	}

	if !u.Status.IsValid() {
		return user.ErrInvalidStatus
	}

	return nil
}

func (s *Service) validateStudentMetadata(metadata *user.StudentMetadata) error {
	validLevels := map[string]bool{
		user.FitnessLevelBeginner:     true,
		user.FitnessLevelIntermediate: true,
		user.FitnessLevelAdvanced:     true,
	}

	if metadata.FitnessLevel != "" && !validLevels[metadata.FitnessLevel] {
		return user.ErrInvalidStatus
	}

	return nil
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
