package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/merdernoty/job-hunter/internal/users/domain"
	"github.com/merdernoty/job-hunter/pkg/logger"
	"github.com/merdernoty/job-hunter/pkg/telegram"
)

type userService struct {
	userRepo     domain.UserRepository
	telegramAuth *telegram.TelegramAuth
	logger       logger.Logger
}

func NewUserService(
	userRepo domain.UserRepository,
	telegramAuth *telegram.TelegramAuth,
	logger logger.Logger,
) domain.UserService {
	return &userService{
		userRepo:     userRepo,
		telegramAuth: telegramAuth,
		logger:       logger,
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
		
		s.logger.Infof("Created new user from Telegram: %s (%d)", user.TelegramID)
	} else if err != nil {
		s.logger.Errorf("Database error getting user: %v", err)
		return nil, "", fmt.Errorf("database error")
	}
	
	token := fmt.Sprintf("user_%s_%d", user.ID.String(), time.Now().Unix()) //TODO JWT
	
	s.logger.Infof("User authenticated: %s (%d)", user.TelegramID)
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