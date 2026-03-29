package dynamodb

import (
	"fmt"
	"time"

	"iantraining/internal/domain/routine"
)

type DynamoRoutine struct {
	PK          string    `dynamodbav:"PK"`
	SK          string    `dynamodbav:"SK"`
	ID          string    `dynamodbav:"id"`
	Name        string    `dynamodbav:"name"`
	TrainerID   string    `dynamodbav:"trainerId"`
	Description string    `dynamodbav:"description"`
	Status      string    `dynamodbav:"status"`
	WeekCount   int       `dynamodbav:"weekCount"`
	CreatedAt   time.Time `dynamodbav:"createdAt"`
	UpdatedAt   time.Time `dynamodbav:"updatedAt"`
	EntityType  string    `dynamodbav:"entityType"`
	GSI1PK      string    `dynamodbav:"GSI1PK"`
	GSI1SK      string    `dynamodbav:"GSI1SK"`
	GSI2PK      string    `dynamodbav:"GSI2PK"`
	GSI2SK      string    `dynamodbav:"GSI2SK"`
}

type DynamoWorkoutDay struct {
	PK         string              `dynamodbav:"PK"`
	SK         string              `dynamodbav:"SK"`
	RoutineID  string              `dynamodbav:"routineId"`
	WeekNumber int                 `dynamodbav:"weekNumber"`
	DayNumber  int                 `dynamodbav:"dayNumber"`
	DayName    string              `dynamodbav:"dayName"`
	IsRestDay  bool                `dynamodbav:"isRestDay"`
	Exercises  []DynamoExerciseSet `dynamodbav:"exercises"`
	EntityType string              `dynamodbav:"entityType"`
}

type DynamoExerciseSet struct {
	ExerciseID  string `dynamodbav:"exerciseId"`
	Order       int    `dynamodbav:"order"`
	Sets        int    `dynamodbav:"sets"`
	Reps        string `dynamodbav:"reps"`
	RestSeconds int    `dynamodbav:"restSeconds"`
	Notes       string `dynamodbav:"notes"`
	Tempo       string `dynamodbav:"tempo"`
	RPE         int    `dynamodbav:"rpe"`
}

type DynamoWorkoutLog struct {
	ID               string             `dynamodbav:"id"`
	StudentID        string             `dynamodbav:"studentId"`
	RoutineID        string             `dynamodbav:"routineId"`
	ExerciseID       string             `dynamodbav:"exerciseId"`
	ExerciseName     string             `dynamodbav:"exerciseName"`
	WeekNumber       int                `dynamodbav:"weekNumber"`
	DayNumber        int                `dynamodbav:"dayNumber"`
	CompletedAt      time.Time          `dynamodbav:"completedAt"`
	Date             string             `dynamodbav:"date"`
	Sets             []DynamoWorkoutSet `dynamodbav:"sets"`
	TotalDurationSec int                `dynamodbav:"totalDurationSeconds"`
	Feeling          string             `dynamodbav:"feeling"`
	Notes            string             `dynamodbav:"notes"`
	Status           string             `dynamodbav:"status"`
	EntityType       string             `dynamodbav:"entityType"`
	GSI1PK           string             `dynamodbav:"GSI1PK"`
	GSI1SK           string             `dynamodbav:"GSI1SK"`
	GSI2PK           string             `dynamodbav:"GSI2PK"`
	GSI2SK           string             `dynamodbav:"GSI2SK"`
}

type DynamoWorkoutSet struct {
	SetNumber  int     `dynamodbav:"setNumber"`
	Reps       int     `dynamodbav:"reps"`
	Weight     float64 `dynamodbav:"weight"`
	WeightUnit string  `dynamodbav:"weightUnit"`
	Completed  bool    `dynamodbav:"completed"`
	RPE        int     `dynamodbav:"rpe"`
	Notes      string  `dynamodbav:"notes"`
}

type DynamoDailySummary struct {
	StudentID            string    `dynamodbav:"studentId"`
	RoutineID            string    `dynamodbav:"routineId"`
	Date                 string    `dynamodbav:"date"`
	WeekNumber           int       `dynamodbav:"weekNumber"`
	DayNumber            int       `dynamodbav:"dayNumber"`
	TotalExercises       int       `dynamodbav:"totalExercises"`
	CompletedExercises   int       `dynamodbav:"completedExercises"`
	TotalDurationSec     int       `dynamodbav:"totalDurationSeconds"`
	CompletionPercentage float64   `dynamodbav:"completionPercentage"`
	OverallFeeling       string    `dynamodbav:"overallFeeling"`
	Notes                string    `dynamodbav:"notes"`
	CompletedAt          time.Time `dynamodbav:"completedAt"`
	EntityType           string    `dynamodbav:"entityType"`
	GSI1PK               string    `dynamodbav:"GSI1PK"`
	GSI1SK               string    `dynamodbav:"GSI1SK"`
}

func routineToDynamoItem(r *routine.Routine) *DynamoRoutine {
	return &DynamoRoutine{
		PK:          "ROUTINE#" + r.ID,
		SK:          "ROUTINE#" + r.ID,
		ID:          r.ID,
		Name:        r.Name,
		TrainerID:   r.TrainerID,
		Description: r.Description,
		Status:      string(r.Status),
		WeekCount:   r.WeekCount,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		EntityType:  EntityTypeRoutine,
		GSI1PK:      "TRAINER#" + r.TrainerID,
		GSI1SK:      "ROUTINE#" + r.CreatedAt.Format(time.RFC3339),
		GSI2PK:      "TRAINER#" + r.TrainerID,
		GSI2SK:      "ROUTINE#" + r.CreatedAt.Format(time.RFC3339),
	}
}

func dynamoItemToRoutine(d *DynamoRoutine) *routine.Routine {
	return &routine.Routine{
		ID:          d.ID,
		Name:        d.Name,
		TrainerID:   d.TrainerID,
		Description: d.Description,
		Status:      routine.RoutineStatus(d.Status),
		WeekCount:   d.WeekCount,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func workoutDayToDynamoItem(w *routine.WorkoutDay) *DynamoWorkoutDay {
	exercises := make([]DynamoExerciseSet, len(w.Exercises))
	for i, ex := range w.Exercises {
		exercises[i] = DynamoExerciseSet{
			ExerciseID:  ex.ExerciseID,
			Order:       ex.Order,
			Sets:        ex.Sets,
			Reps:        ex.Reps,
			RestSeconds: ex.RestSeconds,
			Notes:       ex.Notes,
			Tempo:       ex.Tempo,
			RPE:         ex.RPE,
		}
	}

	return &DynamoWorkoutDay{
		PK:         "ROUTINE#" + w.RoutineID,
		SK:         fmt.Sprintf("WORKOUTDAY#W%dD%d", w.WeekNumber, w.DayNumber),
		RoutineID:  w.RoutineID,
		WeekNumber: w.WeekNumber,
		DayNumber:  w.DayNumber,
		DayName:    w.DayName,
		IsRestDay:  w.IsRestDay,
		Exercises:  exercises,
		EntityType: EntityTypeWorkoutDay,
	}
}

func dynamoItemToWorkoutDay(d *DynamoWorkoutDay) *routine.WorkoutDay {
	exercises := make([]routine.ExerciseSet, len(d.Exercises))
	for i, ex := range d.Exercises {
		exercises[i] = routine.ExerciseSet{
			ExerciseID:  ex.ExerciseID,
			Order:       ex.Order,
			Sets:        ex.Sets,
			Reps:        ex.Reps,
			RestSeconds: ex.RestSeconds,
			Notes:       ex.Notes,
			Tempo:       ex.Tempo,
			RPE:         ex.RPE,
		}
	}

	return &routine.WorkoutDay{
		RoutineID:  d.RoutineID,
		WeekNumber: d.WeekNumber,
		DayNumber:  d.DayNumber,
		DayName:    d.DayName,
		IsRestDay:  d.IsRestDay,
		Exercises:  exercises,
	}
}

func workoutLogToDynamoItem(w *routine.WorkoutLog) *DynamoWorkoutLog {
	sets := make([]DynamoWorkoutSet, len(w.Sets))
	for i, set := range w.Sets {
		sets[i] = DynamoWorkoutSet{
			SetNumber:  set.SetNumber,
			Reps:       set.Reps,
			Weight:     set.Weight,
			WeightUnit: set.WeightUnit,
			Completed:  set.Completed,
			RPE:        set.RPE,
			Notes:      set.Notes,
		}
	}

	return &DynamoWorkoutLog{
		ID:               w.ID,
		StudentID:        w.StudentID,
		RoutineID:        w.RoutineID,
		ExerciseID:       w.ExerciseID,
		ExerciseName:     w.ExerciseName,
		WeekNumber:       w.WeekNumber,
		DayNumber:        w.DayNumber,
		CompletedAt:      w.CompletedAt,
		Date:             w.Date,
		Sets:             sets,
		TotalDurationSec: w.TotalDurationSec,
		Feeling:          string(w.Feeling),
		Notes:            w.Notes,
		Status:           string(w.Status),
		EntityType:       EntityTypeWorkoutLog,
		GSI1PK:           "ROUTINE#" + w.RoutineID,
		GSI1SK:           "WORKOUT#" + string(rune(w.CompletedAt.Unix())),
		GSI2PK:           "STUDENT#" + w.StudentID + "#DATE#" + w.Date,
		GSI2SK:           "WORKOUT#" + string(rune(w.CompletedAt.Unix())),
	}
}

func dynamoItemToWorkoutLog(d *DynamoWorkoutLog) *routine.WorkoutLog {
	sets := make([]routine.WorkoutSet, len(d.Sets))
	for i, set := range d.Sets {
		sets[i] = routine.WorkoutSet{
			SetNumber:  set.SetNumber,
			Reps:       set.Reps,
			Weight:     set.Weight,
			WeightUnit: set.WeightUnit,
			Completed:  set.Completed,
			RPE:        set.RPE,
			Notes:      set.Notes,
		}
	}

	return &routine.WorkoutLog{
		ID:               d.ID,
		StudentID:        d.StudentID,
		RoutineID:        d.RoutineID,
		ExerciseID:       d.ExerciseID,
		ExerciseName:     d.ExerciseName,
		WeekNumber:       d.WeekNumber,
		DayNumber:        d.DayNumber,
		CompletedAt:      d.CompletedAt,
		Date:             d.Date,
		Sets:             sets,
		TotalDurationSec: d.TotalDurationSec,
		Feeling:          routine.WorkoutFeeling(d.Feeling),
		Notes:            d.Notes,
		Status:           routine.WorkoutLogStatus(d.Status),
	}
}

func dailySummaryToDynamoItem(d *routine.DailySummary) *DynamoDailySummary {
	return &DynamoDailySummary{
		StudentID:            d.StudentID,
		RoutineID:            d.RoutineID,
		Date:                 d.Date,
		WeekNumber:           d.WeekNumber,
		DayNumber:            d.DayNumber,
		TotalExercises:       d.TotalExercises,
		CompletedExercises:   d.CompletedExercises,
		TotalDurationSec:     d.TotalDurationSec,
		CompletionPercentage: d.CompletionPercentage,
		OverallFeeling:       string(d.OverallFeeling),
		Notes:                d.Notes,
		CompletedAt:          d.CompletedAt,
		EntityType:           EntityTypeDailySummary,
		GSI1PK:               "ROUTINE#" + d.RoutineID,
		GSI1SK:               "SUMMARY#" + d.Date,
	}
}

func dynamoItemToDailySummary(d *DynamoDailySummary) *routine.DailySummary {
	return &routine.DailySummary{
		StudentID:            d.StudentID,
		RoutineID:            d.RoutineID,
		Date:                 d.Date,
		WeekNumber:           d.WeekNumber,
		DayNumber:            d.DayNumber,
		TotalExercises:       d.TotalExercises,
		CompletedExercises:   d.CompletedExercises,
		TotalDurationSec:     d.TotalDurationSec,
		CompletionPercentage: d.CompletionPercentage,
		OverallFeeling:       routine.WorkoutFeeling(d.OverallFeeling),
		Notes:                d.Notes,
		CompletedAt:          d.CompletedAt,
	}
}
