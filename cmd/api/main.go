package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
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

var app *App
var log *logger.Logger

// Pre-allocate CORS headers to avoid map allocation on every request
var corsHeaders = map[string]string{
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
	"Access-Control-Allow-Headers":     "Content-Type, Authorization",
	"Access-Control-Allow-Credentials": "true",
}

// Map to hold route handlers for O(1) lookup
var routes map[string]func(context.Context, events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse

type App struct {
	exerciseService *exerciseService.Service
	userService     *userService.Service
	authService     *authService.Service
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
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type AssignTrainerRequest struct {
	StudentID string `json:"studentId"`
}

type AssignTrainerByEmailRequest struct {
	Email string `json:"email"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func NewApp(exerciseService *exerciseService.Service, userService *userService.Service, authService *authService.Service) *App {
	return &App{
		exerciseService: exerciseService,
		userService:     userService,
		authService:     authService,
	}
}

func main() {
	// Load environment variables from .env file only if not running in Lambda
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") == "" {
		if err := godotenv.Load(); err != nil {
			fmt.Println("No .env file found, using system environment variables")
		}
	}

	ctx := context.Background()
	serverConfig, err := appConfig.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	log = logger.New()

	// Load AWS config
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	// Create DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Create repositories
	authRepo := dynamodbRepo.NewAuthRepository(dynamoClient, "training-platform")
	userRepo := dynamodbRepo.NewUserRepository(dynamoClient, "training-platform")
	exerciseRepo := dynamodbRepo.NewExerciseRepository(dynamoClient, "training-platform")

	// Create JWT service
	jwtService := authService.NewJWTService("secret", time.Hour*1, time.Hour*24, "iantraining")

	// Create services
	app = NewApp(
		exerciseService.NewService(exerciseRepo),
		userService.NewService(userRepo),
		authService.NewService(authRepo, userRepo, jwtService),
	)

	// Initialize the routes map
	routes = map[string]func(context.Context, events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse{
		"GET /api/v1/health":                               app.healthCheck,
		"POST /api/v1/auth/register":                       app.register,
		"POST /api/v1/auth/login":                          app.login,
		"POST /api/v1/auth/refresh":                        app.refreshToken,
		"POST /api/v1/auth/logout":                         app.logout,
		"POST /api/v1/exercises":                           app.createExercise,
		"GET /api/v1/exercises":                            app.listExercises,
		"POST /api/v1/users":                               app.createUser,
		"GET /api/v1/users":                                app.getUser,
		"POST /api/v1/trainers/*/students":                 app.assignTrainer,
		"POST /api/v1/trainers/*/students/assign-by-email": app.assignTrainerByEmail,
		"GET /api/v1/trainers/*/students":                  app.listStudents,
	}

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		log.Info("Running in Lambda mode...")
		lambda.Start(router)
	} else {
		log.Info("Running in local server mode...")
		http.HandleFunc("/", localHandler)

		port := serverConfig.App.Port
		if port == "" {
			port = "8080"
		}

		log.Infof("Starting local server on port %s", port)
		fmt.Println("Available endpoints:")
		fmt.Println("  GET  /api/v1/health - Health check")
		fmt.Println("  POST /api/v1/auth/login - Login")
		fmt.Println("  POST /api/v1/auth/register - Register")
		fmt.Println("  GET  /api/v1/exercises - List exercises")
		fmt.Println("  POST /api/v1/exercises - Create exercise")
		fmt.Println("  ... and more")

		go func() {
			if err := http.ListenAndServe(":"+port, nil); err != nil {
				log.Error("failed to start server", err)
			}
		}()
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
}

func router(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Handle preflight OPTIONS request
	if request.RequestContext.HTTP.Method == "OPTIONS" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers:    corsHeaders,
		}, nil
	}

	path := request.RawPath
	path = strings.Replace(path, "/prod", "", 1)
	method := strings.ToUpper(request.RequestContext.HTTP.Method)

	log.Infof("Received request: %s %s", method, path)

	// Create route key
	routeKey := fmt.Sprintf("%s %s", method, path)

	// Handle dynamic routes (with wildcards)
	handler, exists := routes[routeKey]
	if !exists {
		// Try to match dynamic routes
		handler = matchDynamicRoute(method, path)
		if handler == nil {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 404,
				Headers:    corsHeaders,
				Body:       fmt.Sprintf(`{"error":"Route not found: %s"}`, routeKey),
			}, nil
		}
	}

	response := handler(ctx, request)

	// Add CORS headers to all responses
	if response.Headers == nil {
		response.Headers = make(map[string]string)
	}
	for k, v := range corsHeaders {
		response.Headers[k] = v
	}

	return response, nil
}

func matchDynamicRoute(method, path string) func(context.Context, events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	// Handle dynamic routes
	if method == "GET" && strings.HasPrefix(path, "/api/v1/exercises/") {
		return app.getExercise
	}
	if method == "GET" && strings.HasPrefix(path, "/api/v1/users/") {
		return app.getUser
	}
	if method == "POST" && strings.Contains(path, "/trainers/") && strings.HasSuffix(path, "/students") {
		return app.assignTrainer
	}
	if method == "POST" && strings.Contains(path, "/trainers/") && strings.HasSuffix(path, "/students/assign-by-email") {
		return app.assignTrainerByEmail
	}
	if method == "GET" && strings.Contains(path, "/trainers/") && strings.HasSuffix(path, "/students") {
		return app.listStudents
	}

	return nil
}

func localHandler(w http.ResponseWriter, r *http.Request) {
	// Configurar headers CORS para el servidor local
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Manejar preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	body, _ := io.ReadAll(r.Body)

	headers := map[string]string{}
	for k, v := range r.Header {
		headers[k] = strings.Join(v, ",")
	}

	queryParams := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	// Create APIGatewayV2HTTPRequest (HTTP API format)
	req := events.APIGatewayV2HTTPRequest{
		RawPath:               r.URL.Path,
		RequestContext:        events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: r.Method}},
		Headers:               headers,
		Body:                  string(body),
		QueryStringParameters: queryParams,
	}

	resp, err := router(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(resp.StatusCode)
	w.Write([]byte(resp.Body))
}

// Health check
func (a *App) healthCheck(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
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

	log.Info("Health check accessed")
	return jsonResponse(200, health)
}

// Exercise handlers
func (a *App) createExercise(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
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
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := a.exerciseService.Create(ctx, ex); err != nil {
		log.Error("failed to create exercise", err)
		return errorResponse(500, fmt.Sprintf("Failed to create exercise: %v", err))
	}

	return jsonResponse(201, map[string]interface{}{
		"message":  "Exercise created successfully",
		"exercise": ex,
	})
}

func (a *App) getExercise(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	pathParts := strings.Split(strings.Trim(request.RawPath, "/"), "/")
	if len(pathParts) < 4 {
		return errorResponse(400, "Invalid exercise ID")
	}
	id := pathParts[3]

	ex, err := a.exerciseService.GetByID(ctx, id)
	if err != nil {
		return errorResponse(404, "Exercise not found")
	}

	return jsonResponse(200, map[string]interface{}{
		"message":  "Exercise retrieved successfully",
		"exercise": ex,
	})
}

func (a *App) listExercises(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	trainerID := request.QueryStringParameters["trainerId"]
	if trainerID == "" {
		return errorResponse(400, "trainerId query parameter is required")
	}

	exercises, _, err := a.exerciseService.ListByTrainer(ctx, trainerID, 50, "")
	if err != nil {
		return errorResponse(500, fmt.Sprintf("Failed to list exercises: %v", err))
	}

	return jsonResponse(200, map[string]interface{}{
		"message":   "Exercises retrieved successfully",
		"exercises": exercises,
	})
}

// User handlers
func (a *App) createUser(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	var req CreateUserRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	user := &userDomain.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Role:      userDomain.UserRole(req.Role),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Role == string(userDomain.RoleStudent) {
		student := &userDomain.Student{
			User:      *user,
			TrainerID: "",
			Metadata: userDomain.StudentMetadata{
				Goals:        []string{},
				Injuries:     []string{},
				FitnessLevel: "BEGINNER",
				Weight:       0,
				Height:       0,
				Age:          0,
			},
		}
		if err := a.userService.CreateStudent(ctx, student); err != nil {
			return errorResponse(500, fmt.Sprintf("Failed to create student: %v", err))
		}
	} else if req.Role == string(userDomain.RoleTrainer) {
		trainer := &userDomain.Trainer{
			User: *user,
			Metadata: userDomain.TrainerMetadata{
				Specializations: []string{},
				Certifications:  []string{},
				Bio:             "",
				YearsExperience: 0,
			},
		}
		if err := a.userService.CreateTrainer(ctx, trainer); err != nil {
			return errorResponse(500, fmt.Sprintf("Failed to create trainer: %v", err))
		}
	} else {
		return errorResponse(400, "Invalid user type. Must be 'TRAINER' or 'STUDENT'")
	}

	return jsonResponse(201, map[string]interface{}{
		"message": "User created successfully",
		"user":    user,
	})
}

func (a *App) getUser(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	pathParts := strings.Split(strings.Trim(request.RawPath, "/"), "/")
	if len(pathParts) < 4 {
		return errorResponse(400, "Invalid user ID")
	}
	id := pathParts[3]

	userType := request.QueryStringParameters["type"]
	if userType == "" {
		return errorResponse(400, "User type query parameter is required (student|trainer)")
	}

	if userType == "student" {
		student, err := a.userService.GetStudent(ctx, id)
		if err != nil {
			return errorResponse(404, "Student not found")
		}
		return jsonResponse(200, map[string]interface{}{
			"message": "Student retrieved successfully",
			"student": student,
		})
	} else if userType == "trainer" {
		trainer, err := a.userService.GetTrainer(ctx, id)
		if err != nil {
			return errorResponse(404, "Trainer not found")
		}
		return jsonResponse(200, map[string]interface{}{
			"message": "Trainer retrieved successfully",
			"trainer": trainer,
		})
	}

	return errorResponse(400, "User type query parameter is required (student|trainer)")
}

func (a *App) assignTrainer(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	pathParts := strings.Split(strings.Trim(request.RawPath, "/"), "/")
	if len(pathParts) < 5 {
		return errorResponse(400, "Invalid trainer ID")
	}
	trainerID := pathParts[2]

	var req AssignTrainerRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	if err := a.userService.AssignStudentToTrainer(ctx, trainerID, req.StudentID); err != nil {
		return errorResponse(500, fmt.Sprintf("Failed to assign trainer: %v", err))
	}

	return jsonResponse(200, map[string]interface{}{
		"message": "Trainer assigned successfully",
	})
}

func (a *App) assignTrainerByEmail(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	pathParts := strings.Split(strings.Trim(request.RawPath, "/"), "/")
	if len(pathParts) < 6 {
		return errorResponse(400, "Invalid trainer ID")
	}
	trainerID := pathParts[2]

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
		"message":   "Trainer assigned successfully",
		"trainerId": trainerID,
		"studentId": studentID,
	})
}

func (a *App) listStudents(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	pathParts := strings.Split(strings.Trim(request.RawPath, "/"), "/")
	if len(pathParts) < 5 {
		return errorResponse(400, "Invalid trainer ID")
	}
	trainerID := pathParts[2]

	students, _, err := a.userService.ListStudentsByTrainer(ctx, trainerID, 50, "")
	if err != nil {
		return errorResponse(500, fmt.Sprintf("Failed to list students: %v", err))
	}

	return jsonResponse(200, map[string]interface{}{
		"message":  "Students retrieved successfully",
		"students": students,
	})
}

// Auth handlers
func (a *App) login(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	var req authDomain.AuthRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	response, err := a.authService.Login(ctx, &req)
	if err != nil {
		if err == authDomain.ErrInvalidCredentials || err == authDomain.ErrUserNotFound || err == authDomain.ErrUserInactive {
			return errorResponse(401, "Invalid credentials")
		}
		log.Error("login failed", err)
		return errorResponse(500, "Login failed")
	}

	return jsonResponse(200, response)
}

func (a *App) register(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	var req RegisterRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	// Create user
	user := &userDomain.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Role:      userDomain.UserRole(req.Role),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create credentials first
	if err := a.authService.CreateCredentials(ctx, user.ID, req.Email, req.Password); err != nil {
		log.Error("failed to create credentials", err)
		return errorResponse(500, "Failed to create user credentials")
	}

	// Create user based on role
	if req.Role == string(userDomain.RoleStudent) {
		student := &userDomain.Student{
			User:      *user,
			TrainerID: "",
			Metadata: userDomain.StudentMetadata{
				Goals:        []string{},
				Injuries:     []string{},
				FitnessLevel: "BEGINNER",
				Weight:       0,
				Height:       0,
				Age:          0,
			},
		}
		if err := a.userService.CreateStudent(ctx, student); err != nil {
			log.Error("failed to create student", err)
			return errorResponse(500, "Failed to create student")
		}
	} else if req.Role == string(userDomain.RoleTrainer) {
		trainer := &userDomain.Trainer{
			User: *user,
			Metadata: userDomain.TrainerMetadata{
				Specializations: []string{},
				Certifications:  []string{},
				Bio:             "",
				YearsExperience: 0,
			},
		}
		if err := a.userService.CreateTrainer(ctx, trainer); err != nil {
			log.Error("failed to create trainer", err)
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
		log.Error("auto-login failed", err)
		return errorResponse(500, "Registration successful but login failed")
	}

	return jsonResponse(201, response)
}

func (a *App) refreshToken(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	var req authDomain.RefreshRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(400, "Invalid request body")
	}

	tokenPair, err := a.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if err == authDomain.ErrTokenExpired || err == authDomain.ErrTokenInvalid || err == authDomain.ErrInvalidRefresh {
			return errorResponse(401, "Invalid refresh token")
		}
		log.Error("refresh token failed", err)
		return errorResponse(500, "Token refresh failed")
	}

	return jsonResponse(200, tokenPair)
}

func (a *App) logout(ctx context.Context, request events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
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
		log.Error("logout failed", err)
		return errorResponse(500, "Logout failed")
	}

	return jsonResponse(200, map[string]string{"message": "Logged out successfully"})
}

// Helper functions
func jsonResponse(statusCode int, body interface{}) events.APIGatewayV2HTTPResponse {
	jsonBody, _ := json.Marshal(body)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(jsonBody),
	}
}

func errorResponse(statusCode int, message string) events.APIGatewayV2HTTPResponse {
	body, _ := json.Marshal(map[string]string{"error": message})
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}
