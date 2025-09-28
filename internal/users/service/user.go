package service

import (
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/merdernoty/job-hunter/internal/users/domain"
	"github.com/merdernoty/job-hunter/pkg/logger"
	"github.com/merdernoty/job-hunter/pkg/telegram"
)

type userService struct {
	userRepo      domain.UserRepository
	telegramAuth  *telegram.TelegramAuth
	avatarService *AvatarService
	logger        logger.Logger
}

func NewUserService(
	userRepo domain.UserRepository,
	telegramAuth *telegram.TelegramAuth,
	avatarService *AvatarService,
	logger logger.Logger,
) domain.UserService {
	return &userService{
		userRepo:      userRepo,
		telegramAuth:  telegramAuth,
		avatarService: avatarService,
		logger:        logger,
	}
}

func (s *userService) AuthFromTelegram(initData string) (*domain.User, string, error) {
	webAppData, err := s.telegramAuth.ValidateWebAppData(initData)
	if err != nil {
		s.logger.Errorf("Invalid telegram data: %v", err)
		return nil, "", fmt.Errorf("invalid telegram data")
	}

	user, err := s.userRepo.GetByTelegramID(webAppData.User.ID)
	if err != nil && err.Error() == "user not found" {
		user = &domain.User{
			ID:         uuid.New(),
			TelegramID: webAppData.User.ID,
			Username:   webAppData.User.Username,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := s.userRepo.Create(user); err != nil {
			s.logger.Errorf("Failed to create user: %v", err)
			return nil, "", fmt.Errorf("failed to create user")
		}

		s.logger.Infof("Created new user from Telegram: %s (%d)", user.Username, user.TelegramID)
	} else if err != nil {
		s.logger.Errorf("Database error getting user: %v", err)
		return nil, "", fmt.Errorf("database error")
	}

	token := fmt.Sprintf("user_%s_%d", user.ID.String(), time.Now().Unix())

	s.logger.Infof("User authenticated: %s (%d)", user.Username, user.TelegramID)
	return user, token, nil
}

func (s *userService) GetUser(id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		s.logger.Errorf("Failed to get user %s: %v", id, err)
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateUser(id uuid.UUID, req domain.UpdateUserRequest) (*domain.User, error) {
	existingUser, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Update(id, req); err != nil {
		s.logger.Errorf("Failed to update user %s: %v", id, err)
		return nil, fmt.Errorf("failed to update user")
	}

	updatedUser, err := s.userRepo.GetByID(id)
	if err != nil {
		s.logger.Errorf("Failed to get updated user %s: %v", id, err)
		return existingUser, nil
	}

	s.logger.Infof("Updated user: %s", id)
	return updatedUser, nil
}

func (s *userService) UpdateUserAvatar(userID uuid.UUID, file io.Reader, fileName string, fileSize int64, contentType string) (string, error) {
	existingUser, err := s.userRepo.GetByID(userID)
	if err != nil {
		return "", err
	}

	if existingUser.AvatarURL != "" {
		if err := s.avatarService.DeleteAvatar(existingUser.AvatarURL); err != nil {
			s.logger.Warnf("Failed to delete old avatar for user %s: %v", userID, err)
		}
	}

	uploadReq := UploadAvatarRequest{
		UserID:      userID,
		File:        file,
		FileName:    fileName,
		FileSize:    fileSize,
		ContentType: contentType,
	}

	avatarURL, err := s.avatarService.UploadAvatar(uploadReq)
	if err != nil {
		s.logger.Errorf("Failed to upload avatar for user %s: %v", userID, err)
		return "", err
	}

	updateReq := domain.UpdateUserRequest{
		AvatarURL: &avatarURL,
	}

	updatedUser, err := s.UpdateUser(userID, updateReq)
	if err != nil {
		if delErr := s.avatarService.DeleteAvatar(avatarURL); delErr != nil {
			s.logger.Errorf("Failed to cleanup avatar after DB error: %v", delErr)
		}
		return "", fmt.Errorf("failed to update user avatar in database: %w", err)
	}

	s.logger.Infof("Successfully updated avatar for user %s: %s", userID, avatarURL)
	return updatedUser.AvatarURL, nil
}

func (s *userService) DeleteUserAvatar(userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if user.AvatarURL == "" {
		return fmt.Errorf("user has no avatar")
	}

	if err := s.avatarService.DeleteAvatar(user.AvatarURL); err != nil {
		s.logger.Errorf("Failed to delete avatar file: %v", err)
	}

	emptyURL := ""
	updateReq := domain.UpdateUserRequest{
		AvatarURL: &emptyURL,
	}

	if _, err := s.UpdateUser(userID, updateReq); err != nil {
		return fmt.Errorf("failed to update user avatar in database: %w", err)
	}

	s.logger.Infof("Successfully deleted avatar for user %s", userID)
	return nil
}
