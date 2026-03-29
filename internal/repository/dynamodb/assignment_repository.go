package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"iantraining/internal/domain/routine"
)

type AssignmentRepository struct {
	client     *dynamodb.Client
	tableName  string
	keyBuilder *KeyBuilder
}

func NewAssignmentRepository(client *dynamodb.Client, tableName string) *AssignmentRepository {
	return &AssignmentRepository{
		client:     client,
		tableName:  tableName,
		keyBuilder: NewKeyBuilder(),
	}
}

func (r *AssignmentRepository) CreateAssignment(ctx context.Context, assignment *routine.RoutineAssignment) error {
	if assignment.ID == "" {
		assignment.ID = uuid.New().String()
	}
	assignment.CreatedAt = time.Now()
	assignment.UpdatedAt = time.Now()

	item, err := attributevalue.MarshalMap(assignmentToDynamoItem(assignment))
	if err != nil {
		return fmt.Errorf("failed to marshal assignment: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	return nil
}

func (r *AssignmentRepository) GetAssignment(ctx context.Context, assignmentID, studentID string) (*routine.RoutineAssignment, error) {
	resp, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "STUDENT#" + studentID},
			"SK": &types.AttributeValueMemberS{Value: "ASSIGNMENT#" + assignmentID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}

	if len(resp.Item) == 0 {
		return nil, routine.ErrRoutineNotFound
	}

	var item DynamoRoutineAssignment
	if err := attributevalue.UnmarshalMap(resp.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal assignment: %w", err)
	}

	return dynamoItemToAssignment(&item), nil
}

func (r *AssignmentRepository) UpdateAssignment(ctx context.Context, assignment *routine.RoutineAssignment) error {
	assignment.UpdatedAt = time.Now()

	item, err := attributevalue.MarshalMap(assignmentToDynamoItem(assignment))
	if err != nil {
		return fmt.Errorf("failed to marshal assignment: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update assignment: %w", err)
	}

	return nil
}

func (r *AssignmentRepository) DeleteAssignment(ctx context.Context, assignmentID, studentID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "STUDENT#" + studentID},
			"SK": &types.AttributeValueMemberS{Value: "ASSIGNMENT#" + assignmentID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete assignment: %w", err)
	}

	return nil
}

// GetAssignmentsByStudent returns all assignments for a student
func (r *AssignmentRepository) GetAssignmentsByStudent(ctx context.Context, studentID string) ([]*routine.RoutineAssignment, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :studentId AND begins_with(SK, :assignmentPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":studentId":         &types.AttributeValueMemberS{Value: "STUDENT#" + studentID},
			":assignmentPrefix": &types.AttributeValueMemberS{Value: "ASSIGNMENT#"},
		},
	}

	resp, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignments by student: %w", err)
	}

	assignments := make([]*routine.RoutineAssignment, 0, len(resp.Items))
	for _, item := range resp.Items {
		var dynamoItem DynamoRoutineAssignment
		if err := attributevalue.UnmarshalMap(item, &dynamoItem); err != nil {
			return nil, fmt.Errorf("failed to unmarshal assignment: %w", err)
		}
		assignments = append(assignments, dynamoItemToAssignment(&dynamoItem))
	}

	return assignments, nil
}

// GetActiveAssignmentForStudent returns the active assignment for a student
func (r *AssignmentRepository) GetActiveAssignmentForStudent(ctx context.Context, studentID string) (*routine.RoutineAssignment, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI2"),
		KeyConditionExpression: aws.String("GSI2PK = :gsi2pk AND begins_with(GSI2SK, :assignmentPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi2pk":           &types.AttributeValueMemberS{Value: "STUDENT#" + studentID + "#STATUS#ACTIVE"},
			":assignmentPrefix": &types.AttributeValueMemberS{Value: "ASSIGNMENT#"},
		},
		Limit: aws.Int32(1),
	}

	resp, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query active assignment: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, routine.ErrRoutineNotFound
	}

	var dynamoItem DynamoRoutineAssignment
	if err := attributevalue.UnmarshalMap(resp.Items[0], &dynamoItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal assignment: %w", err)
	}

	return dynamoItemToAssignment(&dynamoItem), nil
}

// GetAssignmentsByRoutine returns all assignments for a routine
func (r *AssignmentRepository) GetAssignmentsByRoutine(ctx context.Context, routineID string) ([]*routine.RoutineAssignment, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :routineId AND begins_with(GSI1SK, :assignmentPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":routineId":        &types.AttributeValueMemberS{Value: "ROUTINE#" + routineID},
			":assignmentPrefix": &types.AttributeValueMemberS{Value: "ASSIGNMENT#"},
		},
	}

	resp, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query assignments by routine: %w", err)
	}

	assignments := make([]*routine.RoutineAssignment, 0, len(resp.Items))
	for _, item := range resp.Items {
		var dynamoItem DynamoRoutineAssignment
		if err := attributevalue.UnmarshalMap(item, &dynamoItem); err != nil {
			return nil, fmt.Errorf("failed to unmarshal assignment: %w", err)
		}
		assignments = append(assignments, dynamoItemToAssignment(&dynamoItem))
	}

	return assignments, nil
}
