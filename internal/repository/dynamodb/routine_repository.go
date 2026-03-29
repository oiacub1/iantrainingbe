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
	"github.com/google/uuid"

	"iantraining/internal/domain/routine"
)

type RoutineRepository struct {
	client     *dynamodb.Client
	tableName  string
	keyBuilder *KeyBuilder
}

func NewRoutineRepository(client *dynamodb.Client, tableName string) *RoutineRepository {
	return &RoutineRepository{
		client:     client,
		tableName:  tableName,
		keyBuilder: NewKeyBuilder(),
	}
}

func (r *RoutineRepository) CreateRoutine(ctx context.Context, routine *routine.Routine) error {
	if routine.ID == "" {
		routine.ID = uuid.New().String()
	}

	now := time.Now()
	if routine.CreatedAt.IsZero() {
		routine.CreatedAt = now
	}
	routine.UpdatedAt = now

	item, err := attributevalue.MarshalMap(routineToDynamoItem(routine))
	if err != nil {
		return fmt.Errorf("failed to marshal routine: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create routine: %w", err)
	}

	return nil
}

func (r *RoutineRepository) GetRoutine(ctx context.Context, id string) (*routine.Routine, error) {
	resp, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.keyBuilder.RoutinePK(id)},
			"SK": &types.AttributeValueMemberS{Value: r.keyBuilder.RoutinePK(id)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get routine: %w", err)
	}

	if len(resp.Item) == 0 {
		return nil, routine.ErrRoutineNotFound
	}

	var item DynamoRoutine
	if err := attributevalue.UnmarshalMap(resp.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal routine: %w", err)
	}

	return dynamoItemToRoutine(&item), nil
}

func (r *RoutineRepository) UpdateRoutine(ctx context.Context, routine *routine.Routine) error {
	routine.UpdatedAt = time.Now()

	item, err := attributevalue.MarshalMap(routineToDynamoItem(routine))
	if err != nil {
		return fmt.Errorf("failed to marshal routine: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update routine: %w", err)
	}

	return nil
}

func (r *RoutineRepository) DeleteRoutine(ctx context.Context, id string) error {
	// El SK para una rutina es ROUTINE#<routineID>, no METADATA
	// Según routine_models.go: SK: "ROUTINE#" + r.ID
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.keyBuilder.RoutinePK(id)},
			"SK": &types.AttributeValueMemberS{Value: r.keyBuilder.RoutinePK(id)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete routine: %w", err)
	}

	return nil
}

func (r *RoutineRepository) ListRoutinesByTrainer(ctx context.Context, trainerID string, limit int, startKey string) ([]*routine.Routine, string, error) {
	var exclusiveStartKey map[string]types.AttributeValue
	if startKey != "" {
		if err := json.Unmarshal([]byte(startKey), &exclusiveStartKey); err != nil {
			return nil, "", fmt.Errorf("invalid start key: %w", err)
		}
	}

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI2"),
		KeyConditionExpression: aws.String("GSI2PK = :trainerId AND begins_with(GSI2SK, :routinePrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":trainerId":     &types.AttributeValueMemberS{Value: r.keyBuilder.TrainerGSI2PK(trainerID)},
			":routinePrefix": &types.AttributeValueMemberS{Value: "ROUTINE#"},
		},
		Limit:             aws.Int32(int32(limit)),
		ExclusiveStartKey: exclusiveStartKey,
	}

	resp, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query routines by trainer: %w", err)
	}

	routines := make([]*routine.Routine, 0, len(resp.Items))
	for _, item := range resp.Items {
		var dynamoItem DynamoRoutine
		if err := attributevalue.UnmarshalMap(item, &dynamoItem); err != nil {
			continue
		}
		routines = append(routines, dynamoItemToRoutine(&dynamoItem))
	}

	var nextStartKey string
	if len(resp.LastEvaluatedKey) > 0 {
		keyBytes, _ := json.Marshal(resp.LastEvaluatedKey)
		nextStartKey = string(keyBytes)
	}

	return routines, nextStartKey, nil
}
