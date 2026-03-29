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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	appConfig "iantraining/internal/config"
	authDomain "iantraining/internal/domain/auth"
	exerciseDomain "iantraining/internal/domain/exercise"
	routineDomain "iantraining/internal/domain/routine"
	userDomain "iantraining/internal/domain/user"
	dynamodbRepo "iantraining/internal/repository/dynamodb"
	assignmentService "iantraining/internal/service/assignment"
	authService "iantraining/internal/service/auth"
	exerciseService "iantraining/internal/service/exercise"
	routineService "iantraining/internal/service/routine"
	userService "iantraining/internal/service/user"
	"iantraining/pkg/logger"
)

type App struct {
	exerciseService   *exerciseService.Service
	userService       *userService.Service
	routineService    *routineService.Service
	assignmentService *assignmentService.Service
	authService       *authService.Service
	authRepo          authDomain.Repository
	log               *logger.Logger
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

func init() {
	appCfg, err := appConfig.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	// Force local DynamoDB configuration if RUN_LOCAL is true
	if appCfg.App.RunLocal || appCfg.DynamoDB.Endpoint != "" {
		endpoint := appCfg.DynamoDB.Endpoint
		if appCfg.App.RunLocal && endpoint == "" {
			endpoint = "http://localhost:8000" // Default for RUN_LOCAL
		}

		fmt.Printf("Using DynamoDB endpoint: %s (RUN_LOCAL=%v)\n", endpoint, appCfg.App.RunLocal)

		cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{
					URL:           endpoint,
					SigningRegion: appCfg.DynamoDB.Region,
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
		})

		// For local DynamoDB, we don't need real credentials
		cfg.Credentials = aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "dummy",
				SecretAccessKey: "dummy",
				Source:          "local-dynamodb",
			}, nil
		})
	} else {
		fmt.Printf("Using AWS DynamoDB (RUN_LOCAL=%v)\n", appCfg.App.RunLocal)
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)

	exerciseRepo := dynamodbRepo.NewExerciseRepository(dynamoClient, appCfg.DynamoDB.TableName)
	userRepo := dynamodbRepo.NewUserRepository(dynamoClient, appCfg.DynamoDB.TableName)
	routineRepo := dynamodbRepo.NewRoutineRepository(dynamoClient, appCfg.DynamoDB.TableName)
	assignmentRepo := dynamodbRepo.NewAssignmentRepository(dynamoClient, appCfg.DynamoDB.TableName)
	authRepo := dynamodbRepo.NewAuthRepository(dynamoClient, appCfg.DynamoDB.TableName)

	// Initialize JWT service
	jwtService := authService.NewJWTService(
		"your-secret-key-change-in-production",
		time.Hour*1,    // 1 hour access token
		time.Hour*24*7, // 7 days refresh token
		"iantraining-api",
	)

	app = &App{
		exerciseService:   exerciseService.NewService(exerciseRepo),
		userService:       userService.NewService(userRepo),
		routineService:    routineService.NewService(routineRepo, userRepo),
		assignmentService: assignmentService.NewService(assignmentRepo, routineRepo, userRepo),
		authService:       authService.NewService(authRepo, userRepo, jwtService),
		authRepo:          authRepo,
		log:               logger.New(),
	}
}

func (a *App) createExercise(w http.ResponseWriter, r *http.Request) {
	var req CreateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
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

	if err := a.exerciseService.Create(r.Context(), ex); err != nil {
		a.log.Error("failed to create exercise", err)
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create exercise: %v", err))
		return
	}

	a.jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message":    "Exercise created successfully",
		"exerciseId": ex.ID,
	})
}

func (a *App) getExercise(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	ex, err := a.exerciseService.GetByID(r.Context(), id)
	if err != nil {
		a.errorResponse(w, http.StatusNotFound, "Exercise not found")
		return
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Exercise retrieved successfully",
		"data":    ex,
	})
}

func (a *App) listExercises(w http.ResponseWriter, r *http.Request) {
	trainerID := r.URL.Query().Get("trainerId")
	if trainerID == "" {
		a.errorResponse(w, http.StatusBadRequest, "trainerId query parameter is required")
		return
	}

	exercises, _, err := a.exerciseService.ListByTrainer(r.Context(), trainerID, 50, "")
	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list exercises: %v", err))
		return
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Exercises retrieved successfully",
		"data":    exercises,
	})
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
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
		if err := a.userService.CreateStudent(r.Context(), student); err != nil {
			a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create student: %v", err))
			return
		}
	} else if req.Role == string(userDomain.RoleTrainer) {
		trainer := &userDomain.Trainer{
			User:     *user,
			Metadata: userDomain.TrainerMetadata{},
		}
		if err := a.userService.CreateTrainer(r.Context(), trainer); err != nil {
			a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create trainer: %v", err))
			return
		}
	} else {
		a.errorResponse(w, http.StatusBadRequest, "Invalid user type. Must be 'TRAINER' or 'STUDENT'")
		return
	}

	a.jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "User created successfully",
		"userId":  user.ID,
	})
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userType := r.URL.Query().Get("type")

	if userType == "student" {
		student, err := a.userService.GetStudent(r.Context(), id)
		if err != nil {
			a.errorResponse(w, http.StatusNotFound, "Student not found")
			return
		}
		a.jsonResponse(w, http.StatusOK, map[string]interface{}{
			"message": "Student retrieved successfully",
			"data":    student,
		})
	} else if userType == "trainer" {
		trainer, err := a.userService.GetTrainer(r.Context(), id)
		if err != nil {
			a.errorResponse(w, http.StatusNotFound, "Trainer not found")
			return
		}
		a.jsonResponse(w, http.StatusOK, map[string]interface{}{
			"message": "Trainer retrieved successfully",
			"data":    trainer,
		})
	} else {
		a.errorResponse(w, http.StatusBadRequest, "User type query parameter is required (student|trainer)")
	}
}

func (a *App) assignTrainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainerID := vars["trainerId"]

	var req AssignTrainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.TrainerID = trainerID

	if err := a.userService.AssignStudentToTrainer(r.Context(), req.TrainerID, req.StudentID); err != nil {
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to assign trainer: %v", err))
		return
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Trainer assigned successfully",
	})
}

func (a *App) assignTrainerByEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainerID := vars["trainerId"]

	var req AssignTrainerByEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		a.errorResponse(w, http.StatusBadRequest, "email is required")
		return
	}

	studentID, err := a.authService.ResolveUserIDByEmail(r.Context(), req.Email)
	if err != nil {
		if err == authDomain.ErrUserNotFound {
			a.errorResponse(w, http.StatusNotFound, "Student not found")
			return
		}
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to find user by email: %v", err))
		return
	}

	if _, err := a.userService.GetStudent(r.Context(), studentID); err != nil {
		if _, trainerErr := a.userService.GetTrainer(r.Context(), studentID); trainerErr == nil {
			a.errorResponse(w, http.StatusBadRequest, "User is not a student")
			return
		}
		a.errorResponse(w, http.StatusNotFound, "Student not found")
		return
	}

	if err := a.userService.AssignStudentToTrainer(r.Context(), trainerID, studentID); err != nil {
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to assign trainer: %v", err))
		return
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Trainer assigned successfully",
	})
}

func (a *App) listStudents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainerID := vars["trainerId"]

	students, _, err := a.userService.ListStudentsByTrainer(r.Context(), trainerID, 50, "")
	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list students: %v", err))
		return
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Students retrieved successfully",
		"data":    students,
	})
}

func (a *App) jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (a *App) errorResponse(w http.ResponseWriter, statusCode int, message string) {
	a.jsonResponse(w, statusCode, map[string]interface{}{
		"error": message,
	})
}

// Register handler
func (a *App) register(w http.ResponseWriter, r *http.Request) {
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check if email already exists
	_, err := a.authRepo.GetCredentialsByEmail(r.Context(), req.Email)
	if err == nil {
		a.errorResponse(w, http.StatusConflict, "Email already registered")
		return
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
	if err := a.authService.CreateCredentials(r.Context(), user.ID, req.Email, req.Password); err != nil {
		a.log.Error("failed to create credentials", err)
		a.errorResponse(w, http.StatusInternalServerError, "Failed to create user credentials")
		return
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
		if err := a.userService.CreateStudent(r.Context(), student); err != nil {
			a.log.Error("failed to create student", err)
			a.errorResponse(w, http.StatusInternalServerError, "Failed to create student")
			return
		}
	} else if req.Role == string(userDomain.RoleTrainer) {
		trainer := &userDomain.Trainer{
			User:     *user,
			Metadata: userDomain.TrainerMetadata{},
		}
		if err := a.userService.CreateTrainer(r.Context(), trainer); err != nil {
			a.log.Error("failed to create trainer", err)
			a.errorResponse(w, http.StatusInternalServerError, "Failed to create trainer")
			return
		}
	} else {
		a.errorResponse(w, http.StatusBadRequest, "Invalid user type. Must be 'TRAINER' or 'STUDENT'")
		return
	}

	// Auto-login after registration
	loginReq := &authDomain.AuthRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	response, err := a.authService.Login(r.Context(), loginReq)
	if err != nil {
		a.log.Error("auto-login after registration failed", err)
		// Return success anyway, user can login manually
		a.jsonResponse(w, http.StatusCreated, map[string]interface{}{
			"message": "User created successfully",
			"userId":  user.ID,
		})
		return
	}

	a.jsonResponse(w, http.StatusCreated, response)
}

// Auth handlers
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var req authDomain.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := a.authService.Login(r.Context(), &req)
	if err != nil {
		if err == authDomain.ErrInvalidCredentials || err == authDomain.ErrUserNotFound || err == authDomain.ErrUserInactive {
			a.errorResponse(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		a.log.Error("login failed", err)
		a.errorResponse(w, http.StatusInternalServerError, "Login failed")
		return
	}

	a.jsonResponse(w, http.StatusOK, response)
}

func (a *App) refreshToken(w http.ResponseWriter, r *http.Request) {
	var req authDomain.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tokenPair, err := a.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if err == authDomain.ErrTokenExpired || err == authDomain.ErrTokenInvalid || err == authDomain.ErrInvalidRefresh {
			a.errorResponse(w, http.StatusUnauthorized, "Invalid refresh token")
			return
		}
		a.log.Error("refresh token failed", err)
		a.errorResponse(w, http.StatusInternalServerError, "Token refresh failed")
		return
	}

	a.jsonResponse(w, http.StatusOK, tokenPair)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		a.errorResponse(w, http.StatusUnauthorized, "Authorization header required")
		return
	}

	// Extract Bearer token
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		a.errorResponse(w, http.StatusUnauthorized, "Invalid authorization header")
		return
	}

	if err := a.authService.LogoutFromToken(r.Context(), tokenParts[1]); err != nil {
		a.log.Error("logout failed", err)
		a.errorResponse(w, http.StatusInternalServerError, "Logout failed")
		return
	}

	a.jsonResponse(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	} else {
		fmt.Println(".env file loaded successfully")
	}

	// Debug: Check if RUN_LOCAL is loaded
	runLocal := os.Getenv("RUN_LOCAL")
	fmt.Printf("RUN_LOCAL from env: %s\n", runLocal)

	router := mux.NewRouter()

	// Catch-all OPTIONS handler for CORS
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		// If not OPTIONS, continue to next handler
		http.NotFound(w, r)
	}).Methods("OPTIONS")

	// Exercise routes
	router.HandleFunc("/api/v1/exercises", app.createExercise).Methods("POST")
	router.HandleFunc("/api/v1/exercises/{id}", app.getExercise).Methods("GET")
	router.HandleFunc("/api/v1/exercises", app.listExercises).Methods("GET")

	// Auth routes
	router.HandleFunc("/api/v1/auth/register", app.register).Methods("POST")
	router.HandleFunc("/api/v1/auth/login", app.login).Methods("POST")
	router.HandleFunc("/api/v1/auth/refresh", app.refreshToken).Methods("POST")
	router.HandleFunc("/api/v1/auth/logout", app.logout).Methods("POST")

	// User routes
	router.HandleFunc("/api/v1/users", app.createUser).Methods("POST")
	router.HandleFunc("/api/v1/users/{id}", app.getUser).Methods("GET")
	router.HandleFunc("/api/v1/trainers/{trainerId}/students", app.assignTrainer).Methods("POST")
	router.HandleFunc("/api/v1/trainers/{trainerId}/students/assign-by-email", app.assignTrainerByEmail).Methods("POST")
	router.HandleFunc("/api/v1/trainers/{trainerId}/students", app.listStudents).Methods("GET")

	// Routine routes
	router.HandleFunc("/api/v1/trainers/{trainerId}/routines", app.createRoutine).Methods("POST")
	router.HandleFunc("/api/v1/routines/{id}", app.getRoutine).Methods("GET")
	router.HandleFunc("/api/v1/trainers/{trainerId}/routines/{routineId}", app.updateRoutine).Methods("PUT")
	router.HandleFunc("/api/v1/trainers/{trainerId}/routines/{routineId}", app.deleteRoutine).Methods("DELETE")
	router.HandleFunc("/api/v1/trainers/{trainerId}/routines", app.listRoutinesByTrainer).Methods("GET")
	router.HandleFunc("/api/v1/students/{studentId}/routines", app.listRoutinesByStudent).Methods("GET")
	router.HandleFunc("/api/v1/students/{studentId}/routines/active", app.getActiveRoutineForStudent).Methods("GET")
	router.HandleFunc("/api/v1/trainers/{trainerId}/routines/{routineId}/workout-days", app.createWorkoutDay).Methods("POST")
	router.HandleFunc("/api/v1/routines/{routineId}/workout-days", app.getWorkoutDays).Methods("GET")
	router.HandleFunc("/api/v1/trainers/{trainerId}/routines/{routineId}/workout-days/{weekNumber}/{dayNumber}", app.updateWorkoutDay).Methods("PUT")
	router.HandleFunc("/api/v1/trainers/{trainerId}/routines/assign", app.assignRoutine).Methods("POST")

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

	appCfg, _ := appConfig.Load()
	port := appCfg.App.Port
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		panic(err)
	}
}

// Routine handlers
func (a *App) createRoutine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainerID := vars["trainerId"]

	var req CreateRoutineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	domainReq := &routineDomain.CreateRoutineRequest{
		Name:        req.Name,
		WeekCount:   req.WeekCount,
		Description: req.Description,
	}

	created, err := a.routineService.CreateRoutine(r.Context(), trainerID, domainReq)
	if err != nil {
		a.log.Error("failed to create routine", err)
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create routine: %v", err))
		return
	}

	a.jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message":   "Routine created successfully",
		"routineId": created.ID,
		"data":      created,
	})
}

func (a *App) getRoutine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	found, err := a.routineService.GetRoutine(r.Context(), id)
	if err != nil {
		if errors.Is(err, routineDomain.ErrRoutineNotFound) {
			a.errorResponse(w, http.StatusNotFound, "Routine not found")
			return
		}
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get routine: %v", err))
		return
	}

	// Obtener los workout days de la rutina
	workoutDays, err := a.routineService.GetWorkoutDays(r.Context(), id)
	if err != nil {
		a.log.Error("failed to get workout days", err)
		// No fallar si no hay workout days, solo loguear el error
		workoutDays = []*routineDomain.WorkoutDay{}
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "Routine retrieved successfully",
		"data":        found,
		"workoutDays": workoutDays,
	})
}

func (a *App) updateRoutine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainerID := vars["trainerId"]
	routineID := vars["routineId"]

	var req UpdateRoutineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	domainReq := &routineDomain.UpdateRoutineRequest{
		Name:        req.Name,
		Description: req.Description,
		Status:      routineDomain.RoutineStatus(req.Status),
	}

	updated, err := a.routineService.UpdateRoutine(r.Context(), trainerID, routineID, domainReq)
	if err != nil {
		if errors.Is(err, routineDomain.ErrRoutineNotFound) {
			a.errorResponse(w, http.StatusNotFound, "Routine not found")
			return
		}
		a.log.Error("failed to update routine", err)
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update routine: %v", err))
		return
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Routine updated successfully",
		"data":    updated,
	})
}

func (a *App) deleteRoutine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainerID := vars["trainerId"]
	routineID := vars["routineId"]

	if err := a.routineService.DeleteRoutine(r.Context(), trainerID, routineID); err != nil {
		if errors.Is(err, routineDomain.ErrRoutineNotFound) {
			a.errorResponse(w, http.StatusNotFound, "Routine not found")
			return
		}
		a.log.Error("failed to delete routine", err)
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete routine: %v", err))
		return
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Routine deleted successfully",
	})
}

func (a *App) listRoutinesByTrainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainerID := vars["trainerId"]

	limit := 50
	startKey := r.URL.Query().Get("startKey")

	routines, nextKey, err := a.routineService.ListRoutinesByTrainer(r.Context(), trainerID, limit, startKey)
	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list routines: %v", err))
		return
	}

	response := map[string]interface{}{
		"message": "Routines retrieved successfully",
		"data":    routines,
	}
	if nextKey != "" {
		response["nextKey"] = nextKey
	}

	a.jsonResponse(w, http.StatusOK, response)
}

// listRoutinesByStudent now returns assignments instead of routines
func (a *App) listRoutinesByStudent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	studentID := vars["studentId"]

	assignments, err := a.assignmentService.GetAssignmentsByStudent(r.Context(), studentID)
	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list assignments: %v", err))
		return
	}

	// Enrich assignments with routine data
	type AssignmentWithRoutine struct {
		Assignment *routineDomain.RoutineAssignment `json:"assignment"`
		Routine    *routineDomain.Routine           `json:"routine"`
	}

	enrichedAssignments := make([]AssignmentWithRoutine, 0, len(assignments))
	for _, assignment := range assignments {
		routine, err := a.routineService.GetRoutine(r.Context(), assignment.RoutineID)
		if err != nil {
			a.log.Error("failed to get routine for assignment", err)
			continue
		}
		enrichedAssignments = append(enrichedAssignments, AssignmentWithRoutine{
			Assignment: assignment,
			Routine:    routine,
		})
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Assignments retrieved successfully",
		"data":    enrichedAssignments,
	})
}

// getActiveRoutineForStudent now returns the active assignment with routine data
func (a *App) getActiveRoutineForStudent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	studentID := vars["studentId"]

	assignment, err := a.assignmentService.GetActiveAssignmentForStudent(r.Context(), studentID)
	if err != nil {
		if errors.Is(err, routineDomain.ErrRoutineNotFound) {
			a.jsonResponse(w, http.StatusOK, map[string]interface{}{
				"message": "No active assignment found",
				"data":    nil,
			})
			return
		}
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get active assignment: %v", err))
		return
	}

	// Get the routine data
	routine, err := a.routineService.GetRoutine(r.Context(), assignment.RoutineID)
	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get routine: %v", err))
		return
	}

	// Get workout days
	workoutDays, err := a.routineService.GetWorkoutDays(r.Context(), assignment.RoutineID)
	if err != nil {
		a.log.Error("failed to get workout days", err)
		workoutDays = []*routineDomain.WorkoutDay{}
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "Active assignment retrieved successfully",
		"assignment":  assignment,
		"routine":     routine,
		"workoutDays": workoutDays,
	})
}

func (a *App) createWorkoutDay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainerID := vars["trainerId"]
	routineID := vars["routineId"]

	var req CreateWorkoutDayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	exercises := make([]routineDomain.ExerciseSet, len(req.Exercises))
	for i, ex := range req.Exercises {
		exercises[i] = routineDomain.ExerciseSet{
			ExerciseID:  ex.ExerciseID,
			Order:       ex.Order,
			Sets:        ex.Sets,
			Reps:        ex.Reps,
			RestSeconds: ex.RestSeconds,
			Notes:       ex.Notes,
			Tempo:       ex.Tempo,
			RPE:         ex.RPE,
		}
	}

	workoutDay := &routineDomain.WorkoutDay{
		WeekNumber: req.WeekNumber,
		DayNumber:  req.DayNumber,
		DayName:    req.DayName,
		IsRestDay:  req.IsRestDay,
		Exercises:  exercises,
	}

	if err := a.routineService.CreateWorkoutDay(r.Context(), trainerID, routineID, workoutDay); err != nil {
		if errors.Is(err, routineDomain.ErrRoutineNotFound) {
			a.errorResponse(w, http.StatusNotFound, "Routine not found")
			return
		}
		a.log.Error("failed to create workout day", err)
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create workout day: %v", err))
		return
	}

	a.jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "Workout day created successfully",
		"data":    workoutDay,
	})
}

func (a *App) getWorkoutDays(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	routineID := vars["routineId"]

	workoutDays, err := a.routineService.GetWorkoutDays(r.Context(), routineID)
	if err != nil {
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get workout days: %v", err))
		return
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Workout days retrieved successfully",
		"data":    workoutDays,
	})
}

func (a *App) updateWorkoutDay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainerID := vars["trainerId"]
	routineID := vars["routineId"]
	weekNumber := vars["weekNumber"]
	dayNumber := vars["dayNumber"]

	var req CreateWorkoutDayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	exercises := make([]routineDomain.ExerciseSet, len(req.Exercises))
	for i, ex := range req.Exercises {
		exercises[i] = routineDomain.ExerciseSet{
			ExerciseID:  ex.ExerciseID,
			Order:       ex.Order,
			Sets:        ex.Sets,
			Reps:        ex.Reps,
			RestSeconds: ex.RestSeconds,
			Notes:       ex.Notes,
			Tempo:       ex.Tempo,
			RPE:         ex.RPE,
		}
	}

	// Parse week and day numbers from URL
	var weekNum, dayNum int
	fmt.Sscanf(weekNumber, "%d", &weekNum)
	fmt.Sscanf(dayNumber, "%d", &dayNum)

	workoutDay := &routineDomain.WorkoutDay{
		WeekNumber: weekNum,
		DayNumber:  dayNum,
		DayName:    req.DayName,
		IsRestDay:  req.IsRestDay,
		Exercises:  exercises,
	}

	if err := a.routineService.UpdateWorkoutDay(r.Context(), trainerID, routineID, workoutDay); err != nil {
		if errors.Is(err, routineDomain.ErrRoutineNotFound) {
			a.errorResponse(w, http.StatusNotFound, "Routine not found")
			return
		}
		a.log.Error("failed to update workout day", err)
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update workout day: %v", err))
		return
	}

	a.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Workout day updated successfully",
		"data":    workoutDay,
	})
}

func (a *App) assignRoutine(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainerID := vars["trainerId"]

	var req AssignRoutineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RoutineID == "" || req.StudentID == "" {
		a.errorResponse(w, http.StatusBadRequest, "routineId and studentId are required")
		return
	}

	// Parse dates or use defaults
	var startDate, endDate time.Time
	var err error

	if req.StartDate != "" {
		startDate, err = time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			a.errorResponse(w, http.StatusBadRequest, "Invalid startDate format. Use YYYY-MM-DD")
			return
		}
	} else {
		startDate = time.Now().UTC()
	}

	if req.EndDate != "" {
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			a.errorResponse(w, http.StatusBadRequest, "Invalid endDate format. Use YYYY-MM-DD")
			return
		}
	} else {
		// Get routine to calculate end date based on week count
		routine, err := a.routineService.GetRoutine(r.Context(), req.RoutineID)
		if err != nil {
			a.errorResponse(w, http.StatusNotFound, "Routine not found")
			return
		}
		endDate = startDate.AddDate(0, 0, (routine.WeekCount*7)-1)
	}

	if endDate.Before(startDate) {
		a.errorResponse(w, http.StatusBadRequest, "endDate must be after startDate")
		return
	}

	// Create assignment using the assignment service
	createReq := &routineDomain.CreateAssignmentRequest{
		RoutineID:  req.RoutineID,
		StudentIDs: []string{req.StudentID},
		StartDate:  startDate.Format("2006-01-02"),
		EndDate:    endDate.Format("2006-01-02"),
	}

	assignments, err := a.assignmentService.CreateAssignments(r.Context(), trainerID, createReq)
	if err != nil {
		a.log.Error("failed to assign routine", err)
		a.errorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to assign routine: %v", err))
		return
	}

	if len(assignments) == 0 {
		a.errorResponse(w, http.StatusInternalServerError, "No assignment was created")
		return
	}

	assignment := assignments[0]
	response := AssignRoutineResponse{
		ID:        assignment.ID,
		RoutineID: assignment.RoutineID,
		StudentID: assignment.StudentID,
		TrainerID: trainerID,
		StartDate: assignment.StartDate,
		EndDate:   assignment.EndDate,
		Status:    string(assignment.Status),
		CreatedAt: assignment.CreatedAt,
		UpdatedAt: assignment.UpdatedAt,
	}

	a.jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "Routine assigned successfully",
		"data":    response,
	})
}
