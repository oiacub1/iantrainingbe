package dynamodb

import "fmt"

const (
	EntityTypeUser              = "USER"
	EntityTypeTrainer           = "TRAINER"
	EntityTypeStudent           = "STUDENT"
	EntityTypeTrainerStudent    = "TRAINER_STUDENT"
	EntityTypeExercise          = "EXERCISE"
	EntityTypeRoutine           = "ROUTINE"
	EntityTypeRoutineAssignment = "ROUTINE_ASSIGNMENT"
	EntityTypeWorkoutDay        = "WORKOUT_DAY"
	EntityTypeWorkoutLog        = "WORKOUT_LOG"
	EntityTypeDailySummary      = "DAILY_SUMMARY"
)

type KeyBuilder struct{}

func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{}
}

func (kb *KeyBuilder) UserPK(userID string) string {
	return fmt.Sprintf("USER#%s", userID)
}

func (kb *KeyBuilder) ProfileSK() string {
	return "PROFILE"
}

func (kb *KeyBuilder) StudentSK(studentID string) string {
	return fmt.Sprintf("STUDENT#%s", studentID)
}

func (kb *KeyBuilder) TrainerSK(trainerID string) string {
	return fmt.Sprintf("TRAINER#%s", trainerID)
}

func (kb *KeyBuilder) ExercisePK(exerciseID string) string {
	return fmt.Sprintf("EXERCISE#%s", exerciseID)
}

func (kb *KeyBuilder) MetadataSK() string {
	return "METADATA"
}

func (kb *KeyBuilder) ExercisesByTrainerGSI1PK(trainerID string) string {
	return fmt.Sprintf("EXERCISES_BY_TRAINER#%s", trainerID)
}

func (kb *KeyBuilder) ExerciseGSI1SK(timestamp int64) string {
	return fmt.Sprintf("EXERCISE#%d", timestamp)
}

func (kb *KeyBuilder) RoutinePK(routineID string) string {
	return fmt.Sprintf("ROUTINE#%s", routineID)
}

func (kb *KeyBuilder) WorkoutDaySK(weekNumber, dayNumber int) string {
	return fmt.Sprintf("WEEK#%d#DAY#%d", weekNumber, dayNumber)
}

func (kb *KeyBuilder) StudentGSI1PK(studentID string) string {
	return fmt.Sprintf("STUDENT#%s", studentID)
}

func (kb *KeyBuilder) RoutineGSI1SK(startDate string) string {
	return fmt.Sprintf("ROUTINE#%s", startDate)
}

func (kb *KeyBuilder) TrainerGSI2PK(trainerID string) string {
	return fmt.Sprintf("TRAINER#%s", trainerID)
}

func (kb *KeyBuilder) RoutineGSI2SK(startDate string) string {
	return fmt.Sprintf("ROUTINE#%s", startDate)
}

func (kb *KeyBuilder) StudentPK(studentID string) string {
	return fmt.Sprintf("STUDENT#%s", studentID)
}

func (kb *KeyBuilder) WorkoutLogSK(timestamp int64, exerciseID string) string {
	return fmt.Sprintf("WORKOUT#%d#%s", timestamp, exerciseID)
}

func (kb *KeyBuilder) DailySummarySK(date string) string {
	return fmt.Sprintf("DAILY_SUMMARY#%s", date)
}

func (kb *KeyBuilder) RoutineGSI1PK(routineID string) string {
	return fmt.Sprintf("ROUTINE#%s", routineID)
}

func (kb *KeyBuilder) WorkoutGSI1SK(timestamp int64) string {
	return fmt.Sprintf("WORKOUT#%d", timestamp)
}

func (kb *KeyBuilder) StudentDateGSI2PK(studentID, date string) string {
	return fmt.Sprintf("STUDENT#%s#DATE#%s", studentID, date)
}

func (kb *KeyBuilder) WorkoutGSI2SK(timestamp int64) string {
	return fmt.Sprintf("WORKOUT#%d", timestamp)
}

func (kb *KeyBuilder) SummaryGSI1SK(date string) string {
	return fmt.Sprintf("SUMMARY#%s", date)
}
