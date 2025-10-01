package service

import (
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/merdernoty/job-hunter/internal/users/domain"
	"github.com/merdernoty/job-hunter/pkg/jwt"
	"github.com/merdernoty/job-hunter/pkg/logger"
	"github.com/merdernoty/job-hunter/pkg/telegram"
)

type userService struct {
	userRepo      domain.UserRepository
	jwtService    *jwt.JWTService
	dailyViewRepo domain.UserDailyViewRepository
	telegramAuth  *telegram.TelegramAuth
	avatarService *AvatarService
	logger        logger.Logger
}

func NewUserService(
	userRepo domain.UserRepository,
	telegramAuth *telegram.TelegramAuth,
	jwtService *jwt.JWTService,
	dailyViewRepo domain.UserDailyViewRepository,
	avatarService *AvatarService,
	logger logger.Logger,
) domain.UserService {
	return &userService{
		userRepo:      userRepo,
		jwtService:    jwtService,
		telegramAuth:  telegramAuth,
		avatarService: avatarService,
		dailyViewRepo: dailyViewRepo,
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
		username := webAppData.User.Username
		handle := ""
		if username != "" {
			handle = "@" + username
		} else {
			handle = fmt.Sprintf("id%d", webAppData.User.ID)
		}

		user = &domain.User{
			ID:             uuid.New(),
			TelegramID:     webAppData.User.ID,
			Username:       webAppData.User.Username,
			TelegramHandle: handle,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
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

	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed generate jwt token: %w", err)
	}

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

	if existingUser.AvatarURL != nil && *existingUser.AvatarURL != "" {
		if err := s.avatarService.DeleteAvatar(*existingUser.AvatarURL); err != nil {
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
	return *updatedUser.AvatarURL, nil
}

func (s *userService) DeleteUserAvatar(userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if user.AvatarURL == nil || *user.AvatarURL == "" {
		return fmt.Errorf("user has no avatar")
	}

	if err := s.avatarService.DeleteAvatar(*user.AvatarURL); err != nil {
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
func (s *userService) GetRandomUser(viewerID uuid.UUID) (*domain.User, error) {
	shownToday, err := s.dailyViewRepo.GetTodaysShownUsers(viewerID)
	if err != nil {
		s.logger.Warnf("Failed to get today's shown users: %v", err)
		shownToday = []uuid.UUID{}
	}

	excludeMap := make(map[uuid.UUID]bool)
	excludeMap[viewerID] = true

	for _, userID := range shownToday {
		excludeMap[userID] = true
	}

	excludeUserIDs := make([]uuid.UUID, 0, len(excludeMap))
	for userID := range excludeMap {
		excludeUserIDs = append(excludeUserIDs, userID)
	}

	s.logger.Infof("Excluding %d unique users for viewer %s", len(excludeUserIDs)-1, viewerID)

	user, err := s.userRepo.GetRandomUser(excludeUserIDs)
	if err != nil {
		if err.Error() == "no available users found" {
			s.logger.Infof("All users shown to viewer %s today - no more users available", viewerID)
			return nil, fmt.Errorf("no more users available today")
		} else {
			s.logger.Errorf("Database error getting random user: %v", err)
			return nil, fmt.Errorf("database error")
		}
	}

	dailyView := &domain.UserDailyView{
		ViewerID:    viewerID,
		ShownUserID: user.ID,
		ViewDate:    time.Now().UTC().Truncate(24 * time.Hour),
		CreatedAt:   time.Now(),
	}

	if err := s.dailyViewRepo.Create(dailyView); err != nil {
		s.logger.Warnf("Failed to create daily view record: %v", err)
	}

	s.logger.Infof("Selected user %s (%s) for viewer %s", user.Username, user.ID, viewerID)
	return user, nil
}

func (s *userService) GetAllUsers() ([]domain.User, error) {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		s.logger.Warnf("Failed to get users %v", err)
		return nil, fmt.Errorf("database error %w", err)
	}

	return users, nil
}
