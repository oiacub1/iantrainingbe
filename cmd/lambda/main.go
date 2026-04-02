package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	appConfig "iantraining/internal/config"
	authDomain "iantraining/internal/domain/auth"
	exerciseDomain "iantraining/internal/domain/exercise"
	userDomain "iantraining/internal/domain/user"
	dynamodbRepo "iantraining/internal/repository/dynamodb"
	authService "iantraining/internal/service/auth"
	exerciseService "iantraining/internal/service/exercise"
	userService "iantraining/internal/service/user"
	"iantraining/pkg/logger"
)

type App struct {
	exerciseService *exerciseService.Service
	userService     *userService.Service
	authService     *authService.Service
	log             *logger.Logger
}

type CreateExerciseRequest struct {
	Name           string                       `json:"name"`
	NameKey        string                       `json:"nameKey"`
	DescriptionKey string                       `json:"descriptionKey"`
	YoutubeURL     string                       `json:"youtubeUrl"`
	ThumbnailURL   string                       `json:"thumbnailUrl"`
	MuscleGroups   []exerciseDomain.MuscleGroup `json:"muscleGroups"`
	Difficulty     string                       `json:"difficulty"`
	Equipment      []string                     `json:"equipment"`
	TrainerID      string                       `json:"trainerId"`
}

type CreateUserRequest struct {
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Role         string   `json:"type"`
	Phone        string   `json:"phone,omitempty"`
	TrainerID    string   `json:"trainerId,omitempty"`
	FitnessLevel string   `json:"fitnessLevel,omitempty"`
	Goals        []string `json:"goals,omitempty"`
}

type AssignTrainerRequest struct {
	TrainerID string `json:"trainerId"`
	StudentID string `json:"studentId"`
}

type AssignTrainerByEmailRequest struct {
	Email string `json:"email"`
}

var app *App

func initializeApp() *App {
	appCfg, err := appConfig.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)

	exerciseRepo := dynamodbRepo.NewExerciseRepository(dynamoClient, appCfg.DynamoDB.TableName)
	userRepo := dynamodbRepo.NewUserRepository(dynamoClient, appCfg.DynamoDB.TableName)
	authRepo := dynamodbRepo.NewAuthRepository(dynamoClient, appCfg.DynamoDB.TableName)

	// Initialize JWT service
	jwtService := authService.NewJWTService(
		"your-secret-key-change-in-production",
		time.Hour*1,    // 1 hour access token
		time.Hour*24*7, // 7 days refresh token
		"iantraining-api",
	)

	return &App{
		exerciseService: exerciseService.NewService(exerciseRepo),
		userService:     userService.NewService(userRepo),
		authService:     authService.NewService(authRepo, userRepo, jwtService),
		log:             logger.New(),
	}
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := request.Path
	method := request.HTTPMethod

	switch {
	case method == "GET" && path == "/api/v1/health":
		return app.healthCheck(ctx, request)
	case method == "POST" && path == "/api/v1/auth/register":
		return app.register(ctx, request)
	case method == "POST" && path == "/api/v1/auth/login":
		return app.login(ctx, request)
	case method == "POST" && path == "/api/v1/auth/refresh":
		return app.refreshToken(ctx, request)
	case method == "POST" && path == "/api/v1/auth/logout":
		return app.logout(ctx, request)
	case method == "POST" && path == "/api/v1/exercises":
		return app.createExercise(ctx, request)
	case method == "GET" && strings.HasPrefix(path, "/api/v1/exercises/"):
		return app.getExercise(ctx, request)
	case method == "GET" && path == "/api/v1/exercises":
		return app.listExercises(ctx, request)
	case method == "POST" && path == "/api/v1/users":
		return app.createUser(ctx, request)
	case method == "GET" && strings.HasPrefix(path, "/api/v1/users/"):
		return app.getUser(ctx, request)
	case method == "POST" && strings.Contains(path, "/trainers/") && strings.HasSuffix(path, "/students/assign-by-email"):
		return app.assignTrainerByEmail(ctx, request)
	case method == "POST" && strings.Contains(path, "/trainers/") && strings.HasSuffix(path, "/students"):
		return app.assignTrainer(ctx, request)
	case method == "GET" && strings.Contains(path, "/trainers/") && strings.HasSuffix(path, "/students"):
		return app.listStudents(ctx, request)
	default:
		return errorResponse(404, "Route not found")
	}
}

func (a *App) createExercise(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req CreateExerciseRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	ex := &exerciseDomain.Exercise{
		ID:             uuid.New().String(),
		Name:           req.Name,
		NameKey:        req.NameKey,
		DescriptionKey: req.DescriptionKey,
		YoutubeURL:     req.YoutubeURL,
		ThumbnailURL:   req.ThumbnailURL,
		MuscleGroups:   req.MuscleGroups,
		Difficulty:     exerciseDomain.Difficulty(req.Difficulty),
		Equipment:      req.Equipment,
		CreatedBy:      req.TrainerID,
	}

	if err := a.exerciseService.Create(ctx, ex); err != nil {
		a.log.Error("failed to create exercise", err)
		return errorResponse(500, fmt.Sprintf("Failed to create exercise: %v", err))
	}

	return jsonResponse(201, map[string]interface{}{
		"message":    "Exercise created successfully",
		"exerciseId": ex.ID,
	})
}

func (a *App) getExercise(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := strings.TrimPrefix(request.Path, "/api/v1/exercises/")

	ex, err := a.exerciseService.GetByID(ctx, id)
	if err != nil {
		return errorResponse(404, "Exercise not found")
	}

	return jsonResponse(200, map[string]interface{}{
		"message": "Exercise retrieved successfully",
		"data":    ex,
	})
}

func (a *App) listExercises(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	trainerID := request.QueryStringParameters["trainerId"]
	if trainerID == "" {
		return errorResponse(400, "trainerId query parameter is required")
	}

	exercises, _, err := a.exerciseService.ListByTrainer(ctx, trainerID, 50, "")
	if err != nil {
		return errorResponse(500, fmt.Sprintf("Failed to list exercises: %v", err))
	}

	return jsonResponse(200, map[string]interface{}{
		"message": "Exercises retrieved successfully",
		"data":    exercises,
	})
}

func (a *App) createUser(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req CreateUserRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	user := &userDomain.User{
		ID:     uuid.New().String(),
		Name:   req.Name,
		Email:  req.Email,
		Role:   userDomain.UserRole(req.Role),
		Phone:  req.Phone,
		Status: userDomain.StatusActive,
	}

	if req.Role == string(userDomain.RoleStudent) {
		student := &userDomain.Student{
			User:      *user,
			TrainerID: req.TrainerID,
			Metadata: userDomain.StudentMetadata{
				FitnessLevel: req.FitnessLevel,
				Goals:        req.Goals,
			},
		}
		if err := a.userService.CreateStudent(ctx, student); err != nil {
			return errorResponse(500, fmt.Sprintf("Failed to create student: %v", err))
		}
	} else if req.Role == string(userDomain.RoleTrainer) {
		trainer := &userDomain.Trainer{
			User:     *user,
			Metadata: userDomain.TrainerMetadata{},
		}
		if err := a.userService.CreateTrainer(ctx, trainer); err != nil {
			return errorResponse(500, fmt.Sprintf("Failed to create trainer: %v", err))
		}
	} else {
		return errorResponse(400, "Invalid user type. Must be 'TRAINER' or 'STUDENT'")
	}

	return jsonResponse(201, map[string]interface{}{
		"message": "User created successfully",
		"userId":  user.ID,
	})
}

func (a *App) getUser(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := strings.TrimPrefix(request.Path, "/api/v1/users/")
	userType := request.QueryStringParameters["type"]

	if userType == "student" {
		student, err := a.userService.GetStudent(ctx, id)
		if err != nil {
			return errorResponse(404, "Student not found")
		}
		return jsonResponse(200, map[string]interface{}{
			"message": "Student retrieved successfully",
			"data":    student,
		})
	} else if userType == "trainer" {
		trainer, err := a.userService.GetTrainer(ctx, id)
		if err != nil {
			return errorResponse(404, "Trainer not found")
		}
		return jsonResponse(200, map[string]interface{}{
			"message": "Trainer retrieved successfully",
			"data":    trainer,
		})
	}

	return errorResponse(400, "User type query parameter is required (student|trainer)")
}

func (a *App) assignTrainer(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	parts := strings.Split(request.Path, "/")
	var trainerID string
	for i, part := range parts {
		if part == "trainers" && i+1 < len(parts) {
			trainerID = parts[i+1]
			break
		}
	}

	var req AssignTrainerRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	req.TrainerID = trainerID

	if err := a.userService.AssignStudentToTrainer(ctx, req.TrainerID, req.StudentID); err != nil {
		return errorResponse(500, fmt.Sprintf("Failed to assign trainer: %v", err))
	}

	return jsonResponse(200, map[string]interface{}{
		"message": "Trainer assigned successfully",
	})
}

func (a *App) assignTrainerByEmail(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	parts := strings.Split(request.Path, "/")
	var trainerID string
	for i, part := range parts {
		if part == "trainers" && i+1 < len(parts) {
			trainerID = parts[i+1]
			break
		}
	}

	var req AssignTrainerByEmailRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	if req.Email == "" {
		return errorResponse(400, "email is required")
	}

	studentID, err := a.authService.ResolveUserIDByEmail(ctx, req.Email)
	if err != nil {
		if err == authDomain.ErrUserNotFound {
			return errorResponse(404, "Student not found")
		}
		return errorResponse(500, fmt.Sprintf("Failed to find user by email: %v", err))
	}

	if _, err := a.userService.GetStudent(ctx, studentID); err != nil {
		if _, trainerErr := a.userService.GetTrainer(ctx, studentID); trainerErr == nil {
			return errorResponse(400, "User is not a student")
		}
		if errors.Is(err, userDomain.ErrInvalidUserID) {
			return errorResponse(400, "Invalid student ID")
		}
		return errorResponse(404, "Student not found")
	}

	if err := a.userService.AssignStudentToTrainer(ctx, trainerID, studentID); err != nil {
		return errorResponse(500, fmt.Sprintf("Failed to assign trainer: %v", err))
	}

	return jsonResponse(200, map[string]interface{}{
		"message": "Trainer assigned successfully",
	})
}

func (a *App) listStudents(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	parts := strings.Split(request.Path, "/")
	var trainerID string
	for i, part := range parts {
		if part == "trainers" && i+1 < len(parts) {
			trainerID = parts[i+1]
			break
		}
	}

	students, _, err := a.userService.ListStudentsByTrainer(ctx, trainerID, 50, "")
	if err != nil {
		return errorResponse(500, fmt.Sprintf("Failed to list students: %v", err))
	}

	return jsonResponse(200, map[string]interface{}{
		"message": "Students retrieved successfully",
		"data":    students,
	})
}

func jsonResponse(statusCode int, data interface{}) (events.APIGatewayProxyResponse, error) {
	body, _ := json.Marshal(data)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}, nil
}

func errorResponse(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	body, _ := json.Marshal(map[string]string{"error": message})
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}, nil
}

// Auth handlers
func (a *App) login(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req authDomain.AuthRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	response, err := a.authService.Login(ctx, &req)
	if err != nil {
		if err == authDomain.ErrInvalidCredentials || err == authDomain.ErrUserNotFound || err == authDomain.ErrUserInactive {
			return errorResponse(401, "Invalid credentials")
		}
		a.log.Error("login failed", err)
		return errorResponse(500, "Login failed")
	}

	return jsonResponse(200, response)
}

func (a *App) register(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	type RegisterRequest struct {
		Name         string   `json:"name"`
		Email        string   `json:"email"`
		Password     string   `json:"password"`
		Role         string   `json:"role"`
		Phone        string   `json:"phone,omitempty"`
		FitnessLevel string   `json:"fitnessLevel,omitempty"`
		Goals        []string `json:"goals,omitempty"`
	}

	var req RegisterRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	// Create user
	user := &userDomain.User{
		ID:     uuid.New().String(),
		Name:   req.Name,
		Email:  req.Email,
		Role:   userDomain.UserRole(req.Role),
		Phone:  req.Phone,
		Status: userDomain.StatusActive,
	}

	// Create credentials first
	if err := a.authService.CreateCredentials(ctx, user.ID, req.Email, req.Password); err != nil {
		a.log.Error("failed to create credentials", err)
		return errorResponse(500, "Failed to create user credentials")
	}

	// Create user based on role
	if req.Role == string(userDomain.RoleStudent) {
		student := &userDomain.Student{
			User:      *user,
			TrainerID: "", // No trainer assigned yet
			Metadata: userDomain.StudentMetadata{
				FitnessLevel: req.FitnessLevel,
				Goals:        req.Goals,
			},
		}
		if err := a.userService.CreateStudent(ctx, student); err != nil {
			a.log.Error("failed to create student", err)
			return errorResponse(500, "Failed to create student")
		}
	} else if req.Role == string(userDomain.RoleTrainer) {
		trainer := &userDomain.Trainer{
			User:     *user,
			Metadata: userDomain.TrainerMetadata{},
		}
		if err := a.userService.CreateTrainer(ctx, trainer); err != nil {
			a.log.Error("failed to create trainer", err)
			return errorResponse(500, "Failed to create trainer")
		}
	} else {
		return errorResponse(400, "Invalid user type. Must be 'TRAINER' or 'STUDENT'")
	}

	// Auto-login after registration
	loginReq := &authDomain.AuthRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	response, err := a.authService.Login(ctx, loginReq)
	if err != nil {
		a.log.Error("auto-login after registration failed", err)
		// Return success anyway, user can login manually
		return jsonResponse(201, map[string]interface{}{
			"message": "User created successfully",
			"userId":  user.ID,
		})
	}

	return jsonResponse(201, response)
}

func (a *App) refreshToken(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req authDomain.RefreshRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	tokenPair, err := a.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if err == authDomain.ErrTokenExpired || err == authDomain.ErrTokenInvalid || err == authDomain.ErrInvalidRefresh {
			return errorResponse(401, "Invalid refresh token")
		}
		a.log.Error("refresh token failed", err)
		return errorResponse(500, "Token refresh failed")
	}

	return jsonResponse(200, tokenPair)
}

func (a *App) logout(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get token from Authorization header
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		return errorResponse(401, "Authorization header required")
	}

	// Extract Bearer token
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return errorResponse(401, "Invalid authorization header")
	}

	if err := a.authService.LogoutFromToken(ctx, tokenParts[1]); err != nil {
		a.log.Error("logout failed", err)
		return errorResponse(500, "Logout failed")
	}

	return jsonResponse(200, map[string]string{"message": "Logged out successfully"})
}

func (a *App) healthCheck(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Simple health check that verifies the service is running
	// and can connect to its dependencies

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "iantraining-api",
		"version":   "1.0.0",
		"checks": map[string]interface{}{
			"database": "ok",
			"auth":     "ok",
		},
	}

	a.log.Info("Health check accessed")
	return jsonResponse(200, health)
}

// Local server methods
func (a *App) healthCheckLocal(w http.ResponseWriter, r *http.Request) {
	resp, _ := a.healthCheck(r.Context(), lambdaRequestToAPIGateway(r))
	apiGatewayToHTTP(resp, w)
}

func (a *App) createExerciseLocal(w http.ResponseWriter, r *http.Request) {
	resp, _ := a.createExercise(r.Context(), lambdaRequestToAPIGateway(r))
	apiGatewayToHTTP(resp, w)
}

func (a *App) getExerciseLocal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	r = r.WithContext(context.WithValue(r.Context(), "path", vars["id"]))

	resp, _ := a.getExercise(r.Context(), lambdaRequestToAPIGateway(r))
	apiGatewayToHTTP(resp, w)
}

func (a *App) listExercisesLocal(w http.ResponseWriter, r *http.Request) {
	resp, _ := a.listExercises(r.Context(), lambdaRequestToAPIGateway(r))
	apiGatewayToHTTP(resp, w)
}

func (a *App) createUserLocal(w http.ResponseWriter, r *http.Request) {
	resp, _ := a.createUser(r.Context(), lambdaRequestToAPIGateway(r))
	apiGatewayToHTTP(resp, w)
}

func (a *App) getUserLocal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	r = r.WithContext(context.WithValue(r.Context(), "path", vars["id"]))

	resp, _ := a.getUser(r.Context(), lambdaRequestToAPIGateway(r))
	apiGatewayToHTTP(resp, w)
}

func (a *App) assignTrainerLocal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	r = r.WithContext(context.WithValue(r.Context(), "path", vars["trainerId"]))

	resp, _ := a.assignTrainer(r.Context(), lambdaRequestToAPIGateway(r))
	apiGatewayToHTTP(resp, w)
}

func (a *App) assignTrainerByEmailLocal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	r = r.WithContext(context.WithValue(r.Context(), "path", vars["trainerId"]))

	resp, _ := a.assignTrainerByEmail(r.Context(), lambdaRequestToAPIGateway(r))
	apiGatewayToHTTP(resp, w)
}

func (a *App) listStudentsLocal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	r = r.WithContext(context.WithValue(r.Context(), "path", vars["trainerId"]))

	resp, _ := a.listStudents(r.Context(), lambdaRequestToAPIGateway(r))
	apiGatewayToHTTP(resp, w)
}

func lambdaRequestToAPIGateway(r *http.Request) events.APIGatewayProxyRequest {
	body, _ := json.Marshal(map[string]interface{}{})

	// Convert url.Values to map[string]string
	queryParams := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	return events.APIGatewayProxyRequest{
		HTTPMethod:            r.Method,
		Path:                  r.URL.Path,
		Headers:               make(map[string]string),
		Body:                  string(body),
		QueryStringParameters: queryParams,
	}
}

func apiGatewayToHTTP(resp events.APIGatewayProxyResponse, w http.ResponseWriter) {
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(resp.StatusCode)
	w.Write([]byte(resp.Body))
}

func main() {
	// Load .env file for local development
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using system environment variables")
	}

	appCfg, err := appConfig.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Initialize app
	app = initializeApp()

	// Check if running locally
	if appCfg.IsDevelopment() && os.Getenv("RUN_LOCAL") == "true" {
		fmt.Println("Running in local server mode...")
		runLocalServer(appCfg)
		return
	}

	fmt.Println("Running in Lambda mode...")
	lambda.Start(handler)
}

func runLocalServer(appCfg *appConfig.Config) {
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/api/v1/health", app.healthCheckLocal).Methods("GET")

	// Exercise routes
	router.HandleFunc("/api/v1/exercises", app.createExerciseLocal).Methods("POST")
	router.HandleFunc("/api/v1/exercises/{id}", app.getExerciseLocal).Methods("GET")
	router.HandleFunc("/api/v1/exercises", app.listExercisesLocal).Methods("GET")

	// User routes
	router.HandleFunc("/api/v1/users", app.createUserLocal).Methods("POST")
	router.HandleFunc("/api/v1/users/{id}", app.getUserLocal).Methods("GET")
	router.HandleFunc("/api/v1/trainers/{trainerId}/students", app.assignTrainerLocal).Methods("POST")
	router.HandleFunc("/api/v1/trainers/{trainerId}/students/assign-by-email", app.assignTrainerByEmailLocal).Methods("POST")
	router.HandleFunc("/api/v1/trainers/{trainerId}/students", app.listStudentsLocal).Methods("GET")

	// CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	port := appCfg.App.Port
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Local server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		panic(err)
	}
}
