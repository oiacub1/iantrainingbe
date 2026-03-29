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

	"iantraining/internal/domain/user"
)

type UserRepository struct {
	client     *dynamodb.Client
	tableName  string
	keyBuilder *KeyBuilder
}

func NewUserRepository(client *dynamodb.Client, tableName string) *UserRepository {
	return &UserRepository{
		client:     client,
		tableName:  tableName,
		keyBuilder: NewKeyBuilder(),
	}
}

type userItem struct {
	PK         string                 `dynamodbav:"PK"`
	SK         string                 `dynamodbav:"SK"`
	EntityType string                 `dynamodbav:"entityType"`
	ID         string                 `dynamodbav:"id"`
	Email      string                 `dynamodbav:"email"`
	Name       string                 `dynamodbav:"name"`
	Phone      string                 `dynamodbav:"phone"`
	Role       string                 `dynamodbav:"role"`
	Status     string                 `dynamodbav:"status"`
	TrainerID  string                 `dynamodbav:"trainerId,omitempty"`
	Metadata   map[string]interface{} `dynamodbav:"metadata,omitempty"`
	CreatedAt  int64                  `dynamodbav:"createdAt"`
	UpdatedAt  int64                  `dynamodbav:"updatedAt"`
}

type trainerStudentItem struct {
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	EntityType   string `dynamodbav:"entityType"`
	StudentName  string `dynamodbav:"studentName"`
	StudentEmail string `dynamodbav:"studentEmail"`
	AssignedAt   int64  `dynamodbav:"assignedAt"`
	Status       string `dynamodbav:"status"`
	GSI1PK       string `dynamodbav:"GSI1PK"`
	GSI1SK       string `dynamodbav:"GSI1SK"`
}

func (r *UserRepository) CreateUser(ctx context.Context, u *user.User) error {
	now := time.Now().Unix()

	item := userItem{
		PK:         r.keyBuilder.UserPK(u.ID),
		SK:         r.keyBuilder.ProfileSK(),
		EntityType: EntityTypeUser,
		ID:         u.ID,
		Email:      u.Email,
		Name:       u.Name,
		Phone:      u.Phone,
		Role:       string(u.Role),
		Status:     string(u.Status),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put user: %w", err)
	}

	u.CreatedAt = time.Unix(now, 0)
	u.UpdatedAt = time.Unix(now, 0)

	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*user.User, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.keyBuilder.UserPK(userID)},
			"SK": &types.AttributeValueMemberS{Value: r.keyBuilder.ProfileSK()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if result.Item == nil {
		return nil, user.ErrUserNotFound
	}

	var item userItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return r.itemToUser(&item), nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return nil, fmt.Errorf("not implemented: use GSI with email as key")
}

func (r *UserRepository) UpdateUser(ctx context.Context, u *user.User) error {
	now := time.Now().Unix()

	item := userItem{
		PK:         r.keyBuilder.UserPK(u.ID),
		SK:         r.keyBuilder.ProfileSK(),
		EntityType: EntityTypeUser,
		ID:         u.ID,
		Email:      u.Email,
		Name:       u.Name,
		Phone:      u.Phone,
		Role:       string(u.Role),
		Status:     string(u.Status),
		CreatedAt:  u.CreatedAt.Unix(),
		UpdatedAt:  now,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	u.UpdatedAt = time.Unix(now, 0)

	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.keyBuilder.UserPK(userID)},
			"SK": &types.AttributeValueMemberS{Value: r.keyBuilder.ProfileSK()},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *UserRepository) CreateTrainer(ctx context.Context, trainer *user.Trainer) error {
	now := time.Now().Unix()

	metadata := map[string]interface{}{
		"specializations": trainer.Metadata.Specializations,
		"certifications":  trainer.Metadata.Certifications,
		"bio":             trainer.Metadata.Bio,
		"yearsExperience": trainer.Metadata.YearsExperience,
	}

	item := userItem{
		PK:         r.keyBuilder.UserPK(trainer.ID),
		SK:         r.keyBuilder.ProfileSK(),
		EntityType: EntityTypeTrainer,
		ID:         trainer.ID,
		Email:      trainer.Email,
		Name:       trainer.Name,
		Phone:      trainer.Phone,
		Role:       string(RoleTrainer),
		Status:     string(trainer.Status),
		Metadata:   metadata,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal trainer: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put trainer: %w", err)
	}

	trainer.CreatedAt = time.Unix(now, 0)
	trainer.UpdatedAt = time.Unix(now, 0)

	return nil
}

func (r *UserRepository) GetTrainer(ctx context.Context, trainerID string) (*user.Trainer, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.keyBuilder.UserPK(trainerID)},
			"SK": &types.AttributeValueMemberS{Value: r.keyBuilder.ProfileSK()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get trainer: %w", err)
	}

	if result.Item == nil {
		return nil, user.ErrTrainerNotFound
	}

	var item userItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trainer: %w", err)
	}

	if item.EntityType != EntityTypeTrainer && item.Role != string(RoleTrainer) {
		return nil, user.ErrTrainerNotFound
	}

	return r.itemToTrainer(&item), nil
}

func (r *UserRepository) CreateStudent(ctx context.Context, student *user.Student) error {
	now := time.Now().Unix()

	metadata := map[string]interface{}{
		"goals":        student.Metadata.Goals,
		"injuries":     student.Metadata.Injuries,
		"fitnessLevel": student.Metadata.FitnessLevel,
		"weight":       student.Metadata.Weight,
		"height":       student.Metadata.Height,
		"age":          student.Metadata.Age,
	}

	item := userItem{
		PK:         r.keyBuilder.UserPK(student.ID),
		SK:         r.keyBuilder.ProfileSK(),
		EntityType: EntityTypeStudent,
		ID:         student.ID,
		Email:      student.Email,
		Name:       student.Name,
		Phone:      student.Phone,
		Role:       string(RoleStudent),
		Status:     string(student.Status),
		TrainerID:  student.TrainerID,
		Metadata:   metadata,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal student: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put student: %w", err)
	}

	student.CreatedAt = time.Unix(now, 0)
	student.UpdatedAt = time.Unix(now, 0)

	return nil
}

func (r *UserRepository) GetStudent(ctx context.Context, studentID string) (*user.Student, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.keyBuilder.UserPK(studentID)},
			"SK": &types.AttributeValueMemberS{Value: r.keyBuilder.ProfileSK()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get student: %w", err)
	}

	if result.Item == nil {
		return nil, user.ErrStudentNotFound
	}

	var item userItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal student: %w", err)
	}

	if item.EntityType != EntityTypeStudent && item.Role != string(RoleStudent) {
		return nil, user.ErrStudentNotFound
	}

	return r.itemToStudent(&item), nil
}

func (r *UserRepository) UpdateStudent(ctx context.Context, student *user.Student) error {
	now := time.Now().Unix()

	metadata := map[string]interface{}{
		"goals":        student.Metadata.Goals,
		"injuries":     student.Metadata.Injuries,
		"fitnessLevel": student.Metadata.FitnessLevel,
		"weight":       student.Metadata.Weight,
		"height":       student.Metadata.Height,
		"age":          student.Metadata.Age,
	}

	item := userItem{
		PK:         r.keyBuilder.UserPK(student.ID),
		SK:         r.keyBuilder.ProfileSK(),
		EntityType: EntityTypeStudent,
		ID:         student.ID,
		Email:      student.Email,
		Name:       student.Name,
		Phone:      student.Phone,
		Role:       string(RoleStudent),
		Status:     string(student.Status),
		TrainerID:  student.TrainerID,
		Metadata:   metadata,
		CreatedAt:  student.CreatedAt.Unix(),
		UpdatedAt:  now,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal student: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to update student: %w", err)
	}

	student.UpdatedAt = time.Unix(now, 0)

	return nil
}

func (r *UserRepository) AssignStudentToTrainer(ctx context.Context, relation *user.TrainerStudent) error {
	now := time.Now().Unix()

	item := trainerStudentItem{
		PK:           r.keyBuilder.UserPK(relation.TrainerID),
		SK:           r.keyBuilder.StudentSK(relation.StudentID),
		EntityType:   EntityTypeTrainerStudent,
		StudentName:  relation.StudentName,
		StudentEmail: relation.StudentEmail,
		AssignedAt:   now,
		Status:       string(relation.Status),
		GSI1PK:       r.keyBuilder.UserPK(relation.StudentID),
		GSI1SK:       r.keyBuilder.TrainerSK(relation.TrainerID),
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal trainer-student relation: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put trainer-student relation: %w", err)
	}

	relation.AssignedAt = time.Unix(now, 0)

	return nil
}

func (r *UserRepository) ListStudentsByTrainer(ctx context.Context, trainerID string, limit int, lastKey string) ([]*user.TrainerStudent, string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: r.keyBuilder.UserPK(trainerID)},
			":sk": &types.AttributeValueMemberS{Value: "STUDENT#"},
		},
		Limit: aws.Int32(int32(limit)),
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
		return nil, "", fmt.Errorf("failed to query students: %w", err)
	}

	students := make([]*user.TrainerStudent, 0, len(result.Items))
	for _, item := range result.Items {
		var tsItem trainerStudentItem
		if err := attributevalue.UnmarshalMap(item, &tsItem); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal trainer-student: %w", err)
		}
		students = append(students, r.itemToTrainerStudent(&tsItem, trainerID))
	}

	var nextKey string
	if result.LastEvaluatedKey != nil {
		keyBytes, err := json.Marshal(result.LastEvaluatedKey)
		if err != nil {
			return nil, "", fmt.Errorf("failed to marshal last key: %w", err)
		}
		nextKey = string(keyBytes)
	}

	return students, nextKey, nil
}

func (r *UserRepository) GetTrainerByStudent(ctx context.Context, studentID string) (*user.Trainer, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk AND begins_with(GSI1SK, :gsi1sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: r.keyBuilder.UserPK(studentID)},
			":gsi1sk": &types.AttributeValueMemberS{Value: "TRAINER#"},
		},
		Limit: aws.Int32(1),
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query trainer: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, user.ErrTrainerNotFound
	}

	var tsItem trainerStudentItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &tsItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trainer-student: %w", err)
	}

	trainerID := tsItem.PK[5:]
	return r.GetTrainer(ctx, trainerID)
}

func (r *UserRepository) RemoveStudentFromTrainer(ctx context.Context, trainerID, studentID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.keyBuilder.UserPK(trainerID)},
			"SK": &types.AttributeValueMemberS{Value: r.keyBuilder.StudentSK(studentID)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to remove student from trainer: %w", err)
	}

	return nil
}

func (r *UserRepository) itemToUser(item *userItem) *user.User {
	return &user.User{
		ID:        item.ID,
		Email:     item.Email,
		Name:      item.Name,
		Phone:     item.Phone,
		Role:      user.UserRole(item.Role),
		Status:    user.UserStatus(item.Status),
		CreatedAt: time.Unix(item.CreatedAt, 0),
		UpdatedAt: time.Unix(item.UpdatedAt, 0),
	}
}

func (r *UserRepository) itemToTrainer(item *userItem) *user.Trainer {
	trainer := &user.Trainer{
		User: user.User{
			ID:        item.ID,
			Email:     item.Email,
			Name:      item.Name,
			Phone:     item.Phone,
			Role:      user.RoleTrainer,
			Status:    user.UserStatus(item.Status),
			CreatedAt: time.Unix(item.CreatedAt, 0),
			UpdatedAt: time.Unix(item.UpdatedAt, 0),
		},
	}

	if item.Metadata != nil {
		if specs, ok := item.Metadata["specializations"].([]interface{}); ok {
			for _, s := range specs {
				if str, ok := s.(string); ok {
					trainer.Metadata.Specializations = append(trainer.Metadata.Specializations, str)
				}
			}
		}
		if certs, ok := item.Metadata["certifications"].([]interface{}); ok {
			for _, c := range certs {
				if str, ok := c.(string); ok {
					trainer.Metadata.Certifications = append(trainer.Metadata.Certifications, str)
				}
			}
		}
		if bio, ok := item.Metadata["bio"].(string); ok {
			trainer.Metadata.Bio = bio
		}
		if years, ok := item.Metadata["yearsExperience"].(float64); ok {
			trainer.Metadata.YearsExperience = int(years)
		}
	}

	return trainer
}

func (r *UserRepository) itemToStudent(item *userItem) *user.Student {
	student := &user.Student{
		User: user.User{
			ID:        item.ID,
			Email:     item.Email,
			Name:      item.Name,
			Phone:     item.Phone,
			Role:      user.RoleStudent,
			Status:    user.UserStatus(item.Status),
			CreatedAt: time.Unix(item.CreatedAt, 0),
			UpdatedAt: time.Unix(item.UpdatedAt, 0),
		},
		TrainerID: item.TrainerID,
	}

	if item.Metadata != nil {
		if goals, ok := item.Metadata["goals"].([]interface{}); ok {
			for _, g := range goals {
				if str, ok := g.(string); ok {
					student.Metadata.Goals = append(student.Metadata.Goals, str)
				}
			}
		}
		if injuries, ok := item.Metadata["injuries"].([]interface{}); ok {
			for _, i := range injuries {
				if str, ok := i.(string); ok {
					student.Metadata.Injuries = append(student.Metadata.Injuries, str)
				}
			}
		}
		if level, ok := item.Metadata["fitnessLevel"].(string); ok {
			student.Metadata.FitnessLevel = level
		}
		if weight, ok := item.Metadata["weight"].(float64); ok {
			student.Metadata.Weight = weight
		}
		if height, ok := item.Metadata["height"].(float64); ok {
			student.Metadata.Height = height
		}
		if age, ok := item.Metadata["age"].(float64); ok {
			student.Metadata.Age = int(age)
		}
	}

	return student
}

func (r *UserRepository) itemToTrainerStudent(item *trainerStudentItem, trainerID string) *user.TrainerStudent {
	studentID := item.SK[8:]

	return &user.TrainerStudent{
		TrainerID:    trainerID,
		StudentID:    studentID,
		StudentName:  item.StudentName,
		StudentEmail: item.StudentEmail,
		AssignedAt:   time.Unix(item.AssignedAt, 0),
		Status:       user.UserStatus(item.Status),
	}
}

const (
	RoleTrainer = "TRAINER"
	RoleStudent = "STUDENT"
)
