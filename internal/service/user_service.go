package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/yourusername/golf_messenger/internal/models"
	"github.com/yourusername/golf_messenger/internal/repository"
	"github.com/yourusername/golf_messenger/pkg/storage"
)

type UserService struct {
	userRepo repository.UserRepository
	s3Client *storage.S3Client
}

func NewUserService(userRepo repository.UserRepository, s3Client *storage.S3Client) *UserService {
	return &UserService{
		userRepo: userRepo,
		s3Client: s3Client,
	}
}

func (s *UserService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserService) UpdateProfile(userID uuid.UUID, firstName, lastName string, handicap *float64, phone *string) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if firstName != "" {
		user.FirstName = firstName
	}
	if lastName != "" {
		user.LastName = lastName
	}
	if handicap != nil {
		user.Handicap = handicap
	}
	if phone != nil {
		user.Phone = phone
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *UserService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	if !user.CheckPassword(oldPassword) {
		return errors.New("invalid old password")
	}

	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to set new password: %w", err)
	}

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *UserService) UploadAvatar(ctx context.Context, userID uuid.UUID, file io.Reader, filename string, contentType string) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if user.AvatarURL != nil && *user.AvatarURL != "" {
		if err := s.s3Client.DeleteFile(ctx, *user.AvatarURL); err != nil {
			return nil, fmt.Errorf("failed to delete old avatar: %w", err)
		}
	}

	avatarURL, err := s.s3Client.UploadFile(ctx, file, filename, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	user.AvatarURL = &avatarURL

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user with avatar URL: %w", err)
	}

	return user, nil
}

func (s *UserService) DeleteAvatar(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if user.AvatarURL != nil && *user.AvatarURL != "" {
		if err := s.s3Client.DeleteFile(ctx, *user.AvatarURL); err != nil {
			return nil, fmt.Errorf("failed to delete avatar from S3: %w", err)
		}
	}

	user.AvatarURL = nil

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *UserService) SearchUsers(query string, limit, offset int) ([]*models.User, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []*models.User{}, nil
	}

	users, err := s.userRepo.Search(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}

func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}
