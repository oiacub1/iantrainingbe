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

func (r *RoutineRepository) CreateWorkoutDay(ctx context.Context, workoutDay *routine.WorkoutDay) error {
	item, err := attributevalue.MarshalMap(workoutDayToDynamoItem(workoutDay))
	if err != nil {
		return fmt.Errorf("failed to marshal workout day: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create workout day: %w", err)
	}

	return nil
}

func (r *RoutineRepository) GetWorkoutDays(ctx context.Context, routineID string) ([]*routine.WorkoutDay, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :routineId AND begins_with(SK, :workoutDayPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":routineId":        &types.AttributeValueMemberS{Value: r.keyBuilder.RoutinePK(routineID)},
			":workoutDayPrefix": &types.AttributeValueMemberS{Value: "WORKOUTDAY#"},
		},
	}

	resp, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query workout days: %w", err)
	}

	workoutDays := make([]*routine.WorkoutDay, 0, len(resp.Items))
	for _, item := range resp.Items {
		var dynamoItem DynamoWorkoutDay
		if err := attributevalue.UnmarshalMap(item, &dynamoItem); err != nil {
			continue
		}
		workoutDays = append(workoutDays, dynamoItemToWorkoutDay(&dynamoItem))
	}

	return workoutDays, nil
}

func (r *RoutineRepository) UpdateWorkoutDay(ctx context.Context, workoutDay *routine.WorkoutDay) error {
	item, err := attributevalue.MarshalMap(workoutDayToDynamoItem(workoutDay))
	if err != nil {
		return fmt.Errorf("failed to marshal workout day: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update workout day: %w", err)
	}

	return nil
}

func (r *RoutineRepository) DeleteWorkoutDays(ctx context.Context, routineID string) error {
	workoutDays, err := r.GetWorkoutDays(ctx, routineID)
	if err != nil {
		return fmt.Errorf("failed to get workout days for deletion: %w", err)
	}

	for _, workoutDay := range workoutDays {
		_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(r.tableName),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: r.keyBuilder.RoutinePK(routineID)},
				"SK": &types.AttributeValueMemberS{Value: r.keyBuilder.WorkoutDaySK(workoutDay.WeekNumber, workoutDay.DayNumber)},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete workout day: %w", err)
		}
	}

	return nil
}

func (r *RoutineRepository) CreateWorkoutLog(ctx context.Context, log *routine.WorkoutLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}

	item, err := attributevalue.MarshalMap(workoutLogToDynamoItem(log))
	if err != nil {
		return fmt.Errorf("failed to marshal workout log: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create workout log: %w", err)
	}

	return nil
}

func (r *RoutineRepository) GetWorkoutLogs(ctx context.Context, studentID string, startDate, endDate time.Time) ([]*routine.WorkoutLog, error) {
	startTimestamp := startDate.Unix()
	endTimestamp := endDate.Unix()

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :studentId AND SK BETWEEN :startTimestamp AND :endTimestamp"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":studentId":      &types.AttributeValueMemberS{Value: r.keyBuilder.StudentPK(studentID)},
			":startTimestamp": &types.AttributeValueMemberS{Value: fmt.Sprintf("WORKOUT#%d", startTimestamp)},
			":endTimestamp":   &types.AttributeValueMemberS{Value: fmt.Sprintf("WORKOUT#%d", endTimestamp)},
		},
	}

	resp, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query workout logs: %w", err)
	}

	workoutLogs := make([]*routine.WorkoutLog, 0, len(resp.Items))
	for _, item := range resp.Items {
		var dynamoItem DynamoWorkoutLog
		if err := attributevalue.UnmarshalMap(item, &dynamoItem); err != nil {
			continue
		}
		workoutLogs = append(workoutLogs, dynamoItemToWorkoutLog(&dynamoItem))
	}

	return workoutLogs, nil
}

func (r *RoutineRepository) GetWorkoutLogsByRoutine(ctx context.Context, routineID string) ([]*routine.WorkoutLog, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :routineId AND begins_with(GSI1SK, :workoutPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":routineId":     &types.AttributeValueMemberS{Value: r.keyBuilder.RoutineGSI1PK(routineID)},
			":workoutPrefix": &types.AttributeValueMemberS{Value: "WORKOUT#"},
		},
	}

	resp, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query workout logs by routine: %w", err)
	}

	workoutLogs := make([]*routine.WorkoutLog, 0, len(resp.Items))
	for _, item := range resp.Items {
		var dynamoItem DynamoWorkoutLog
		if err := attributevalue.UnmarshalMap(item, &dynamoItem); err != nil {
			continue
		}
		workoutLogs = append(workoutLogs, dynamoItemToWorkoutLog(&dynamoItem))
	}

	return workoutLogs, nil
}

func (r *RoutineRepository) UpdateWorkoutLog(ctx context.Context, log *routine.WorkoutLog) error {
	item, err := attributevalue.MarshalMap(workoutLogToDynamoItem(log))
	if err != nil {
		return fmt.Errorf("failed to marshal workout log: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update workout log: %w", err)
	}

	return nil
}

func (r *RoutineRepository) CreateDailySummary(ctx context.Context, summary *routine.DailySummary) error {
	item, err := attributevalue.MarshalMap(dailySummaryToDynamoItem(summary))
	if err != nil {
		return fmt.Errorf("failed to marshal daily summary: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create daily summary: %w", err)
	}

	return nil
}

func (r *RoutineRepository) GetDailySummaries(ctx context.Context, studentID string, startDate, endDate time.Time) ([]*routine.DailySummary, error) {
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :studentId AND SK BETWEEN :startDate AND :endDate"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":studentId": &types.AttributeValueMemberS{Value: r.keyBuilder.StudentPK(studentID)},
			":startDate": &types.AttributeValueMemberS{Value: r.keyBuilder.DailySummarySK(startDateStr)},
			":endDate":   &types.AttributeValueMemberS{Value: r.keyBuilder.DailySummarySK(endDateStr)},
		},
	}

	resp, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily summaries: %w", err)
	}

	summaries := make([]*routine.DailySummary, 0, len(resp.Items))
	for _, item := range resp.Items {
		var dynamoItem DynamoDailySummary
		if err := attributevalue.UnmarshalMap(item, &dynamoItem); err != nil {
			continue
		}
		summaries = append(summaries, dynamoItemToDailySummary(&dynamoItem))
	}

	return summaries, nil
}

func (r *RoutineRepository) GetDailySummary(ctx context.Context, studentID string, date string) (*routine.DailySummary, error) {
	resp, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.keyBuilder.StudentPK(studentID)},
			"SK": &types.AttributeValueMemberS{Value: r.keyBuilder.DailySummarySK(date)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get daily summary: %w", err)
	}

	if len(resp.Item) == 0 {
		return nil, routine.ErrWorkoutLogNotFound
	}

	var dynamoItem DynamoDailySummary
	if err := attributevalue.UnmarshalMap(resp.Item, &dynamoItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal daily summary: %w", err)
	}

	return dynamoItemToDailySummary(&dynamoItem), nil
}
