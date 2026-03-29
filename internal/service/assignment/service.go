package assignment

import (
	"context"
	"fmt"
	"time"

	"iantraining/internal/domain/routine"
	"iantraining/internal/domain/user"
)

type AssignmentRepository interface {
	CreateAssignment(ctx context.Context, assignment *routine.RoutineAssignment) error
	GetAssignment(ctx context.Context, assignmentID, studentID string) (*routine.RoutineAssignment, error)
	UpdateAssignment(ctx context.Context, assignment *routine.RoutineAssignment) error
	DeleteAssignment(ctx context.Context, assignmentID, studentID string) error
	GetAssignmentsByStudent(ctx context.Context, studentID string) ([]*routine.RoutineAssignment, error)
	GetActiveAssignmentForStudent(ctx context.Context, studentID string) (*routine.RoutineAssignment, error)
	GetAssignmentsByRoutine(ctx context.Context, routineID string) ([]*routine.RoutineAssignment, error)
}

type Service struct {
	assignmentRepo AssignmentRepository
	routineRepo    routine.Repository
	userRepo       user.Repository
}

func NewService(assignmentRepo AssignmentRepository, routineRepo routine.Repository, userRepo user.Repository) *Service {
	return &Service{
		assignmentRepo: assignmentRepo,
		routineRepo:    routineRepo,
		userRepo:       userRepo,
	}
}

// CreateAssignments asigna una rutina a múltiples estudiantes
func (s *Service) CreateAssignments(ctx context.Context, trainerID string, req *routine.CreateAssignmentRequest) ([]*routine.RoutineAssignment, error) {
	// Validar que la rutina existe y pertenece al trainer
	routineEntity, err := s.routineRepo.GetRoutine(ctx, req.RoutineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routine: %w", err)
	}

	if routineEntity.TrainerID != trainerID {
		return nil, routine.ErrRoutineNotFound
	}

	// Las rutinas son plantillas y pueden estar en cualquier estado (DRAFT o ACTIVE)
	// La asignación es la que tendrá estado ACTIVE cuando se asigne al estudiante

	// Parsear fechas
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, routine.ErrInvalidDateRange
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, routine.ErrInvalidDateRange
	}

	if endDate.Before(startDate) {
		return nil, routine.ErrInvalidDateRange
	}

	// Crear asignaciones para cada estudiante
	assignments := make([]*routine.RoutineAssignment, 0, len(req.StudentIDs))

	for _, studentID := range req.StudentIDs {
		// Validar relación trainer-estudiante
		if err := s.validateTrainerStudentRelationship(ctx, trainerID, studentID); err != nil {
			return nil, fmt.Errorf("student %s: %w", studentID, err)
		}

		// Verificar si el estudiante ya tiene una asignación activa
		activeAssignment, err := s.assignmentRepo.GetActiveAssignmentForStudent(ctx, studentID)
		if err == nil && activeAssignment != nil {
			// Pausar la asignación activa anterior
			activeAssignment.Status = routine.AssignmentStatusPaused
			if err := s.assignmentRepo.UpdateAssignment(ctx, activeAssignment); err != nil {
				return nil, fmt.Errorf("failed to pause previous assignment for student %s: %w", studentID, err)
			}
		}

		// Crear nueva asignación
		assignment := &routine.RoutineAssignment{
			RoutineID: req.RoutineID,
			StudentID: studentID,
			StartDate: startDate,
			EndDate:   endDate,
			Status:    routine.AssignmentStatusActive,
		}

		if err := s.assignmentRepo.CreateAssignment(ctx, assignment); err != nil {
			return nil, fmt.Errorf("failed to create assignment for student %s: %w", studentID, err)
		}

		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// GetAssignment obtiene una asignación específica
func (s *Service) GetAssignment(ctx context.Context, assignmentID, studentID string) (*routine.RoutineAssignment, error) {
	assignment, err := s.assignmentRepo.GetAssignment(ctx, assignmentID, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}
	return assignment, nil
}

// UpdateAssignment actualiza una asignación existente
func (s *Service) UpdateAssignment(ctx context.Context, assignmentID, studentID string, req *routine.UpdateAssignmentRequest) (*routine.RoutineAssignment, error) {
	// Obtener asignación existente
	assignment, err := s.assignmentRepo.GetAssignment(ctx, assignmentID, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	// Actualizar campos
	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, routine.ErrInvalidDateRange
		}
		assignment.StartDate = startDate
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, routine.ErrInvalidDateRange
		}
		assignment.EndDate = endDate
	}

	if req.Status != "" && req.Status.IsValid() {
		assignment.Status = req.Status
	}

	// Validar fechas
	if assignment.EndDate.Before(assignment.StartDate) {
		return nil, routine.ErrInvalidDateRange
	}

	if err := s.assignmentRepo.UpdateAssignment(ctx, assignment); err != nil {
		return nil, fmt.Errorf("failed to update assignment: %w", err)
	}

	return assignment, nil
}

// DeleteAssignment elimina una asignación
func (s *Service) DeleteAssignment(ctx context.Context, assignmentID, studentID string) error {
	if err := s.assignmentRepo.DeleteAssignment(ctx, assignmentID, studentID); err != nil {
		return fmt.Errorf("failed to delete assignment: %w", err)
	}
	return nil
}

// GetAssignmentsByStudent obtiene todas las asignaciones de un estudiante
func (s *Service) GetAssignmentsByStudent(ctx context.Context, studentID string) ([]*routine.RoutineAssignment, error) {
	assignments, err := s.assignmentRepo.GetAssignmentsByStudent(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments by student: %w", err)
	}
	return assignments, nil
}

// GetActiveAssignmentForStudent obtiene la asignación activa de un estudiante
func (s *Service) GetActiveAssignmentForStudent(ctx context.Context, studentID string) (*routine.RoutineAssignment, error) {
	assignment, err := s.assignmentRepo.GetActiveAssignmentForStudent(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active assignment: %w", err)
	}
	return assignment, nil
}

// GetAssignmentsByRoutine obtiene todas las asignaciones de una rutina
func (s *Service) GetAssignmentsByRoutine(ctx context.Context, routineID string) ([]*routine.RoutineAssignment, error) {
	assignments, err := s.assignmentRepo.GetAssignmentsByRoutine(ctx, routineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments by routine: %w", err)
	}
	return assignments, nil
}

// GetStudentsWithRoutine obtiene los estudiantes que tienen asignada una rutina específica
func (s *Service) GetStudentsWithRoutine(ctx context.Context, routineID string) ([]string, error) {
	assignments, err := s.assignmentRepo.GetAssignmentsByRoutine(ctx, routineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments: %w", err)
	}

	studentIDs := make([]string, 0, len(assignments))
	for _, assignment := range assignments {
		studentIDs = append(studentIDs, assignment.StudentID)
	}

	return studentIDs, nil
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
