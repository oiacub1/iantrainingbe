package user

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidUserID          = errors.New("invalid user ID")
	ErrInvalidEmail           = errors.New("invalid email address")
	ErrInvalidRole            = errors.New("invalid user role")
	ErrInvalidStatus          = errors.New("invalid user status")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrTrainerNotFound        = errors.New("trainer not found")
	ErrStudentNotFound        = errors.New("student not found")
	ErrStudentAlreadyAssigned = errors.New("student already assigned to this trainer")
	ErrCannotAssignSelf       = errors.New("cannot assign user to themselves")
	ErrNameRequired           = errors.New("name is required")
	ErrPhoneInvalid           = errors.New("invalid phone number")
)
