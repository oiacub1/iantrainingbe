package auth

import "fmt"

// ErrorType representa los tipos de error de autenticación
type ErrorType string

const (
	ErrorInvalidCredentials ErrorType = "INVALID_CREDENTIALS"
	ErrorUserNotFound       ErrorType = "USER_NOT_FOUND"
	ErrorUserInactive       ErrorType = "USER_INACTIVE"
	ErrorTokenExpired       ErrorType = "TOKEN_EXPIRED"
	ErrorTokenInvalid       ErrorType = "TOKEN_INVALID"
	ErrorInvalidRefresh     ErrorType = "INVALID_REFRESH_TOKEN"
	ErrorInternal           ErrorType = "INTERNAL_ERROR"
)

// AuthError representa un error de autenticación
type AuthError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("AuthError[%s]: %s", e.Type, e.Message)
}

// NewAuthError crea un nuevo error de autenticación
func NewAuthError(errorType ErrorType, message string) *AuthError {
	return &AuthError{
		Type:    errorType,
		Message: message,
	}
}

// Errores predefinidos
var (
	ErrInvalidCredentials = NewAuthError(ErrorInvalidCredentials, "Invalid email or password")
	ErrUserNotFound       = NewAuthError(ErrorUserNotFound, "User not found")
	ErrUserInactive       = NewAuthError(ErrorUserInactive, "User account is inactive")
	ErrTokenExpired       = NewAuthError(ErrorTokenExpired, "Token has expired")
	ErrTokenInvalid       = NewAuthError(ErrorTokenInvalid, "Invalid token")
	ErrInvalidRefresh     = NewAuthError(ErrorInvalidRefresh, "Invalid refresh token")
)
