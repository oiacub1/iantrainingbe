package routine

import "time"

type AssignmentStatus string

const (
	AssignmentStatusActive    AssignmentStatus = "ACTIVE"
	AssignmentStatusCompleted AssignmentStatus = "COMPLETED"
	AssignmentStatusPaused    AssignmentStatus = "PAUSED"
	AssignmentStatusCancelled AssignmentStatus = "CANCELLED"
)

type RoutineAssignment struct {
	ID         string           `json:"id"`
	RoutineID  string           `json:"routineId"`
	StudentID  string           `json:"studentId"`
	StartDate  time.Time        `json:"startDate"`
	EndDate    time.Time        `json:"endDate"`
	Status     AssignmentStatus `json:"status"`
	CreatedAt  time.Time        `json:"createdAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`
}

type CreateAssignmentRequest struct {
	RoutineID  string   `json:"routineId"`
	StudentIDs []string `json:"studentIds"`
	StartDate  string   `json:"startDate"`
	EndDate    string   `json:"endDate"`
}

type UpdateAssignmentRequest struct {
	StartDate string           `json:"startDate"`
	EndDate   string           `json:"endDate"`
	Status    AssignmentStatus `json:"status"`
}

func (s AssignmentStatus) IsValid() bool {
	switch s {
	case AssignmentStatusActive, AssignmentStatusCompleted, AssignmentStatusPaused, AssignmentStatusCancelled:
		return true
	default:
		return false
	}
}
