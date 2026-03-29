package user

import "time"

type Trainer struct {
	User
	Metadata TrainerMetadata `json:"metadata"`
}

type TrainerMetadata struct {
	Specializations []string `json:"specializations"`
	Certifications  []string `json:"certifications"`
	Bio             string   `json:"bio"`
	YearsExperience int      `json:"yearsExperience"`
}

type TrainerStudent struct {
	TrainerID    string    `json:"trainerId"`
	StudentID    string    `json:"studentId"`
	StudentName  string    `json:"studentName"`
	StudentEmail string    `json:"studentEmail"`
	AssignedAt   time.Time `json:"assignedAt"`
	Status       UserStatus `json:"status"`
}
