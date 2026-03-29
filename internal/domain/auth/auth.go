package auth

import (
	"time"
)

// UserCredentials representa las credenciales de un usuario
type UserCredentials struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	Password  string `json:"password"` // Hasheado con bcrypt
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

// AuthRequest representa una solicitud de autenticación
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse representa la respuesta de autenticación
type AuthResponse struct {
	User         *UserInfo `json:"user"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresIn    int64     `json:"expiresIn"`
}

// RefreshRequest representa una solicitud de refresh token
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// UserInfo representa información pública del usuario
type UserInfo struct {
	ID       string      `json:"id"`
	Email    string      `json:"email"`
	Name     string      `json:"name"`
	Role     string      `json:"role"`
	Status   string      `json:"status"`
	Metadata interface{} `json:"metadata,omitempty"`
}

// TokenPair representa un par de tokens JWT
type TokenPair struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// TokenClaims representa los claims del JWT
type TokenClaims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Iss    int64  `json:"iat"`
	Exp    int64  `json:"exp"`
}
