# DynamoDB Single Table Design - Training Platform

## Tabla Principal: `training-platform`

### Patrones de Acceso y Diseño

#### 1. **Usuarios (Personal Trainers y Alumnos)**

**Personal Trainer:**
```
PK: USER#<trainerId>
SK: PROFILE
Attributes:
  - entityType: "TRAINER"
  - email: string
  - name: string
  - phone: string
  - createdAt: timestamp
  - updatedAt: timestamp
  - status: "ACTIVE" | "INACTIVE"
  - metadata: {
      specializations: []string
      certifications: []string
    }
```

**Alumno:**
```
PK: USER#<studentId>
SK: PROFILE
Attributes:
  - entityType: "STUDENT"
  - email: string
  - name: string
  - phone: string
  - trainerId: string
  - createdAt: timestamp
  - updatedAt: timestamp
  - status: "ACTIVE" | "INACTIVE"
  - metadata: {
      goals: []string
      injuries: []string
      fitnessLevel: string
    }
```

**Relación Trainer-Student:**
```
PK: USER#<trainerId>
SK: STUDENT#<studentId>
Attributes:
  - entityType: "TRAINER_STUDENT"
  - studentName: string
  - studentEmail: string
  - assignedAt: timestamp
  - status: "ACTIVE" | "INACTIVE"
  - GSI1PK: USER#<studentId>
  - GSI1SK: TRAINER#<trainerId>
```

---

#### 2. **Ejercicios**

```
PK: EXERCISE#<exerciseId>
SK: METADATA
Attributes:
  - entityType: "EXERCISE"
  - name: string
  - nameKey: string (i18n key: "exercises.squat.name")
  - descriptionKey: string (i18n key: "exercises.squat.description")
  - youtubeUrl: string
  - thumbnailUrl: string
  - muscleGroups: [
      {
        group: "QUADRICEPS",
        groupKey: "muscleGroups.quadriceps",
        impactPercentage: 60
      },
      {
        group: "GLUTES",
        groupKey: "muscleGroups.glutes",
        impactPercentage: 30
      },
      {
        group: "CORE",
        groupKey: "muscleGroups.core",
        impactPercentage: 10
      }
    ]
  - difficulty: "BEGINNER" | "INTERMEDIATE" | "ADVANCED"
  - equipment: []string
  - createdBy: string (trainerId)
  - createdAt: timestamp
  - updatedAt: timestamp
  - GSI1PK: EXERCISES_BY_TRAINER#<trainerId>
  - GSI1SK: EXERCISE#<createdAt>
```

---

#### 3. **Rutinas**

**Rutina Principal:**
```
PK: ROUTINE#<routineId>
SK: METADATA
Attributes:
  - entityType: "ROUTINE"
  - name: string
  - nameKey: string (i18n key)
  - studentId: string
  - trainerId: string
  - startDate: date (YYYY-MM-DD)
  - endDate: date (YYYY-MM-DD)
  - status: "DRAFT" | "ACTIVE" | "COMPLETED" | "ARCHIVED"
  - weekCount: number
  - createdAt: timestamp
  - updatedAt: timestamp
  - GSI1PK: STUDENT#<studentId>
  - GSI1SK: ROUTINE#<startDate>
  - GSI2PK: TRAINER#<trainerId>
  - GSI2SK: ROUTINE#<startDate>
```

**Día de Entrenamiento:**
```
PK: ROUTINE#<routineId>
SK: WEEK#<weekNumber>#DAY#<dayNumber>
Attributes:
  - entityType: "WORKOUT_DAY"
  - weekNumber: number (1-N)
  - dayNumber: number (1-7)
  - dayName: string
  - dayNameKey: string (i18n: "days.monday")
  - isRestDay: boolean
  - exercises: [
      {
        exerciseId: string
        order: number
        sets: number
        reps: string ("10-12" | "AMRAP" | "30s")
        restSeconds: number
        notes: string
        notesKey: string (i18n key)
        tempo: string ("2-0-2-0")
        rpe: number (1-10)
      }
    ]
```

---

#### 4. **Seguimiento de Ejercicios**

**Registro de Ejercicio Completado:**
```
PK: STUDENT#<studentId>
SK: WORKOUT#<timestamp>#<exerciseId>
Attributes:
  - entityType: "WORKOUT_LOG"
  - routineId: string
  - exerciseId: string
  - exerciseName: string
  - weekNumber: number
  - dayNumber: number
  - completedAt: timestamp
  - date: date (YYYY-MM-DD)
  - sets: [
      {
        setNumber: number
        reps: number
        weight: number
        weightUnit: "KG" | "LBS"
        completed: boolean
        rpe: number
        notes: string
      }
    ]
  - totalDurationSeconds: number
  - feeling: "EXCELLENT" | "GOOD" | "AVERAGE" | "POOR"
  - notes: string
  - GSI1PK: ROUTINE#<routineId>
  - GSI1SK: WORKOUT#<timestamp>
  - GSI2PK: STUDENT#<studentId>#DATE#<date>
  - GSI2SK: WORKOUT#<timestamp>
```

**Resumen Diario:**
```
PK: STUDENT#<studentId>
SK: DAILY_SUMMARY#<date>
Attributes:
  - entityType: "DAILY_SUMMARY"
  - routineId: string
  - date: date (YYYY-MM-DD)
  - weekNumber: number
  - dayNumber: number
  - totalExercises: number
  - completedExercises: number
  - totalDurationSeconds: number
  - completionPercentage: number
  - overallFeeling: string
  - notes: string
  - completedAt: timestamp
  - GSI1PK: ROUTINE#<routineId>
  - GSI1SK: SUMMARY#<date>
```

---

## Índices Globales Secundarios (GSI)

### **GSI1: Relaciones y Búsquedas Inversas**
```
GSI1PK (Partition Key)
GSI1SK (Sort Key)
```

**Casos de Uso:**
- Listar alumnos de un trainer: `GSI1PK = TRAINER#<trainerId>`
- Obtener trainer de un alumno: `GSI1PK = USER#<studentId>` + `GSI1SK begins_with TRAINER#`
- Rutinas de un alumno: `GSI1PK = STUDENT#<studentId>` + `GSI1SK begins_with ROUTINE#`
- Ejercicios de un trainer: `GSI1PK = EXERCISES_BY_TRAINER#<trainerId>`
- Logs de una rutina: `GSI1PK = ROUTINE#<routineId>` + `GSI1SK begins_with WORKOUT#`

### **GSI2: Búsquedas por Fecha**
```
GSI2PK (Partition Key)
GSI2SK (Sort Key)
```

**Casos de Uso:**
- Rutinas de un trainer por fecha: `GSI2PK = TRAINER#<trainerId>` + `GSI2SK begins_with ROUTINE#`
- Workouts de un alumno por fecha: `GSI2PK = STUDENT#<studentId>#DATE#<date>`
- Historial de entrenamientos en rango de fechas

---

## Queries Principales

### 1. **Listar alumnos de un profesor**
```go
// Query en tabla principal
PK = USER#<trainerId>
SK begins_with STUDENT#

// Retorna todos los estudiantes asignados al trainer
```

### 2. **Obtener la rutina vigente de un alumno**
```go
// Query en GSI1
GSI1PK = STUDENT#<studentId>
GSI1SK begins_with ROUTINE#
FilterExpression: status = 'ACTIVE' AND startDate <= today AND endDate >= today

// Luego obtener los días de entrenamiento:
PK = ROUTINE#<routineId>
SK begins_with WEEK#
```

### 3. **Consultar historial de ejercicios completados**

**Por alumno (últimos N días):**
```go
// Query en tabla principal
PK = STUDENT#<studentId>
SK between WORKOUT#<startTimestamp> AND WORKOUT#<endTimestamp>
```

**Por rutina:**
```go
// Query en GSI1
GSI1PK = ROUTINE#<routineId>
GSI1SK begins_with WORKOUT#
```

**Por fecha específica:**
```go
// Query en GSI2
GSI2PK = STUDENT#<studentId>#DATE#<date>
GSI2SK begins_with WORKOUT#
```

### 4. **Obtener resumen semanal de progreso**
```go
// Query en tabla principal
PK = STUDENT#<studentId>
SK between DAILY_SUMMARY#<startDate> AND DAILY_SUMMARY#<endDate>
```

---

## Consideraciones de Diseño

### **Ventajas del Single Table Design:**
1. **Reducción de costos**: Una sola tabla reduce RCU/WCU
2. **Queries eficientes**: Relaciones pre-calculadas en PK/SK
3. **Escalabilidad**: Particionamiento natural por usuario
4. **Atomicidad**: TransactWriteItems para operaciones relacionadas

### **Patrones de Acceso Optimizados:**
- Hot partitions evitadas usando IDs únicos en PK
- SK diseñados para range queries eficientes
- GSI diseñados para queries inversas sin scans
- Datos desnormalizados para reducir queries

### **Límites a Considerar:**
- Item size: Max 400KB
- Query result: Max 1MB (usar pagination)
- GSI: Max 20 por tabla (usamos 2)
- Batch operations: Max 25 items

### **Estrategias de Paginación:**
- Usar `LastEvaluatedKey` para queries grandes
- Implementar cursor-based pagination en API
- Cache de resultados frecuentes en Lambda

### **Backup y Recuperación:**
- Point-in-time recovery habilitado
- Backups automáticos diarios
- DynamoDB Streams para auditoría y replicación
