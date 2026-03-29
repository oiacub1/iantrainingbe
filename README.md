# Training Platform MVP

Plataforma de entrenamiento personalizada con arquitectura serverless en AWS.

## 🏗️ Stack Técnico

- **Frontend**: React (Mobile-first) con TypeScript
- **Backend**: Go 1.21+ con AWS Lambda
- **Base de Datos**: DynamoDB (Single Table Design)
- **Infraestructura**: AWS (Lambda, API Gateway, DynamoDB, Cognito)
- **IaC**: Terraform

## 📁 Estructura del Proyecto

```
iantraining/
├── cmd/                    # Lambda entry points
├── internal/              # Código privado de la aplicación
│   ├── domain/           # Entidades de negocio (DDD)
│   ├── repository/       # Capa de persistencia
│   ├── service/          # Lógica de negocio
│   └── handler/          # HTTP/Lambda handlers
├── pkg/                   # Código reutilizable
├── shared/               # Recursos compartidos (i18n)
├── infrastructure/       # Terraform/CloudFormation
├── tests/                # Tests unitarios e integración
└── docs/                 # Documentación

```

Ver [`docs/PROJECT_STRUCTURE.md`](docs/PROJECT_STRUCTURE.md) para detalles completos.

## 🗄️ Diseño de Base de Datos

DynamoDB Single Table Design optimizado para:
- Listar alumnos de un profesor
- Obtener rutina vigente de un alumno
- Consultar historial de ejercicios completados

Ver [`docs/DYNAMODB_SCHEMA.md`](docs/DYNAMODB_SCHEMA.md) para el esquema completo.

## 🌍 Internacionalización

Sistema i18n compartido entre frontend y backend con validación de keys.

Idiomas soportados:
- 🇬🇧 Inglés (base)
- 🇪🇸 Español
- 🇧🇷 Portugués

Ver [`docs/I18N_STRUCTURE.md`](docs/I18N_STRUCTURE.md) para detalles de implementación.

## 🚀 Quick Start

### Prerrequisitos

```bash
# Go 1.21+
go version

# AWS CLI configurado
aws configure

# Terraform (opcional)
terraform version

# Docker (para DynamoDB local)
docker --version
```

### Instalación

```bash
# Clonar repositorio
git clone <repo-url>
cd iantraining

# Instalar dependencias
make deps

# Configurar variables de entorno
cp .env.example .env
# Editar .env con tus credenciales
```

### Desarrollo Local

```bash
# Iniciar DynamoDB local
make local-dynamodb

# Crear tabla local
aws dynamodb create-table \
  --cli-input-json file://infrastructure/dynamodb-local.json \
  --endpoint-url http://localhost:8000

# Ejecutar API localmente
make run-api

# En otra terminal, ejecutar tests
make test
```

### Build y Deploy

```bash
# Build todas las Lambdas
make build-all

# Build Lambda específica
make build FUNCTION=exercises/create

# Deploy a AWS
make deploy

# Deploy Lambda específica
make deploy-lambda FUNCTION=exercises/create
```

## 📚 Comandos Disponibles

```bash
make help                    # Ver todos los comandos
make build-all              # Build todas las Lambdas
make test                   # Tests unitarios
make test-integration       # Tests de integración
make clean                  # Limpiar artifacts
make lint                   # Linters
make fmt                    # Formatear código
make validate-i18n          # Validar traducciones
```

## 🧪 Testing

```bash
# Tests unitarios con coverage
make test

# Tests de integración
make test-integration

# Ver coverage en browser
open coverage.html
```

## 📖 Documentación

- [Esquema DynamoDB](docs/DYNAMODB_SCHEMA.md)
- [Estructura del Proyecto](docs/PROJECT_STRUCTURE.md)
- [Sistema i18n](docs/I18N_STRUCTURE.md)
- [API Documentation](docs/API.md) (TODO)
- [Deployment Guide](docs/DEPLOYMENT.md) (TODO)

## 🏛️ Arquitectura

### Patrones de Diseño

- **Domain-Driven Design (DDD)**: Separación clara de capas
- **Repository Pattern**: Abstracción de persistencia
- **Dependency Injection**: Testabilidad y flexibilidad
- **Single Table Design**: Optimización de DynamoDB

### Flujo de Request

```
API Gateway → Lambda → Handler → Service → Repository → DynamoDB
```

### Entidades Principales

1. **Users**: Trainers y Students con roles diferenciados
2. **Exercises**: Catálogo con metadata de grupos musculares
3. **Routines**: Estructura semanal de entrenamientos
4. **Workout Logs**: Seguimiento en tiempo real

## 🔐 Seguridad

- Autenticación: AWS Cognito + JWT
- Autorización: Role-based access control
- Encriptación: DynamoDB encryption at rest
- Secrets: AWS Secrets Manager

## 🚢 CI/CD

GitHub Actions workflows implementados:

### Workflows Disponibles:

1. **Test and Validate** (`.github/workflows/test.yml`):
   - Se ejecuta en cada PR y push a main
   - Ejecuta linters, tests unitarios y build de Lambdas
   - Valida archivos i18n

2. **Deploy Lambda and Setup DynamoDB** (`.github/workflows/deploy.yml`):
   - Se ejecuta en push a main
   - Build y deploy de funciones Lambda a AWS
   - Verifica y crea tablas DynamoDB si no existen
   - Entorno único de producción

### Configuración Requerida:

1. **Secrets de GitHub**:
   - `AWS_ACCESS_KEY_ID`: Access Key ID de IAM User
   - `AWS_SECRET_ACCESS_KEY`: Secret Access Key de IAM User
   - `AWS_REGION`: Región de AWS (opcional, default: us-east-1)

2. **Variables de Entorno**:
   - Tabla DynamoDB: `training-platform`
   - Funciones Lambda: `training-platform-{nombre}`

### Flujo de Deploy:
```
Push a main → Tests → Build Lambdas → Check DynamoDB → Create/Update Tables → Deploy Lambdas → Verify
```

## 📊 Monitoreo

- **Logs**: CloudWatch Logs
- **Metrics**: CloudWatch Metrics
- **Tracing**: AWS X-Ray
- **Alertas**: CloudWatch Alarms + SNS

## 🤝 Contribución

```bash
# Crear feature branch
git checkout -b feature/nueva-funcionalidad

# Hacer cambios y tests
make test

# Formatear código
make fmt

# Commit con conventional commits
git commit -m "feat: agregar endpoint de rutinas"

# Push y crear PR
git push origin feature/nueva-funcionalidad
```

## 📝 Roadmap

### MVP (v1.0)
- [x] Diseño de arquitectura
- [x] Esquema DynamoDB
- [x] Estructura de proyecto
- [x] Sistema i18n
- [ ] Implementación de Lambdas core
- [ ] Frontend React
- [ ] Autenticación con Cognito
- [ ] Deploy a staging

### v1.1
- [ ] Notificaciones push
- [ ] Gráficos de progreso
- [ ] Export de datos
- [ ] Modo offline

### v2.0
- [ ] Planes de nutrición
- [ ] Integración con wearables
- [ ] Videollamadas trainer-alumno
- [ ] Marketplace de rutinas

## 📄 Licencia

MIT License - ver [LICENSE](LICENSE) para detalles.

## 👥 Equipo

- **Arquitecto**: Senior Fullstack Developer
- **Backend**: Go + AWS
- **Frontend**: React + TypeScript
- **DevOps**: Terraform + GitHub Actions

## 📞 Contacto

- Email: support@trainingplatform.com
- Docs: https://docs.trainingplatform.com
- Issues: GitHub Issues

---

**Hecho con ❤️ para transformar el entrenamiento personalizado**
