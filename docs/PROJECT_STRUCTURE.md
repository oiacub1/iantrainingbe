# Go Lambda Project Structure - Training Platform

## Estructura de Carpetas

```
iantraining/
├── cmd/
│   ├── api/
│   │   ├── main.go                    # API Gateway Lambda handler
│   │   └── router.go                  # Route definitions
│   ├── trainers/
│   │   ├── create/
│   │   │   └── main.go               # Lambda: Create trainer
│   │   ├── list-students/
│   │   │   └── main.go               # Lambda: List students of trainer
│   │   └── get/
│   │       └── main.go               # Lambda: Get trainer profile
│   ├── students/
│   │   ├── create/
│   │   │   └── main.go               # Lambda: Create student
│   │   ├── assign-trainer/
│   │   │   └── main.go               # Lambda: Assign trainer to student
│   │   └── get-routine/
│   │       └── main.go               # Lambda: Get active routine
│   ├── exercises/
│   │   ├── create/
│   │   │   └── main.go               # Lambda: Create exercise
│   │   ├── list/
│   │   │   └── main.go               # Lambda: List exercises
│   │   └── get/
│   │       └── main.go               # Lambda: Get exercise details
│   ├── routines/
│   │   ├── create/
│   │   │   └── main.go               # Lambda: Create routine
│   │   ├── get/
│   │   │   └── main.go               # Lambda: Get routine with days
│   │   └── update-status/
│   │       └── main.go               # Lambda: Update routine status
│   ├── workouts/
│   │   ├── log-exercise/
│   │   │   └── main.go               # Lambda: Log completed exercise
│   │   ├── get-history/
│   │   │   └── main.go               # Lambda: Get workout history
│   │   └── daily-summary/
│   │       └── main.go               # Lambda: Get/Create daily summary
│   └── migrations/
│       └── seed-data/
│           └── main.go               # Lambda: Seed initial data
├── internal/
│   ├── domain/
│   │   ├── user/
│   │   │   ├── user.go              # User domain models
│   │   │   ├── trainer.go           # Trainer entity
│   │   │   ├── student.go           # Student entity
│   │   │   └── repository.go        # User repository interface
│   │   ├── exercise/
│   │   │   ├── exercise.go          # Exercise entity
│   │   │   ├── muscle_group.go      # Muscle group value object
│   │   │   └── repository.go        # Exercise repository interface
│   │   ├── routine/
│   │   │   ├── routine.go           # Routine entity
│   │   │   ├── workout_day.go       # Workout day entity
│   │   │   ├── exercise_set.go      # Exercise set value object
│   │   │   └── repository.go        # Routine repository interface
│   │   └── workout/
│   │       ├── workout_log.go       # Workout log entity
│   │       ├── daily_summary.go     # Daily summary entity
│   │       ├── set_log.go           # Set log value object
│   │       └── repository.go        # Workout repository interface
│   ├── repository/
│   │   ├── dynamodb/
│   │   │   ├── client.go            # DynamoDB client wrapper
│   │   │   ├── user_repository.go   # User repository implementation
│   │   │   ├── exercise_repository.go
│   │   │   ├── routine_repository.go
│   │   │   ├── workout_repository.go
│   │   │   └── keys.go              # PK/SK key builders
│   │   └── cache/
│   │       └── redis_cache.go       # Optional: Redis cache layer
│   ├── service/
│   │   ├── user/
│   │   │   ├── service.go           # User business logic
│   │   │   └── validator.go         # User validation
│   │   ├── exercise/
│   │   │   ├── service.go           # Exercise business logic
│   │   │   └── validator.go         # Exercise validation
│   │   ├── routine/
│   │   │   ├── service.go           # Routine business logic
│   │   │   ├── builder.go           # Routine builder
│   │   │   └── validator.go         # Routine validation
│   │   └── workout/
│   │       ├── service.go           # Workout business logic
│   │       ├── progress.go          # Progress calculation
│   │       └── validator.go         # Workout validation
│   ├── handler/
│   │   ├── middleware/
│   │   │   ├── auth.go              # JWT authentication
│   │   │   ├── cors.go              # CORS handler
│   │   │   ├── logger.go            # Request logging
│   │   │   └── error.go             # Error handling
│   │   ├── request/
│   │   │   ├── user.go              # User request DTOs
│   │   │   ├── exercise.go          # Exercise request DTOs
│   │   │   ├── routine.go           # Routine request DTOs
│   │   │   └── workout.go           # Workout request DTOs
│   │   ├── response/
│   │   │   ├── user.go              # User response DTOs
│   │   │   ├── exercise.go          # Exercise response DTOs
│   │   │   ├── routine.go           # Routine response DTOs
│   │   │   ├── workout.go           # Workout response DTOs
│   │   │   └── error.go             # Error response format
│   │   └── lambda/
│   │       ├── base.go              # Base lambda handler
│   │       └── response.go          # Lambda response builder
│   └── i18n/
│       ├── loader.go                # i18n loader
│       ├── validator.go             # i18n key validator
│       └── constants.go             # i18n constants
├── pkg/
│   ├── config/
│   │   ├── config.go                # Configuration loader
│   │   └── aws.go                   # AWS configuration
│   ├── logger/
│   │   └── logger.go                # Structured logger (zap/zerolog)
│   ├── errors/
│   │   ├── errors.go                # Custom error types
│   │   └── codes.go                 # Error codes
│   ├── validator/
│   │   └── validator.go             # Request validator (go-playground/validator)
│   ├── auth/
│   │   ├── jwt.go                   # JWT utilities
│   │   └── cognito.go               # AWS Cognito integration
│   └── utils/
│       ├── time.go                  # Time utilities
│       ├── string.go                # String utilities
│       └── pagination.go            # Pagination helpers
├── shared/
│   └── i18n/
│       ├── en.json                  # English translations
│       ├── es.json                  # Spanish translations
│       ├── pt.json                  # Portuguese translations
│       └── schema.json              # i18n schema validation
├── scripts/
│   ├── build.sh                     # Build all lambdas
│   ├── deploy.sh                    # Deploy to AWS
│   ├── test.sh                      # Run tests
│   └── local-dynamodb.sh            # Start local DynamoDB
├── infrastructure/
│   ├── terraform/
│   │   ├── main.tf                  # Main Terraform config
│   │   ├── dynamodb.tf              # DynamoDB table definition
│   │   ├── lambda.tf                # Lambda functions
│   │   ├── api_gateway.tf           # API Gateway
│   │   ├── cognito.tf               # Cognito user pool
│   │   └── variables.tf             # Variables
│   └── cloudformation/
│       └── template.yaml            # Alternative: CloudFormation/SAM
├── tests/
│   ├── integration/
│   │   ├── user_test.go
│   │   ├── exercise_test.go
│   │   ├── routine_test.go
│   │   └── workout_test.go
│   ├── unit/
│   │   ├── service/
│   │   │   └── ...
│   │   └── repository/
│   │       └── ...
│   └── fixtures/
│       ├── users.json
│       ├── exercises.json
│       └── routines.json
├── docs/
│   ├── DYNAMODB_SCHEMA.md           # This file
│   ├── PROJECT_STRUCTURE.md         # This file
│   ├── API.md                       # API documentation
│   └── DEPLOYMENT.md                # Deployment guide
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Descripción de Componentes

### **cmd/** - Entry Points
Cada subdirectorio contiene un `main.go` que es el entry point de una Lambda específica. Esto permite:
- Despliegue independiente de cada función
- Optimización de cold starts (solo se carga el código necesario)
- Escalado granular por función

### **internal/domain/** - Domain Layer (DDD)
Contiene las entidades de negocio y las interfaces de repositorio. No tiene dependencias externas.
- **Entities**: Estructuras de datos del dominio
- **Value Objects**: Objetos inmutables (MuscleGroup, SetLog)
- **Repository Interfaces**: Contratos para persistencia

### **internal/repository/** - Data Access Layer
Implementaciones concretas de los repositorios. Maneja la persistencia en DynamoDB.
- **Keys.go**: Funciones helper para construir PK/SK
- Mapeo entre domain entities y DynamoDB items

### **internal/service/** - Business Logic Layer
Lógica de negocio, validaciones y orquestación entre repositorios.
- No conoce detalles de HTTP o Lambda
- Reutilizable entre diferentes handlers

### **internal/handler/** - Presentation Layer
Maneja requests/responses HTTP y Lambda events.
- **request/response**: DTOs para API
- **middleware**: Cross-cutting concerns
- **lambda**: Lambda-specific utilities

### **pkg/** - Shared Packages
Código reutilizable que podría ser extraído a librerías externas.
- Sin dependencias de `internal/`
- Puede ser usado por múltiples proyectos

### **shared/i18n/** - Internationalization
Archivos JSON con traducciones compartidas entre frontend y backend.

---

## Ejemplo de Flujo: Crear Rutina

```
1. cmd/routines/create/main.go
   ↓ (recibe Lambda event)
2. internal/handler/lambda/base.go
   ↓ (parsea request)
3. internal/handler/request/routine.go
   ↓ (valida DTO)
4. internal/service/routine/service.go
   ↓ (lógica de negocio)
5. internal/repository/dynamodb/routine_repository.go
   ↓ (persiste en DynamoDB)
6. internal/handler/response/routine.go
   ↓ (formatea respuesta)
7. internal/handler/lambda/response.go
   ↓ (retorna Lambda response)
```

---

## Ventajas de esta Estructura

### **Escalabilidad**
- Cada Lambda es independiente
- Fácil agregar nuevas funciones sin afectar existentes
- Clear separation of concerns

### **Testabilidad**
- Domain layer sin dependencias externas
- Interfaces permiten mocking fácil
- Tests unitarios e integración separados

### **Mantenibilidad**
- Estructura clara y predecible
- Fácil onboarding de nuevos developers
- Cambios aislados por capa

### **Performance**
- Lambdas pequeñas = cold starts rápidos
- Código compartido en layers (opcional)
- Cache layer preparado para Redis/ElastiCache

---

## Comandos Útiles

### Build todas las Lambdas:
```bash
make build-all
```

### Deploy específica Lambda:
```bash
make deploy-lambda FUNCTION=trainers/create
```

### Run tests:
```bash
make test
make test-integration
```

### Local development:
```bash
make local-dynamodb
make run-api
```
