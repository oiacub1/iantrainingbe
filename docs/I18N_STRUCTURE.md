# Internationalization (i18n) Structure

## Arquitectura de Traducciones

### Principios de Diseño
1. **Single Source of Truth**: Archivos JSON en `shared/i18n/`
2. **Type-safe**: Schema JSON validado en backend
3. **Consumible por Frontend**: React puede importar directamente
4. **Validado por Backend**: Go valida keys antes de usar

---

## Estructura de Archivos

### **shared/i18n/en.json** (Inglés - Base)
```json
{
  "version": "1.0.0",
  "locale": "en",
  "translations": {
    "common": {
      "app_name": "Training Platform",
      "welcome": "Welcome",
      "loading": "Loading...",
      "error": "Error",
      "success": "Success",
      "save": "Save",
      "cancel": "Cancel",
      "delete": "Delete",
      "edit": "Edit",
      "create": "Create",
      "back": "Back",
      "next": "Next",
      "previous": "Previous",
      "confirm": "Confirm",
      "search": "Search",
      "filter": "Filter",
      "sort": "Sort"
    },
    "auth": {
      "login": "Login",
      "logout": "Logout",
      "register": "Register",
      "email": "Email",
      "password": "Password",
      "forgot_password": "Forgot password?",
      "reset_password": "Reset password",
      "invalid_credentials": "Invalid email or password",
      "session_expired": "Your session has expired"
    },
    "roles": {
      "trainer": "Personal Trainer",
      "student": "Student",
      "admin": "Administrator"
    },
    "users": {
      "profile": "Profile",
      "name": "Name",
      "phone": "Phone",
      "status": "Status",
      "created_at": "Created at",
      "updated_at": "Updated at"
    },
    "trainers": {
      "my_students": "My Students",
      "add_student": "Add Student",
      "student_count": "{{count}} students",
      "specializations": "Specializations",
      "certifications": "Certifications",
      "no_students": "You don't have any students yet"
    },
    "students": {
      "my_trainer": "My Trainer",
      "assign_trainer": "Assign Trainer",
      "goals": "Goals",
      "injuries": "Injuries",
      "fitness_level": "Fitness Level",
      "current_routine": "Current Routine",
      "no_routine": "No active routine assigned"
    },
    "exercises": {
      "title": "Exercises",
      "create_exercise": "Create Exercise",
      "exercise_name": "Exercise Name",
      "youtube_url": "YouTube URL",
      "thumbnail": "Thumbnail",
      "muscle_groups": "Muscle Groups",
      "difficulty": "Difficulty",
      "equipment": "Equipment",
      "description": "Description",
      "squat": {
        "name": "Barbell Back Squat",
        "description": "Compound lower body exercise targeting quadriceps, glutes, and hamstrings"
      },
      "bench_press": {
        "name": "Barbell Bench Press",
        "description": "Compound upper body exercise targeting chest, shoulders, and triceps"
      },
      "deadlift": {
        "name": "Conventional Deadlift",
        "description": "Full body compound exercise targeting posterior chain"
      },
      "pull_up": {
        "name": "Pull-Up",
        "description": "Bodyweight exercise targeting back and biceps"
      },
      "overhead_press": {
        "name": "Overhead Press",
        "description": "Compound shoulder exercise"
      }
    },
    "muscle_groups": {
      "quadriceps": "Quadriceps",
      "hamstrings": "Hamstrings",
      "glutes": "Glutes",
      "calves": "Calves",
      "chest": "Chest",
      "back": "Back",
      "shoulders": "Shoulders",
      "biceps": "Biceps",
      "triceps": "Triceps",
      "forearms": "Forearms",
      "core": "Core",
      "abs": "Abs",
      "obliques": "Obliques",
      "lower_back": "Lower Back",
      "traps": "Trapezius"
    },
    "difficulty": {
      "beginner": "Beginner",
      "intermediate": "Intermediate",
      "advanced": "Advanced"
    },
    "equipment": {
      "barbell": "Barbell",
      "dumbbell": "Dumbbell",
      "kettlebell": "Kettlebell",
      "bodyweight": "Bodyweight",
      "resistance_band": "Resistance Band",
      "cable": "Cable Machine",
      "machine": "Machine",
      "bench": "Bench",
      "pull_up_bar": "Pull-up Bar",
      "none": "No Equipment"
    },
    "routines": {
      "title": "Routines",
      "create_routine": "Create Routine",
      "routine_name": "Routine Name",
      "start_date": "Start Date",
      "end_date": "End Date",
      "weeks": "Weeks",
      "status": "Status",
      "active_routine": "Active Routine",
      "draft": "Draft",
      "active": "Active",
      "completed": "Completed",
      "archived": "Archived",
      "week": "Week {{number}}",
      "day": "Day {{number}}",
      "rest_day": "Rest Day",
      "workout_day": "Workout Day"
    },
    "days": {
      "monday": "Monday",
      "tuesday": "Tuesday",
      "wednesday": "Wednesday",
      "thursday": "Thursday",
      "friday": "Friday",
      "saturday": "Saturday",
      "sunday": "Sunday"
    },
    "workout": {
      "sets": "Sets",
      "reps": "Reps",
      "rest": "Rest",
      "tempo": "Tempo",
      "rpe": "RPE",
      "weight": "Weight",
      "duration": "Duration",
      "notes": "Notes",
      "log_workout": "Log Workout",
      "complete_set": "Complete Set",
      "mark_complete": "Mark Complete",
      "set_number": "Set {{number}}",
      "rest_seconds": "{{seconds}}s rest",
      "rest_minutes": "{{minutes}}min rest"
    },
    "workout_log": {
      "title": "Workout History",
      "today": "Today",
      "this_week": "This Week",
      "this_month": "This Month",
      "completed_at": "Completed at",
      "total_duration": "Total Duration",
      "exercises_completed": "Exercises Completed",
      "feeling": "How did you feel?",
      "excellent": "Excellent",
      "good": "Good",
      "average": "Average",
      "poor": "Poor",
      "no_workouts": "No workouts logged yet"
    },
    "progress": {
      "title": "Progress",
      "weekly_summary": "Weekly Summary",
      "completion_rate": "Completion Rate",
      "total_workouts": "Total Workouts",
      "total_volume": "Total Volume",
      "personal_records": "Personal Records",
      "streak": "Streak",
      "days_streak": "{{count}} day streak"
    },
    "validation": {
      "required": "This field is required",
      "invalid_email": "Invalid email address",
      "invalid_url": "Invalid URL",
      "min_length": "Minimum length is {{min}} characters",
      "max_length": "Maximum length is {{max}} characters",
      "min_value": "Minimum value is {{min}}",
      "max_value": "Maximum value is {{max}}",
      "invalid_date": "Invalid date",
      "date_in_past": "Date must be in the future",
      "end_before_start": "End date must be after start date"
    },
    "errors": {
      "generic": "An error occurred. Please try again.",
      "network": "Network error. Please check your connection.",
      "unauthorized": "You are not authorized to perform this action.",
      "not_found": "Resource not found.",
      "server_error": "Server error. Please try again later.",
      "validation_failed": "Validation failed. Please check your input."
    },
    "notifications": {
      "routine_assigned": "New routine assigned!",
      "workout_completed": "Workout completed successfully!",
      "student_added": "Student added successfully!",
      "exercise_created": "Exercise created successfully!",
      "profile_updated": "Profile updated successfully!"
    }
  }
}
```

### **shared/i18n/es.json** (Español)
```json
{
  "version": "1.0.0",
  "locale": "es",
  "translations": {
    "common": {
      "app_name": "Plataforma de Entrenamiento",
      "welcome": "Bienvenido",
      "loading": "Cargando...",
      "error": "Error",
      "success": "Éxito",
      "save": "Guardar",
      "cancel": "Cancelar",
      "delete": "Eliminar",
      "edit": "Editar",
      "create": "Crear",
      "back": "Atrás",
      "next": "Siguiente",
      "previous": "Anterior",
      "confirm": "Confirmar",
      "search": "Buscar",
      "filter": "Filtrar",
      "sort": "Ordenar"
    },
    "auth": {
      "login": "Iniciar Sesión",
      "logout": "Cerrar Sesión",
      "register": "Registrarse",
      "email": "Correo Electrónico",
      "password": "Contraseña",
      "forgot_password": "¿Olvidaste tu contraseña?",
      "reset_password": "Restablecer contraseña",
      "invalid_credentials": "Email o contraseña inválidos",
      "session_expired": "Tu sesión ha expirado"
    },
    "roles": {
      "trainer": "Entrenador Personal",
      "student": "Alumno",
      "admin": "Administrador"
    },
    "trainers": {
      "my_students": "Mis Alumnos",
      "add_student": "Agregar Alumno",
      "student_count": "{{count}} alumnos",
      "no_students": "Aún no tienes alumnos"
    },
    "students": {
      "my_trainer": "Mi Entrenador",
      "current_routine": "Rutina Actual",
      "no_routine": "No tienes rutina activa asignada"
    },
    "exercises": {
      "title": "Ejercicios",
      "create_exercise": "Crear Ejercicio",
      "muscle_groups": "Grupos Musculares",
      "difficulty": "Dificultad",
      "equipment": "Equipamiento",
      "squat": {
        "name": "Sentadilla con Barra",
        "description": "Ejercicio compuesto de tren inferior que trabaja cuádriceps, glúteos e isquiotibiales"
      },
      "bench_press": {
        "name": "Press de Banca",
        "description": "Ejercicio compuesto de tren superior que trabaja pecho, hombros y tríceps"
      },
      "deadlift": {
        "name": "Peso Muerto Convencional",
        "description": "Ejercicio compuesto de cuerpo completo que trabaja la cadena posterior"
      }
    },
    "muscle_groups": {
      "quadriceps": "Cuádriceps",
      "hamstrings": "Isquiotibiales",
      "glutes": "Glúteos",
      "calves": "Gemelos",
      "chest": "Pecho",
      "back": "Espalda",
      "shoulders": "Hombros",
      "biceps": "Bíceps",
      "triceps": "Tríceps",
      "core": "Core",
      "abs": "Abdominales"
    },
    "difficulty": {
      "beginner": "Principiante",
      "intermediate": "Intermedio",
      "advanced": "Avanzado"
    },
    "routines": {
      "title": "Rutinas",
      "create_routine": "Crear Rutina",
      "start_date": "Fecha de Inicio",
      "end_date": "Fecha de Fin",
      "weeks": "Semanas",
      "week": "Semana {{number}}",
      "day": "Día {{number}}",
      "rest_day": "Día de Descanso"
    },
    "days": {
      "monday": "Lunes",
      "tuesday": "Martes",
      "wednesday": "Miércoles",
      "thursday": "Jueves",
      "friday": "Viernes",
      "saturday": "Sábado",
      "sunday": "Domingo"
    },
    "workout": {
      "sets": "Series",
      "reps": "Repeticiones",
      "rest": "Descanso",
      "weight": "Peso",
      "notes": "Notas",
      "log_workout": "Registrar Entrenamiento",
      "mark_complete": "Marcar Completado"
    },
    "workout_log": {
      "title": "Historial de Entrenamientos",
      "today": "Hoy",
      "this_week": "Esta Semana",
      "this_month": "Este Mes",
      "feeling": "¿Cómo te sentiste?",
      "excellent": "Excelente",
      "good": "Bien",
      "average": "Regular",
      "poor": "Mal"
    }
  }
}
```

### **shared/i18n/schema.json** (Validación)
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["version", "locale", "translations"],
  "properties": {
    "version": {
      "type": "string",
      "pattern": "^\\d+\\.\\d+\\.\\d+$"
    },
    "locale": {
      "type": "string",
      "enum": ["en", "es", "pt", "fr", "de"]
    },
    "translations": {
      "type": "object",
      "required": ["common", "auth", "exercises", "muscle_groups", "routines", "workout"],
      "properties": {
        "common": { "type": "object" },
        "auth": { "type": "object" },
        "roles": { "type": "object" },
        "users": { "type": "object" },
        "trainers": { "type": "object" },
        "students": { "type": "object" },
        "exercises": { "type": "object" },
        "muscle_groups": { "type": "object" },
        "difficulty": { "type": "object" },
        "equipment": { "type": "object" },
        "routines": { "type": "object" },
        "days": { "type": "object" },
        "workout": { "type": "object" },
        "workout_log": { "type": "object" },
        "progress": { "type": "object" },
        "validation": { "type": "object" },
        "errors": { "type": "object" },
        "notifications": { "type": "object" }
      }
    }
  }
}
```

---

## Implementación Backend (Go)

### **internal/i18n/loader.go**
```go
package i18n

import (
    "embed"
    "encoding/json"
    "fmt"
    "sync"
)

//go:embed ../../shared/i18n/*.json
var translationFiles embed.FS

type Translations struct {
    Version      string                            `json:"version"`
    Locale       string                            `json:"locale"`
    Translations map[string]map[string]interface{} `json:"translations"`
}

type I18nManager struct {
    translations map[string]*Translations
    mu           sync.RWMutex
}

var (
    manager *I18nManager
    once    sync.Once
)

func GetManager() *I18nManager {
    once.Do(func() {
        manager = &I18nManager{
            translations: make(map[string]*Translations),
        }
        manager.loadTranslations()
    })
    return manager
}

func (m *I18nManager) loadTranslations() error {
    locales := []string{"en", "es", "pt"}
    
    for _, locale := range locales {
        filename := fmt.Sprintf("../../shared/i18n/%s.json", locale)
        data, err := translationFiles.ReadFile(filename)
        if err != nil {
            return fmt.Errorf("failed to read %s: %w", filename, err)
        }
        
        var trans Translations
        if err := json.Unmarshal(data, &trans); err != nil {
            return fmt.Errorf("failed to parse %s: %w", filename, err)
        }
        
        m.mu.Lock()
        m.translations[locale] = &trans
        m.mu.Unlock()
    }
    
    return nil
}

func (m *I18nManager) Get(locale, key string) (string, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    trans, ok := m.translations[locale]
    if !ok {
        trans = m.translations["en"] // fallback to English
    }
    
    value, err := m.getNestedValue(trans.Translations, key)
    if err != nil {
        return "", err
    }
    
    return value, nil
}

func (m *I18nManager) getNestedValue(data map[string]map[string]interface{}, key string) (string, error) {
    // Parse key like "exercises.squat.name"
    // Implementation here...
    return "", nil
}

func (m *I18nManager) ValidateKey(key string) bool {
    // Check if key exists in base locale (en)
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    _, err := m.getNestedValue(m.translations["en"].Translations, key)
    return err == nil
}
```

### **internal/i18n/validator.go**
```go
package i18n

import (
    "fmt"
)

type KeyValidator struct {
    manager *I18nManager
}

func NewKeyValidator() *KeyValidator {
    return &KeyValidator{
        manager: GetManager(),
    }
}

func (v *KeyValidator) ValidateExercise(nameKey, descKey string) error {
    if !v.manager.ValidateKey(nameKey) {
        return fmt.Errorf("invalid i18n key for exercise name: %s", nameKey)
    }
    
    if !v.manager.ValidateKey(descKey) {
        return fmt.Errorf("invalid i18n key for exercise description: %s", descKey)
    }
    
    return nil
}

func (v *KeyValidator) ValidateMuscleGroup(groupKey string) error {
    if !v.manager.ValidateKey(groupKey) {
        return fmt.Errorf("invalid i18n key for muscle group: %s", groupKey)
    }
    return nil
}
```

---

## Implementación Frontend (React)

### **src/i18n/index.ts**
```typescript
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import en from '../../../shared/i18n/en.json';
import es from '../../../shared/i18n/es.json';
import pt from '../../../shared/i18n/pt.json';

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en.translations },
      es: { translation: es.translations },
      pt: { translation: pt.translations },
    },
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false,
    },
  });

export default i18n;
```

### **src/i18n/hooks.ts**
```typescript
import { useTranslation } from 'react-i18next';

export const useI18n = () => {
  const { t, i18n } = useTranslation();
  
  return {
    t,
    locale: i18n.language,
    changeLocale: (locale: string) => i18n.changeLanguage(locale),
  };
};

export const useExerciseTranslation = (exerciseKey: string) => {
  const { t } = useTranslation();
  
  return {
    name: t(`exercises.${exerciseKey}.name`),
    description: t(`exercises.${exerciseKey}.description`),
  };
};
```

### **src/components/ExerciseCard.tsx**
```typescript
import React from 'react';
import { useI18n } from '../i18n/hooks';

interface Exercise {
  id: string;
  nameKey: string;
  descriptionKey: string;
  youtubeUrl: string;
  muscleGroups: Array<{
    groupKey: string;
    impactPercentage: number;
  }>;
}

export const ExerciseCard: React.FC<{ exercise: Exercise }> = ({ exercise }) => {
  const { t } = useI18n();
  
  return (
    <div className="exercise-card">
      <h3>{t(exercise.nameKey)}</h3>
      <p>{t(exercise.descriptionKey)}</p>
      
      <div className="muscle-groups">
        {exercise.muscleGroups.map((mg) => (
          <span key={mg.groupKey}>
            {t(mg.groupKey)} ({mg.impactPercentage}%)
          </span>
        ))}
      </div>
    </div>
  );
};
```

---

## Flujo de Trabajo

### **1. Agregar Nueva Traducción**
```bash
# 1. Editar shared/i18n/en.json (base)
# 2. Editar shared/i18n/es.json
# 3. Editar shared/i18n/pt.json
# 4. Validar con schema
npm run validate-i18n

# 5. Backend recarga automáticamente (embed.FS)
# 6. Frontend rebuild
npm run build
```

### **2. Validación en Backend**
```go
// Al crear un ejercicio
func (s *ExerciseService) Create(req *CreateExerciseRequest) error {
    validator := i18n.NewKeyValidator()
    
    if err := validator.ValidateExercise(req.NameKey, req.DescriptionKey); err != nil {
        return err
    }
    
    // Continue with creation...
}
```

### **3. Uso en Frontend**
```typescript
// Componente
const { t } = useI18n();

<button>{t('common.save')}</button>
<h1>{t('exercises.squat.name')}</h1>
```

---

## Ventajas de esta Arquitectura

1. **Type Safety**: Schema JSON valida estructura
2. **Single Source**: Un solo lugar para traducciones
3. **Validación Backend**: Evita keys inválidos en DB
4. **Hot Reload**: Frontend actualiza sin rebuild
5. **Fallback**: Si falta traducción, usa inglés
6. **Interpolación**: Soporte para variables `{{count}}`
7. **Namespace**: Organización por dominio
8. **Versionado**: Control de cambios en traducciones
