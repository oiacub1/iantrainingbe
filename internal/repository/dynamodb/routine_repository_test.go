package dynamodb

import (
	"context"
	"iantraining/internal/domain/routine"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DEPRECATED: Estos tests necesitan ser actualizados para el nuevo modelo de rutinas
// El nuevo modelo separa Routine (plantilla) de RoutineAssignment
// Los tests deben ser reescritos para reflejar esta separación
//
// TODO: Crear nuevos tests para:
// - Routine CRUD operations (sin StudentID, StartDate, EndDate)
// - RoutineAssignment CRUD operations
// - Queries de asignaciones por estudiante y por rutina

/*
import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"iantraining/internal/domain/routine"
)

/*
func setupTestDB(t *testing.T) *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:8000"}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     "dummy",
				SecretAccessKey: "dummy",
				SessionToken:    "",
			},
		}),
	)
	require.NoError(t, err)

	return dynamodb.NewFromConfig(cfg)
}

func TestDeleteRoutine(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRoutineRepository(client, "iantraining-local")
	ctx := context.Background()

	// Crear una rutina de prueba
	testRoutine := &routine.Routine{
		Name:      "Test Routine for Delete",
		StudentID: "test-student-123",
		TrainerID: "test-trainer-456",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, 28),
		Status:    routine.RoutineStatusDraft,
		WeekCount: 4,
	}

	// Crear la rutina
	err := repo.CreateRoutine(ctx, testRoutine)
	require.NoError(t, err, "Failed to create test routine")
	require.NotEmpty(t, testRoutine.ID, "Routine ID should be generated")

	// Crear algunos workout days para la rutina
	workoutDay1 := &routine.WorkoutDay{
		RoutineID:  testRoutine.ID,
		WeekNumber: 1,
		DayNumber:  1,
		DayName:    "Monday",
		IsRestDay:  false,
		Exercises: []routine.ExerciseSet{
			{
				ExerciseID:  "exercise-1",
				Order:       1,
				Sets:        3,
				Reps:        "10",
				RestSeconds: 60,
			},
		},
	}

	workoutDay2 := &routine.WorkoutDay{
		RoutineID:  testRoutine.ID,
		WeekNumber: 1,
		DayNumber:  2,
		DayName:    "Tuesday",
		IsRestDay:  false,
		Exercises: []routine.ExerciseSet{
			{
				ExerciseID:  "exercise-2",
				Order:       1,
				Sets:        4,
				Reps:        "12",
				RestSeconds: 90,
			},
		},
	}

	err = repo.CreateWorkoutDay(ctx, workoutDay1)
	require.NoError(t, err, "Failed to create workout day 1")

	err = repo.CreateWorkoutDay(ctx, workoutDay2)
	require.NoError(t, err, "Failed to create workout day 2")

	// Verificar que la rutina existe
	retrievedRoutine, err := repo.GetRoutine(ctx, testRoutine.ID)
	require.NoError(t, err, "Failed to retrieve routine before deletion")
	assert.Equal(t, testRoutine.ID, retrievedRoutine.ID)

	// Verificar que los workout days existen
	workoutDays, err := repo.GetWorkoutDays(ctx, testRoutine.ID)
	require.NoError(t, err, "Failed to get workout days before deletion")
	assert.Len(t, workoutDays, 2, "Should have 2 workout days")

	// Eliminar los workout days primero
	err = repo.DeleteWorkoutDays(ctx, testRoutine.ID)
	require.NoError(t, err, "Failed to delete workout days")

	// Verificar que los workout days fueron eliminados
	workoutDaysAfterDelete, err := repo.GetWorkoutDays(ctx, testRoutine.ID)
	require.NoError(t, err, "Failed to get workout days after deletion")
	assert.Len(t, workoutDaysAfterDelete, 0, "Should have 0 workout days after deletion")

	// Eliminar la rutina
	err = repo.DeleteRoutine(ctx, testRoutine.ID)
	require.NoError(t, err, "Failed to delete routine")

	// Verificar que la rutina fue eliminada
	_, err = repo.GetRoutine(ctx, testRoutine.ID)
	assert.ErrorIs(t, err, routine.ErrRoutineNotFound, "Routine should not be found after deletion")
}

func TestDeleteRoutineWithoutWorkoutDays(t *testing.T) {
	client := setupTestDB(t)
	repo := NewRoutineRepository(client, "iantraining-local")
	ctx := context.Background()

	// Crear una rutina de prueba sin workout days
	testRoutine := &routine.Routine{
		Name:      "Test Routine Without Workouts",
		StudentID: "test-student-789",
		TrainerID: "test-trainer-101",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, 14),
		Status:    routine.RoutineStatusDraft,
		WeekCount: 2,
	}

	// Crear la rutina
	err := repo.CreateRoutine(ctx, testRoutine)
	require.NoError(t, err, "Failed to create test routine")
	require.NotEmpty(t, testRoutine.ID, "Routine ID should be generated")

	// Verificar que la rutina existe
	retrievedRoutine, err := repo.GetRoutine(ctx, testRoutine.ID)
	require.NoError(t, err, "Failed to retrieve routine before deletion")
	assert.Equal(t, testRoutine.ID, retrievedRoutine.ID)

	// Eliminar la rutina directamente (sin workout days)
	err = repo.DeleteRoutine(ctx, testRoutine.ID)
	require.NoError(t, err, "Failed to delete routine")

	// Verificar que la rutina fue eliminada
	_, err = repo.GetRoutine(ctx, testRoutine.ID)
	assert.ErrorIs(t, err, routine.ErrRoutineNotFound, "Routine should not be found after deletion")
}
*/\n
