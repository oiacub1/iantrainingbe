package routine

// DEPRECATED: Este archivo contiene la lógica antigua de asignación de rutinas.
// La nueva funcionalidad de asignación de rutinas a múltiples estudiantes
// se encuentra en el servicio de asignaciones (assignment service).
//
// Ver: backend/internal/service/assignment/service.go
// Ver: backend/internal/domain/routine/assignment.go
//
// El nuevo modelo separa las rutinas (plantillas) de las asignaciones:
// - Routine: Plantilla de rutina creada por el trainer
// - RoutineAssignment: Asignación de una rutina a un estudiante específico
//
// Esto permite:
// - Asignar una rutina a múltiples estudiantes
// - Fechas de inicio/fin independientes por estudiante
// - Progreso individual por estudiante
// - Estados independientes (activo, completado, pausado, cancelado)
