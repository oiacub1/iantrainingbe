package main

import "time"

type AssignRoutineRequest struct {
	RoutineID string `json:"routineId"`
	StudentID string `json:"studentId"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
}

type AssignRoutineResponse struct {
	ID        string    `json:"id"`
	RoutineID string    `json:"routineId"`
	StudentID string    `json:"studentId"`
	TrainerID string    `json:"trainerId"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
