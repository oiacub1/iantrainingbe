package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	authDomain "iantraining/internal/domain/auth"
)

// JWTService maneja la generación y validación de tokens JWT
type JWTService struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

// NewJWTService crea un nuevo servicio JWT
func NewJWTService(secretKey string, accessTokenTTL, refreshTokenTTL time.Duration, issuer string) *JWTService {
	return &JWTService{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		issuer:          issuer,
	}
}

// GenerateTokenPair genera un par de tokens (access y refresh)
func (s *JWTService) GenerateTokenPair(userID, email, role string) (*authDomain.TokenPair, error) {
	now := time.Now()

	// Generar Access Token
	accessTokenClaims := jwt.MapClaims{
		"userId": userID,
		"email":  email,
		"role":   role,
		"iat":    now.Unix(),
		"exp":    now.Add(s.accessTokenTTL).Unix(),
		"iss":    s.issuer,
		"type":   "access",
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generar Refresh Token (más largo y sin claims sensibles)
	refreshTokenClaims := jwt.MapClaims{
		"userId": userID,
		"iat":    now.Unix(),
		"exp":    now.Add(s.refreshTokenTTL).Unix(),
		"iss":    s.issuer,
		"type":   "refresh",
		"jti":    s.generateJTI(), // Unique ID para refresh token
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &authDomain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(s.accessTokenTTL),
	}, nil
}

// ValidateAccessToken valida un access token y retorna los claims
func (s *JWTService) ValidateAccessToken(tokenString string) (*authDomain.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, authDomain.ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, authDomain.ErrTokenInvalid
	}

	// Verificar que sea un access token
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return nil, authDomain.ErrTokenInvalid
	}

	// Extraer claims
	userID, ok := claims["userId"].(string)
	if !ok {
		return nil, authDomain.ErrTokenInvalid
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, authDomain.ErrTokenInvalid
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, authDomain.ErrTokenInvalid
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, authDomain.ErrTokenInvalid
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, authDomain.ErrTokenInvalid
	}

	return &authDomain.TokenClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Iss:    int64(iat),
		Exp:    int64(exp),
	}, nil
}

// ValidateRefreshToken valida un refresh token y retorna los claims
func (s *JWTService) ValidateRefreshToken(tokenString string) (*authDomain.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, authDomain.ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, authDomain.ErrTokenInvalid
	}

	// Verificar que sea un refresh token
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, authDomain.ErrTokenInvalid
	}

	// Extraer claims
	userID, ok := claims["userId"].(string)
	if !ok {
		return nil, authDomain.ErrTokenInvalid
	}

	email, ok := claims["email"].(string)
	if !ok {
		email = "" // Refresh tokens pueden no tener email
	}

	role, ok := claims["role"].(string)
	if !ok {
		role = "" // Refresh tokens pueden no tener role
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, authDomain.ErrTokenInvalid
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, authDomain.ErrTokenInvalid
	}

	return &authDomain.TokenClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Iss:    int64(iat),
		Exp:    int64(exp),
	}, nil
}

// GetTokenID extrae el JTI (JWT ID) de un refresh token
func (s *JWTService) GetTokenID(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return "", authDomain.ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", authDomain.ErrTokenInvalid
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return "", authDomain.ErrTokenInvalid
	}

	return jti, nil
}

// generateJTI genera un JWT ID único
func (s *JWTService) generateJTI() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
