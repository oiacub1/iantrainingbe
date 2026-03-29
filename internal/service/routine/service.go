package routine

import (
	"context"
	"fmt"

	"iantraining/internal/domain/routine"
	"iantraining/internal/domain/user"
)

type Service struct {
	routineRepo routine.Repository
	userRepo    user.Repository
}

func NewService(routineRepo routine.Repository, userRepo user.Repository) *Service {
	return &Service{
		routineRepo: routineRepo,
		userRepo:    userRepo,
	}
}

func (s *Service) CreateRoutine(ctx context.Context, trainerID string, req *routine.CreateRoutineRequest) (*routine.Routine, error) {
	if err := s.validateCreateRoutineRequest(req); err != nil {
		return nil, err
	}

	routineEntity := &routine.Routine{
		Name:        req.Name,
		TrainerID:   trainerID,
		Description: req.Description,
		Status:      routine.RoutineStatusDraft,
		WeekCount:   req.WeekCount,
	}

	if err := s.routineRepo.CreateRoutine(ctx, routineEntity); err != nil {
		return nil, fmt.Errorf("failed to create routine: %w", err)
	}

	return routineEntity, nil
}

func (s *Service) GetRoutine(ctx context.Context, id string) (*routine.Routine, error) {
	routine, err := s.routineRepo.GetRoutine(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get routine: %w", err)
	}

	return routine, nil
}

func (s *Service) UpdateRoutine(ctx context.Context, trainerID, routineID string, req *routine.UpdateRoutineRequest) (*routine.Routine, error) {
	existingRoutine, err := s.routineRepo.GetRoutine(ctx, routineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing routine: %w", err)
	}

	if existingRoutine.TrainerID != trainerID {
		return nil, routine.ErrRoutineNotFound
	}

	if req.Name != "" {
		existingRoutine.Name = req.Name
	}

	if req.Description != "" {
		existingRoutine.Description = req.Description
	}

	if req.Status != "" && req.Status.IsValid() {
		existingRoutine.Status = req.Status
	}

	if err := s.routineRepo.UpdateRoutine(ctx, existingRoutine); err != nil {
		return nil, fmt.Errorf("failed to update routine: %w", err)
	}

	return existingRoutine, nil
}

func (s *Service) DeleteRoutine(ctx context.Context, trainerID, routineID string) error {
	existingRoutine, err := s.routineRepo.GetRoutine(ctx, routineID)
	if err != nil {
		return fmt.Errorf("failed to get existing routine: %w", err)
	}

	if existingRoutine.TrainerID != trainerID {
		return routine.ErrRoutineNotFound
	}

	if err := s.routineRepo.DeleteWorkoutDays(ctx, routineID); err != nil {
		return fmt.Errorf("failed to delete workout days: %w", err)
	}

	if err := s.routineRepo.DeleteRoutine(ctx, routineID); err != nil {
		return fmt.Errorf("failed to delete routine: %w", err)
	}

	return nil
}

func (s *Service) ListRoutinesByTrainer(ctx context.Context, trainerID string, limit int, startKey string) ([]*routine.Routine, string, error) {
	routines, nextKey, err := s.routineRepo.ListRoutinesByTrainer(ctx, trainerID, limit, startKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list routines by trainer: %w", err)
	}

	return routines, nextKey, nil
}

// ListRoutinesByStudent is DEPRECATED - use AssignmentService.GetAssignmentsByStudent instead
// This method is kept for backward compatibility but will be removed in future versions
func (s *Service) ListRoutinesByStudent(ctx context.Context, studentID string, limit int, startKey string) ([]*routine.Routine, string, error) {
	// This functionality has been moved to the assignment service
	// Routines are now templates and students access them through assignments
	return nil, "", fmt.Errorf("deprecated: use AssignmentService.GetAssignmentsByStudent instead")
}

// GetActiveRoutineForStudent is DEPRECATED - use AssignmentService.GetActiveAssignmentForStudent instead
// This method is kept for backward compatibility but will be removed in future versions
func (s *Service) GetActiveRoutineForStudent(ctx context.Context, studentID string) (*routine.Routine, error) {
	// This functionality has been moved to the assignment service
	// Routines are now templates and students access them through assignments
	return nil, fmt.Errorf("deprecated: use AssignmentService.GetActiveAssignmentForStudent instead")
}

func (s *Service) CreateWorkoutDay(ctx context.Context, trainerID, routineID string, workoutDay *routine.WorkoutDay) error {
	routineData, err := s.routineRepo.GetRoutine(ctx, routineID)
	if err != nil {
		return fmt.Errorf("failed to get routine: %w", err)
	}

	if routineData.TrainerID != trainerID {
		return routine.ErrRoutineNotFound
	}

	if err := s.validateWorkoutDay(workoutDay); err != nil {
		return err
	}

	workoutDay.RoutineID = routineID

	if err := s.routineRepo.CreateWorkoutDay(ctx, workoutDay); err != nil {
		return fmt.Errorf("failed to create workout day: %w", err)
	}

	return nil
}

func (s *Service) GetWorkoutDays(ctx context.Context, routineID string) ([]*routine.WorkoutDay, error) {
	workoutDays, err := s.routineRepo.GetWorkoutDays(ctx, routineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout days: %w", err)
	}

	return workoutDays, nil
}

func (s *Service) UpdateWorkoutDay(ctx context.Context, trainerID, routineID string, workoutDay *routine.WorkoutDay) error {
	routineData, err := s.routineRepo.GetRoutine(ctx, routineID)
	if err != nil {
		return fmt.Errorf("failed to get routine: %w", err)
	}

	if routineData.TrainerID != trainerID {
		return routine.ErrRoutineNotFound
	}

	if err := s.validateWorkoutDay(workoutDay); err != nil {
		return err
	}

	workoutDay.RoutineID = routineID

	if err := s.routineRepo.UpdateWorkoutDay(ctx, workoutDay); err != nil {
		return fmt.Errorf("failed to update workout day: %w", err)
	}

	return nil
}

func (s *Service) validateCreateRoutineRequest(req *routine.CreateRoutineRequest) error {
	if req.Name == "" {
		return fmt.Errorf("routine name is required")
	}
	if req.WeekCount <= 0 || req.WeekCount > 52 {
		return fmt.Errorf("week count must be between 1 and 52")
	}
	return nil
}

func (s *Service) validateTrainerStudentRelationship(ctx context.Context, trainerID, studentID string) error {
	students, _, err := s.userRepo.ListStudentsByTrainer(ctx, trainerID, 100, "")
	if err != nil {
		return fmt.Errorf("failed to validate trainer-student relationship: %w", err)
	}

	for _, student := range students {
		if student.StudentID == studentID {
			return nil
		}
	}

	return routine.ErrStudentNotAssigned
}

func (s *Service) validateWorkoutDay(workoutDay *routine.WorkoutDay) error {
	if workoutDay.WeekNumber <= 0 || workoutDay.WeekNumber > 52 {
		return routine.ErrInvalidWeekNumber
	}
	if workoutDay.DayNumber <= 0 || workoutDay.DayNumber > 7 {
		return routine.ErrInvalidDayNumber
	}
	if workoutDay.DayName == "" {
		return fmt.Errorf("day name is required")
	}

	for i, exercise := range workoutDay.Exercises {
		if exercise.ExerciseID == "" {
			return routine.ErrExerciseNotFound
		}
		if exercise.Order != i+1 {
			return fmt.Errorf("exercise order must be sequential")
		}
		if exercise.Sets <= 0 {
			return routine.ErrInvalidSetData
		}
		if exercise.Reps == "" {
			return routine.ErrInvalidSetData
		}
	}

	return nil
}
