package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	authDomain "iantraining/internal/domain/auth"
	userDomain "iantraining/internal/domain/user"
)

// Service maneja la lógica de negocio de autenticación
type Service struct {
	authRepo     authDomain.Repository
	userRepo     userDomain.Repository
	jwtService   *JWTService
	passwordCost int
}

// NewService crea un nuevo servicio de autenticación
func NewService(
	authRepo authDomain.Repository,
	userRepo userDomain.Repository,
	jwtService *JWTService,
) *Service {
	return &Service{
		authRepo:     authRepo,
		userRepo:     userRepo,
		jwtService:   jwtService,
		passwordCost: bcrypt.DefaultCost,
	}
}

func (s *Service) ResolveUserIDByEmail(ctx context.Context, email string) (string, error) {
	credentials, err := s.authRepo.GetCredentialsByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	return credentials.UserID, nil
}

// Login autentica un usuario y retorna tokens
func (s *Service) Login(ctx context.Context, req *authDomain.AuthRequest) (*authDomain.AuthResponse, error) {
	// 1. Obtener credenciales por email
	credentials, err := s.authRepo.GetCredentialsByEmail(ctx, req.Email)
	if err != nil {
		if err == authDomain.ErrUserNotFound {
			return nil, authDomain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	// 2. Verificar password
	if err := bcrypt.CompareHashAndPassword([]byte(credentials.Password), []byte(req.Password)); err != nil {
		return nil, authDomain.ErrInvalidCredentials
	}

	// 3. Obtener información del usuario
	user, err := s.getUserByID(ctx, credentials.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// 4. Verificar que el usuario esté activo
	if user.Status != userDomain.StatusActive {
		return nil, authDomain.ErrUserInactive
	}

	// 5. Generar tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 6. Guardar refresh token
	tokenID, err := s.jwtService.GetTokenID(tokenPair.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to extract token ID: %w", err)
	}

	if err := s.authRepo.StoreRefreshToken(ctx, user.ID, tokenID, tokenPair.ExpiresAt.Unix()); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// 7. Construir respuesta
	userInfo := &authDomain.UserInfo{
		ID:     user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   string(user.Role),
		Status: string(user.Status),
	}

	// Agregar metadata según el rol
	switch user.Role {
	case userDomain.RoleTrainer:
		if trainer, err := s.userRepo.GetTrainer(ctx, user.ID); err == nil {
			userInfo.Metadata = trainer.Metadata
		}
	case userDomain.RoleStudent:
		if student, err := s.userRepo.GetStudent(ctx, user.ID); err == nil {
			userInfo.Metadata = student.Metadata
		}
	}

	return &authDomain.AuthResponse{
		User:         userInfo,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int64(s.jwtService.accessTokenTTL.Seconds()),
	}, nil
}

// RefreshToken refresca un access token usando un refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*authDomain.TokenPair, error) {
	// 1. Validar refresh token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 2. Verificar que el refresh token exista en la BD
	tokenID, err := s.jwtService.GetTokenID(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to extract token ID: %w", err)
	}

	userID, err := s.authRepo.GetRefreshToken(ctx, tokenID)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found or revoked: %w", err)
	}

	// 3. Verificar que el userID coincida
	if userID != claims.UserID {
		// Token comprometido - revocar todos los tokens del usuario
		s.authRepo.RevokeAllUserTokens(ctx, userID)
		return nil, authDomain.ErrInvalidRefresh
	}

	// 4. Obtener información actualizada del usuario
	user, err := s.getUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// 5. Verificar que el usuario esté activo
	if user.Status != userDomain.StatusActive {
		return nil, authDomain.ErrUserInactive
	}

	// 6. Generar nuevos tokens
	newTokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// 7. Guardar nuevo refresh token
	newTokenID, err := s.jwtService.GetTokenID(newTokenPair.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to extract new token ID: %w", err)
	}

	if err := s.authRepo.StoreRefreshToken(ctx, user.ID, newTokenID, newTokenPair.ExpiresAt.Unix()); err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// 8. Revocar el refresh token anterior
	s.authRepo.RevokeRefreshToken(ctx, tokenID)

	return newTokenPair, nil
}

// Logout revoca los tokens de un usuario
func (s *Service) Logout(ctx context.Context, userID string) error {
	return s.authRepo.RevokeAllUserTokens(ctx, userID)
}

// LogoutFromToken revoca los tokens basados en un token específico
func (s *Service) LogoutFromToken(ctx context.Context, accessToken string) error {
	claims, err := s.jwtService.ValidateAccessToken(accessToken)
	if err != nil {
		return fmt.Errorf("invalid access token: %w", err)
	}

	return s.authRepo.RevokeAllUserTokens(ctx, claims.UserID)
}

// CreateCredentials crea credenciales para un nuevo usuario
func (s *Service) CreateCredentials(ctx context.Context, userID, email, password string) error {
	// Hashear password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.passwordCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	credentials := &authDomain.UserCredentials{
		UserID:    userID,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	return s.authRepo.CreateCredentials(ctx, credentials)
}

// UpdatePassword actualiza el password de un usuario
func (s *Service) UpdatePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// 1. Obtener credenciales actuales
	credentials, err := s.authRepo.GetCredentialsByEmail(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	// 2. Verificar password actual
	if err := bcrypt.CompareHashAndPassword([]byte(credentials.Password), []byte(currentPassword)); err != nil {
		return authDomain.ErrInvalidCredentials
	}

	// 3. Hashear nuevo password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.passwordCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 4. Actualizar credenciales
	credentials.Password = string(hashedPassword)
	credentials.UpdatedAt = time.Now().Unix()

	if err := s.authRepo.UpdateCredentials(ctx, credentials); err != nil {
		return fmt.Errorf("failed to update credentials: %w", err)
	}

	// 5. Revocar todos los tokens (forzar re-login)
	s.authRepo.RevokeAllUserTokens(ctx, userID)

	return nil
}

// ValidateToken valida un access token y retorna los claims
func (s *Service) ValidateToken(ctx context.Context, accessToken string) (*authDomain.TokenClaims, error) {
	return s.jwtService.ValidateAccessToken(accessToken)
}

// getUserByID obtiene información de usuario por ID, manejando diferentes tipos
func (s *Service) getUserByID(ctx context.Context, userID string) (*userDomain.User, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}

// HashPassword hashea un password (util para testing/seeding)
func (s *Service) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.passwordCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// VerifyPassword verifica un password contra su hash
func (s *Service) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
