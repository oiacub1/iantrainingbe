package routine

import "errors"

var (
	ErrRoutineNotFound      = errors.New("routine not found")
	ErrRoutineAlreadyExists = errors.New("routine already exists")
	ErrInvalidRoutineStatus = errors.New("invalid routine status")
	ErrInvalidDateRange     = errors.New("invalid date range")
	ErrStudentNotAssigned   = errors.New("student not assigned to trainer")
	ErrWorkoutDayNotFound   = errors.New("workout day not found")
	ErrWorkoutLogNotFound   = errors.New("workout log not found")
	ErrInvalidWeekNumber    = errors.New("invalid week number")
	ErrInvalidDayNumber     = errors.New("invalid day number")
	ErrExerciseNotFound     = errors.New("exercise not found")
	ErrInvalidSetData       = errors.New("invalid set data")
)
