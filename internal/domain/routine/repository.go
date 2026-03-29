package routine

import (
	"context"
	"time"
)

type Repository interface {
	// Routine operations (now stateless templates)
	CreateRoutine(ctx context.Context, routine *Routine) error
	GetRoutine(ctx context.Context, id string) (*Routine, error)
	UpdateRoutine(ctx context.Context, routine *Routine) error
	DeleteRoutine(ctx context.Context, id string) error
	ListRoutinesByTrainer(ctx context.Context, trainerID string, limit int, startKey string) ([]*Routine, string, error)

	// Workout Day operations
	CreateWorkoutDay(ctx context.Context, workoutDay *WorkoutDay) error
	GetWorkoutDays(ctx context.Context, routineID string) ([]*WorkoutDay, error)
	UpdateWorkoutDay(ctx context.Context, workoutDay *WorkoutDay) error
	DeleteWorkoutDays(ctx context.Context, routineID string) error

	// Workout Log operations
	CreateWorkoutLog(ctx context.Context, log *WorkoutLog) error
	GetWorkoutLogs(ctx context.Context, studentID string, startDate, endDate time.Time) ([]*WorkoutLog, error)
	GetWorkoutLogsByRoutine(ctx context.Context, routineID string) ([]*WorkoutLog, error)
	UpdateWorkoutLog(ctx context.Context, log *WorkoutLog) error

	// Daily Summary operations
	CreateDailySummary(ctx context.Context, summary *DailySummary) error
	GetDailySummaries(ctx context.Context, studentID string, startDate, endDate time.Time) ([]*DailySummary, error)
	GetDailySummary(ctx context.Context, studentID string, date string) (*DailySummary, error)
}
