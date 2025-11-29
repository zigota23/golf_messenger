package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/yourusername/golf_messenger/internal/models"
	"github.com/yourusername/golf_messenger/internal/repository"
	"github.com/yourusername/golf_messenger/pkg/jwt"
)

type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtSecret        string
	accessDuration   time.Duration
	refreshDuration  time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtSecret string,
	accessDuration time.Duration,
	refreshDuration time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
		accessDuration:   accessDuration,
		refreshDuration:  refreshDuration,
	}
}

func (s *AuthService) Register(email, password, firstName, lastName string) (*models.User, *jwt.TokenPair, error) {
	existingUser, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, nil, errors.New("user with this email already exists")
	}

	user := &models.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}

	if err := user.SetPassword(password); err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	tokenPair, err := s.createTokenPair(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tokens: %w", err)
	}

	return user, tokenPair, nil
}

func (s *AuthService) Login(email, password string) (*models.User, *jwt.TokenPair, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, nil, errors.New("invalid email or password")
	}

	if !user.CheckPassword(password) {
		return nil, nil, errors.New("invalid email or password")
	}

	tokenPair, err := s.createTokenPair(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tokens: %w", err)
	}

	return user, tokenPair, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*jwt.TokenPair, error) {
	tokenHash := jwt.HashRefreshToken(refreshToken)

	storedToken, err := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}
	if storedToken == nil {
		return nil, errors.New("invalid refresh token")
	}

	if !storedToken.IsValid() {
		return nil, errors.New("refresh token is invalid or expired")
	}

	if err := s.refreshTokenRepo.RevokeByUserID(storedToken.UserID); err != nil {
		return nil, fmt.Errorf("failed to revoke old tokens: %w", err)
	}

	tokenPair, err := s.createTokenPair(storedToken.User)
	if err != nil {
		return nil, fmt.Errorf("failed to create new tokens: %w", err)
	}

	return tokenPair, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	tokenHash := jwt.HashRefreshToken(refreshToken)

	storedToken, err := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return fmt.Errorf("failed to find refresh token: %w", err)
	}
	if storedToken == nil {
		return errors.New("invalid refresh token")
	}

	if err := s.refreshTokenRepo.RevokeByUserID(storedToken.UserID); err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	return nil
}

func (s *AuthService) createTokenPair(user *models.User) (*jwt.TokenPair, error) {
	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Email, s.jwtSecret, s.accessDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshTokenData, err := jwt.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := time.Now().Add(s.refreshDuration)
	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshTokenData.Hash,
		ExpiresAt: expiresAt,
		Revoked:   false,
	}

	if err := s.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &jwt.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenData.Token,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}
