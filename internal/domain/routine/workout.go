package routine

import "time"

type WorkoutLogStatus string

const (
	WorkoutLogStatusNotStarted WorkoutLogStatus = "NOT_STARTED"
	WorkoutLogStatusInProgress WorkoutLogStatus = "IN_PROGRESS"
	WorkoutLogStatusCompleted  WorkoutLogStatus = "COMPLETED"
	WorkoutLogStatusSkipped    WorkoutLogStatus = "SKIPPED"
)

type WorkoutLog struct {
	ID                string          `json:"id"`
	StudentID         string          `json:"studentId"`
	RoutineID         string          `json:"routineId"`
	ExerciseID        string          `json:"exerciseId"`
	ExerciseName      string          `json:"exerciseName"`
	WeekNumber        int             `json:"weekNumber"`
	DayNumber         int             `json:"dayNumber"`
	CompletedAt       time.Time       `json:"completedAt"`
	Date              string          `json:"date"`
	Sets              []WorkoutSet    `json:"sets"`
	TotalDurationSec  int             `json:"totalDurationSeconds"`
	Feeling           WorkoutFeeling  `json:"feeling"`
	Notes             string          `json:"notes"`
	Status            WorkoutLogStatus `json:"status"`
}

type WorkoutSet struct {
	SetNumber int     `json:"setNumber"`
	Reps      int     `json:"reps"`
	Weight    float64 `json:"weight"`
	WeightUnit string `json:"weightUnit"`
	Completed bool    `json:"completed"`
	RPE       int     `json:"rpe"`
	Notes     string  `json:"notes"`
}

type WorkoutFeeling string

const (
	FeelingExcellent WorkoutFeeling = "EXCELLENT"
	FeelingGood      WorkoutFeeling = "GOOD"
	FeelingAverage   WorkoutFeeling = "AVERAGE"
	FeelingPoor      WorkoutFeeling = "POOR"
)

type DailySummary struct {
	StudentID           string  `json:"studentId"`
	RoutineID           string  `json:"routineId"`
	Date                string  `json:"date"`
	WeekNumber          int     `json:"weekNumber"`
	DayNumber           int     `json:"dayNumber"`
	TotalExercises      int     `json:"totalExercises"`
	CompletedExercises  int     `json:"completedExercises"`
	TotalDurationSec    int     `json:"totalDurationSeconds"`
	CompletionPercentage float64 `json:"completionPercentage"`
	OverallFeeling      WorkoutFeeling `json:"overallFeeling"`
	Notes               string  `json:"notes"`
	CompletedAt         time.Time `json:"completedAt"`
}

type CreateWorkoutLogRequest struct {
	RoutineID    string      `json:"routineId"`
	ExerciseID   string      `json:"exerciseId"`
	WeekNumber   int         `json:"weekNumber"`
	DayNumber    int         `json:"dayNumber"`
	Date         string      `json:"date"`
	Sets         []WorkoutSet `json:"sets"`
	Feeling      WorkoutFeeling `json:"feeling"`
	Notes        string      `json:"notes"`
}

func (s WorkoutLogStatus) IsValid() bool {
	switch s {
	case WorkoutLogStatusNotStarted, WorkoutLogStatusInProgress, WorkoutLogStatusCompleted, WorkoutLogStatusSkipped:
		return true
	default:
		return false
	}
}

func (f WorkoutFeeling) IsValid() bool {
	switch f {
	case FeelingExcellent, FeelingGood, FeelingAverage, FeelingPoor:
		return true
	default:
		return false
	}
}
