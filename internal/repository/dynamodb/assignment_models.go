package dynamodb

import (
	"time"

	"iantraining/internal/domain/routine"
)

type DynamoRoutineAssignment struct {
	PK         string    `dynamodbav:"PK"`
	SK         string    `dynamodbav:"SK"`
	ID         string    `dynamodbav:"id"`
	RoutineID  string    `dynamodbav:"routineId"`
	StudentID  string    `dynamodbav:"studentId"`
	StartDate  string    `dynamodbav:"startDate"`
	EndDate    string    `dynamodbav:"endDate"`
	Status     string    `dynamodbav:"status"`
	CreatedAt  time.Time `dynamodbav:"createdAt"`
	UpdatedAt  time.Time `dynamodbav:"updatedAt"`
	EntityType string    `dynamodbav:"entityType"`
	GSI1PK     string    `dynamodbav:"GSI1PK"`
	GSI1SK     string    `dynamodbav:"GSI1SK"`
	GSI2PK     string    `dynamodbav:"GSI2PK"`
	GSI2SK     string    `dynamodbav:"GSI2SK"`
}

func assignmentToDynamoItem(a *routine.RoutineAssignment) *DynamoRoutineAssignment {
	startDate := a.StartDate.Format("2006-01-02")
	endDate := a.EndDate.Format("2006-01-02")

	return &DynamoRoutineAssignment{
		PK:         "STUDENT#" + a.StudentID,
		SK:         "ASSIGNMENT#" + a.ID,
		ID:         a.ID,
		RoutineID:  a.RoutineID,
		StudentID:  a.StudentID,
		StartDate:  startDate,
		EndDate:    endDate,
		Status:     string(a.Status),
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
		EntityType: EntityTypeRoutineAssignment,
		GSI1PK:     "ROUTINE#" + a.RoutineID,
		GSI1SK:     "ASSIGNMENT#" + a.StudentID,
		GSI2PK:     "STUDENT#" + a.StudentID + "#STATUS#" + string(a.Status),
		GSI2SK:     "ASSIGNMENT#" + startDate,
	}
}

func dynamoItemToAssignment(d *DynamoRoutineAssignment) *routine.RoutineAssignment {
	startDate, _ := time.Parse("2006-01-02", d.StartDate)
	endDate, _ := time.Parse("2006-01-02", d.EndDate)

	return &routine.RoutineAssignment{
		ID:        d.ID,
		RoutineID: d.RoutineID,
		StudentID: d.StudentID,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    routine.AssignmentStatus(d.Status),
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}
