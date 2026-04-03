package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	userDomain "iantraining/internal/domain/user"
	dynamodbRepo "iantraining/internal/repository/dynamodb"
	authService "iantraining/internal/service/auth"
)

func main() {
	ctx := context.Background()

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	// Create DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)
	tableName := "training-platform"

	// Initialize repositories
	userRepo := dynamodbRepo.NewUserRepository(dynamoClient, tableName)
	authRepo := dynamodbRepo.NewAuthRepository(dynamoClient, tableName)

	// Initialize auth service for password hashing
	jwtService := authService.NewJWTService(
		"dummy-secret",
		time.Hour,
		time.Hour*24,
		"iantraining",
	)
	authService := authService.NewService(authRepo, userRepo, jwtService)

	// Create demo users
	users := []struct {
		email    string
		password string
		name     string
		role     userDomain.UserRole
	}{
		{
			email:    "trainer@example.com",
			password: "password123",
			name:     "Demo Trainer",
			role:     userDomain.RoleTrainer,
		},
		{
			email:    "student@example.com",
			password: "password123",
			name:     "Demo Student",
			role:     userDomain.RoleStudent,
		},
	}

	fmt.Println("Creating demo users...")

	for _, user := range users {
		// Create user
		userEntity := &userDomain.User{
			ID:     fmt.Sprintf("user-%d", time.Now().UnixNano()),
			Name:   user.name,
			Email:  user.email,
			Role:   user.role,
			Phone:  "+1234567890",
			Status: userDomain.StatusActive,
		}

		switch user.role {
		case userDomain.RoleTrainer:
			trainer := &userDomain.Trainer{
				User: *userEntity,
				Metadata: userDomain.TrainerMetadata{
					Specializations: []string{"Strength", "Cardio"},
					Certifications:  []string{"NASM-CPT"},
					Bio:             "Experienced personal trainer",
					YearsExperience: 5,
				},
			}
			if err := userRepo.CreateTrainer(ctx, trainer); err != nil {
				log.Printf("failed to create trainer %s: %v", user.email, err)
				continue
			}
		case userDomain.RoleStudent:
			student := &userDomain.Student{
				User:      *userEntity,
				TrainerID: "user-trainer-1", // Assign to demo trainer
				Metadata: userDomain.StudentMetadata{
					Goals:        []string{"Weight Loss", "Muscle Gain"},
					Injuries:     []string{},
					FitnessLevel: "BEGINNER",
					Weight:       70,
					Height:       175,
					Age:          30,
				},
			}
			if err := userRepo.CreateStudent(ctx, student); err != nil {
				log.Printf("failed to create student %s: %v", user.email, err)
				continue
			}
		}

		// Create credentials
		if err := authService.CreateCredentials(ctx, userEntity.ID, user.email, user.password); err != nil {
			log.Printf("failed to create credentials for %s: %v", user.email, err)
			continue
		}

		fmt.Printf("✅ Created user: %s (%s)\n", user.email, user.role)
	}

	fmt.Println("\nDemo users created successfully!")
	fmt.Println("\nLogin credentials:")
	fmt.Println("Trainer: trainer@example.com / password123")
	fmt.Println("Student: student@example.com / password123")
}
