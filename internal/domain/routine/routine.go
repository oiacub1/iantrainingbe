package routine

import "time"

type RoutineStatus string

const (
	RoutineStatusDraft     RoutineStatus = "DRAFT"
	RoutineStatusActive    RoutineStatus = "ACTIVE"
	RoutineStatusCompleted RoutineStatus = "COMPLETED"
	RoutineStatusArchived  RoutineStatus = "ARCHIVED"
)

type Routine struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	TrainerID   string        `json:"trainerId"`
	Status      RoutineStatus `json:"status"`
	WeekCount   int           `json:"weekCount"`
	Description string        `json:"description"`
	WorkoutDays []WorkoutDay  `json:"workoutDays"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

type WorkoutDay struct {
	WeekNumber int           `json:"weekNumber"`
	DayNumber  int           `json:"dayNumber"`
	DayName    string        `json:"dayName"`
	IsRestDay  bool          `json:"isRestDay"`
	Exercises  []ExerciseSet `json:"exercises"`
}

type ExerciseSet struct {
	ExerciseID  string `json:"exerciseId"`
	Order       int    `json:"order"`
	Sets        int    `json:"sets"`
	Reps        string `json:"reps"`
	RestSeconds int    `json:"restSeconds"`
	Notes       string `json:"notes"`
	Tempo       string `json:"tempo"`
	RPE         int    `json:"rpe"`
}

type CreateRoutineRequest struct {
	Name        string `json:"name"`
	WeekCount   int    `json:"weekCount"`
	Description string `json:"description"`
}

type UpdateRoutineRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      RoutineStatus `json:"status"`
	WorkoutDays []WorkoutDay  `json:"workoutDays,omitempty"`
}

type CreateWorkoutDayRequest struct {
	RoutineID  string        `json:"routineId"`
	WeekNumber int           `json:"weekNumber"`
	DayNumber  int           `json:"dayNumber"`
	DayName    string        `json:"dayName"`
	IsRestDay  bool          `json:"isRestDay"`
	Exercises  []ExerciseSet `json:"exercises"`
}

func (s RoutineStatus) IsValid() bool {
	switch s {
	case RoutineStatusDraft, RoutineStatusActive, RoutineStatusCompleted, RoutineStatusArchived:
		return true
	default:
		return false
	}
}
