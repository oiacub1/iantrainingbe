# DynamoDB Setup Scripts

## Scripts para crear las tablas de DynamoDB

### 1. setup-dynamodb.sh
Script para crear la tabla en AWS DynamoDB (producción/development)

**Uso:**
```bash
# Variables de entorno opcionales
export DYNAMODB_TABLE_NAME="iantraining"  # Default: iantraining
export AWS_REGION="us-east-1"             # Default: us-east-1

# Ejecutar el script
./scripts/setup-dynamodb.sh
```

**Requisitos:**
- AWS CLI configurado con credenciales válidas
- Permisos para crear tablas DynamoDB

### 2. setup-dynamodb-local.sh
Script para crear la tabla en DynamoDB Local (desarrollo local)

**Uso:**
```bash
# Variables de entorno opcionales
export DYNAMODB_TABLE_NAME="iantraining"  # Default: iantraining
export DYNAMODB_ENDPOINT="http://localhost:8000"  # Default: localhost:8000

# Ejecutar el script
./scripts/setup-dynamodb-local.sh
```

**Requisitos:**
- DynamoDB Local corriendo en localhost:8000
- AWS CLI instalado

## Estructura de la tabla

La tabla creada incluye:

- **Tabla principal** con clave compuesta (PK/SK)
- **GSI1** para consultas por trainer
- **Billing mode**: Pay-per-request
- **Projection**: ALL para todos los índices

## Variables de entorno

| Variable | Descripción | Default |
|----------|-------------|---------|
| `DYNAMODB_TABLE_NAME` | Nombre de la tabla | `iantraining` |
| `AWS_REGION` | Región AWS | `us-east-1` |
| `DYNAMODB_ENDPOINT` | Endpoint para local | `http://localhost:8000` |
