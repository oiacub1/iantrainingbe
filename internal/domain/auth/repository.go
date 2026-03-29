package auth

import "context"

// Repository define la interfaz para el repositorio de autenticación
type Repository interface {
	// Credenciales
	CreateCredentials(ctx context.Context, credentials *UserCredentials) error
	GetCredentialsByEmail(ctx context.Context, email string) (*UserCredentials, error)
	UpdateCredentials(ctx context.Context, credentials *UserCredentials) error
	DeleteCredentials(ctx context.Context, userID string) error
	
	// Refresh Tokens
	StoreRefreshToken(ctx context.Context, userID, tokenID string, expiresAt int64) error
	GetRefreshToken(ctx context.Context, tokenID string) (string, error)
	RevokeRefreshToken(ctx context.Context, tokenID string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
}
