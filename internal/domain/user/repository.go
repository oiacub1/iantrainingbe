package user

import "context"

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userID string) error
	
	CreateTrainer(ctx context.Context, trainer *Trainer) error
	GetTrainer(ctx context.Context, trainerID string) (*Trainer, error)
	
	CreateStudent(ctx context.Context, student *Student) error
	GetStudent(ctx context.Context, studentID string) (*Student, error)
	UpdateStudent(ctx context.Context, student *Student) error
	
	AssignStudentToTrainer(ctx context.Context, relation *TrainerStudent) error
	ListStudentsByTrainer(ctx context.Context, trainerID string, limit int, lastKey string) ([]*TrainerStudent, string, error)
	GetTrainerByStudent(ctx context.Context, studentID string) (*Trainer, error)
	RemoveStudentFromTrainer(ctx context.Context, trainerID, studentID string) error
}
