package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"iantraining/internal/domain/exercise"
)

type ExerciseRepository struct {
	client    *dynamodb.Client
	tableName string
	keyBuilder *KeyBuilder
}

func NewExerciseRepository(client *dynamodb.Client, tableName string) *ExerciseRepository {
	return &ExerciseRepository{
		client:     client,
		tableName:  tableName,
		keyBuilder: NewKeyBuilder(),
	}
}

type exerciseItem struct {
	PK            string                  `dynamodbav:"PK"`
	SK            string                  `dynamodbav:"SK"`
	EntityType    string                  `dynamodbav:"entityType"`
	ID            string                  `dynamodbav:"id"`
	Name          string                  `dynamodbav:"name"`
	NameKey       string                  `dynamodbav:"nameKey"`
	DescriptionKey string                 `dynamodbav:"descriptionKey"`
	YoutubeURL    string                  `dynamodbav:"youtubeUrl"`
	ThumbnailURL  string                  `dynamodbav:"thumbnailUrl"`
	MuscleGroups  []exercise.MuscleGroup  `dynamodbav:"muscleGroups"`
	Difficulty    string                  `dynamodbav:"difficulty"`
	Equipment     []string                `dynamodbav:"equipment"`
	CreatedBy     string                  `dynamodbav:"createdBy"`
	CreatedAt     int64                   `dynamodbav:"createdAt"`
	UpdatedAt     int64                   `dynamodbav:"updatedAt"`
	GSI1PK        string                  `dynamodbav:"GSI1PK"`
	GSI1SK        string                  `dynamodbav:"GSI1SK"`
}

func (r *ExerciseRepository) Create(ctx context.Context, ex *exercise.Exercise) error {
	now := time.Now().Unix()
	
	item := exerciseItem{
		PK:             r.keyBuilder.ExercisePK(ex.ID),
		SK:             r.keyBuilder.MetadataSK(),
		EntityType:     EntityTypeExercise,
		ID:             ex.ID,
		Name:           ex.Name,
		NameKey:        ex.NameKey,
		DescriptionKey: ex.DescriptionKey,
		YoutubeURL:     ex.YoutubeURL,
		ThumbnailURL:   ex.ThumbnailURL,
		MuscleGroups:   ex.MuscleGroups,
		Difficulty:     string(ex.Difficulty),
		Equipment:      ex.Equipment,
		CreatedBy:      ex.CreatedBy,
		CreatedAt:      now,
		UpdatedAt:      now,
		GSI1PK:         r.keyBuilder.ExercisesByTrainerGSI1PK(ex.CreatedBy),
		GSI1SK:         r.keyBuilder.ExerciseGSI1SK(now),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal exercise: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put exercise: %w", err)
	}

	ex.CreatedAt = time.Unix(now, 0)
	ex.UpdatedAt = time.Unix(now, 0)

	return nil
}

func (r *ExerciseRepository) GetByID(ctx context.Context, exerciseID string) (*exercise.Exercise, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.keyBuilder.ExercisePK(exerciseID)},
			"SK": &types.AttributeValueMemberS{Value: r.keyBuilder.MetadataSK()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise: %w", err)
	}

	if result.Item == nil {
		return nil, exercise.ErrExerciseNotFound
	}

	var item exerciseItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal exercise: %w", err)
	}

	return r.itemToExercise(&item), nil
}

func (r *ExerciseRepository) ListByTrainer(ctx context.Context, trainerID string, limit int, lastKey string) ([]*exercise.Exercise, string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: r.keyBuilder.ExercisesByTrainerGSI1PK(trainerID)},
		},
		Limit:            aws.Int32(int32(limit)),
		ScanIndexForward: aws.Bool(false),
	}

	if lastKey != "" {
		var exclusiveStartKey map[string]types.AttributeValue
		if err := json.Unmarshal([]byte(lastKey), &exclusiveStartKey); err != nil {
			return nil, "", fmt.Errorf("invalid last key: %w", err)
		}
		input.ExclusiveStartKey = exclusiveStartKey
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query exercises: %w", err)
	}

	exercises := make([]*exercise.Exercise, 0, len(result.Items))
	for _, item := range result.Items {
		var exItem exerciseItem
		if err := attributevalue.UnmarshalMap(item, &exItem); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal exercise: %w", err)
		}
		exercises = append(exercises, r.itemToExercise(&exItem))
	}

	var nextKey string
	if result.LastEvaluatedKey != nil {
		keyBytes, err := json.Marshal(result.LastEvaluatedKey)
		if err != nil {
			return nil, "", fmt.Errorf("failed to marshal last key: %w", err)
		}
		nextKey = string(keyBytes)
	}

	return exercises, nextKey, nil
}

func (r *ExerciseRepository) Update(ctx context.Context, ex *exercise.Exercise) error {
	now := time.Now().Unix()
	
	item := exerciseItem{
		PK:             r.keyBuilder.ExercisePK(ex.ID),
		SK:             r.keyBuilder.MetadataSK(),
		EntityType:     EntityTypeExercise,
		ID:             ex.ID,
		Name:           ex.Name,
		NameKey:        ex.NameKey,
		DescriptionKey: ex.DescriptionKey,
		YoutubeURL:     ex.YoutubeURL,
		ThumbnailURL:   ex.ThumbnailURL,
		MuscleGroups:   ex.MuscleGroups,
		Difficulty:     string(ex.Difficulty),
		Equipment:      ex.Equipment,
		CreatedBy:      ex.CreatedBy,
		CreatedAt:      ex.CreatedAt.Unix(),
		UpdatedAt:      now,
		GSI1PK:         r.keyBuilder.ExercisesByTrainerGSI1PK(ex.CreatedBy),
		GSI1SK:         r.keyBuilder.ExerciseGSI1SK(ex.CreatedAt.Unix()),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal exercise: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to update exercise: %w", err)
	}

	ex.UpdatedAt = time.Unix(now, 0)

	return nil
}

func (r *ExerciseRepository) Delete(ctx context.Context, exerciseID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.keyBuilder.ExercisePK(exerciseID)},
			"SK": &types.AttributeValueMemberS{Value: r.keyBuilder.MetadataSK()},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete exercise: %w", err)
	}

	return nil
}

func (r *ExerciseRepository) itemToExercise(item *exerciseItem) *exercise.Exercise {
	return &exercise.Exercise{
		ID:             item.ID,
		Name:           item.Name,
		NameKey:        item.NameKey,
		DescriptionKey: item.DescriptionKey,
		YoutubeURL:     item.YoutubeURL,
		ThumbnailURL:   item.ThumbnailURL,
		MuscleGroups:   item.MuscleGroups,
		Difficulty:     exercise.Difficulty(item.Difficulty),
		Equipment:      item.Equipment,
		CreatedBy:      item.CreatedBy,
		CreatedAt:      time.Unix(item.CreatedAt, 0),
		UpdatedAt:      time.Unix(item.UpdatedAt, 0),
	}
}
