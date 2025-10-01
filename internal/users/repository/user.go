package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/merdernoty/job-hunter/internal/users/domain"
	"github.com/merdernoty/job-hunter/pkg/logger"
)

type userRepository struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewUserRepository(db *sqlx.DB, logger logger.Logger) domain.UserRepository {
	return &userRepository{db: db, logger: logger}
}

func (r *userRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT id, telegram_id, avatar_url, telegram_handle, username, bio, created_at, updated_at
		FROM users 
		WHERE id = $1`
	err := r.db.Get(&user, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		r.logger.Errorf("Failed to get user by ID %s: %v", id, err)
		return nil, fmt.Errorf("database error")
	}
	return &user, nil
}

func (r *userRepository) GetByTelegramID(telegramID int64) (*domain.User, error) {
	var user domain.User

	query := `
		SELECT id, telegram_id, avatar_url, telegram_handle, username, bio, created_at, updated_at
		FROM users 
		WHERE telegram_id = $1`

	err := r.db.Get(&user, query, telegramID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		r.logger.Errorf("Failed to get user by telegram_id %d: %v", telegramID, err)
		return nil, fmt.Errorf("database error")
	}

	return &user, nil
}

func (r *userRepository) Create(user *domain.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	query := `
		INSERT INTO users (id, telegram_id, username, telegram_handle, avatar_url, bio)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	err := r.db.QueryRow(
		query,
		user.ID, user.TelegramID, user.Username, user.TelegramHandle, user.AvatarURL, user.Bio,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		r.logger.Errorf("Failed to create user: %v", err)
		return fmt.Errorf("failed to create user")
	}

	r.logger.Infof("Created user: %s (telegram_id: %d)", user.Username, user.TelegramID)
	return nil
}

func (r *userRepository) Update(id uuid.UUID, updates domain.UpdateUserRequest) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, *updates.Username)
		argIndex++
	}
	if updates.AvatarURL != nil {
		setParts = append(setParts, fmt.Sprintf("avatar_url = $%d", argIndex))
		args = append(args, *updates.AvatarURL)
		argIndex++
	}
	if updates.Bio != nil {
		setParts = append(setParts, fmt.Sprintf("bio = $%d", argIndex))
		args = append(args, *updates.Bio)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE users 
		SET %s 
		WHERE id = $%d`,
		strings.Join(setParts, ", "), argIndex)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		r.logger.Errorf("Failed to update user %s: %v", id, err)
		return fmt.Errorf("failed to update user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("database error")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Infof("Updated user: %s", id)
	return nil
}

func (r *userRepository) GetRandomUser(excludeUserIDs []uuid.UUID) (*domain.User, error) {
	var user domain.User
	var query string
	var args []interface{}

	r.logger.Infof("GetRandomUser called with %d exclusions: %v", len(excludeUserIDs), excludeUserIDs)

	if len(excludeUserIDs) == 0 {
		query = `
			SELECT id, telegram_id, username, telegram_handle, avatar_url, bio, created_at, updated_at
			FROM users
			ORDER BY RANDOM()
			LIMIT 1`
		r.logger.Info("Using query without exclusions")
	} else {
		placeholders := ""
		for i := range excludeUserIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += fmt.Sprintf("$%d", i+1)
			args = append(args, excludeUserIDs[i])
		}

		query = fmt.Sprintf(`
			SELECT id, telegram_id, username, telegram_handle, avatar_url, bio, created_at, updated_at
			FROM users
			WHERE id NOT IN (%s)
			ORDER BY RANDOM()
			LIMIT 1`, placeholders)

		r.logger.Infof("Using query with exclusions: %s", query)
		r.logger.Infof("Query args: %v", args)
	}

	err := r.db.Get(&user, query, args...)
	if err == sql.ErrNoRows {
		r.logger.Warn("No available users found after exclusions")
		return nil, fmt.Errorf("no available users found")
	}
	if err != nil {
		r.logger.Errorf("Failed to get random user: %v", err)
		return nil, fmt.Errorf("database error")
	}

	r.logger.Infof("Retrieved random user: %s (ID: %s)", user.Username, user.ID)
	for _, excludeID := range excludeUserIDs {
		if user.ID == excludeID {
			r.logger.Errorf("CRITICAL: Selected user %s is in exclusion list! This should not happen!", user.ID)
			return nil, fmt.Errorf("no available users found")
		}
	}

	return &user, nil
}
func (r *userRepository) GetTodaysDailyUser(viewerID uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT u.id, u.telegram_id, u.username, u.telegram_handle, u.avatar_url, u.bio, u.created_at, u.updated_at
		FROM users u
		JOIN user_daily_views udv ON u.id = udv.shown_user_id
		WHERE udv.viewer_id = $1 AND udv.view_date = CURRENT_DATE
		ORDER BY udv.created_at ASC
		LIMIT 1`

	err := r.db.Get(&user, query, viewerID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no daily user found for today")
	}
	if err != nil {
		r.logger.Errorf("Failed to get today's daily user for viewer %s: %v", viewerID, err)
		return nil, fmt.Errorf("database error")
	}

	return &user, nil
}

func (r *userRepository) GetTodaysShownUsers(viewerID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	query := `
		SELECT shown_user_id 
		FROM user_daily_views 
		WHERE viewer_id = $1 AND view_date = CURRENT_DATE`

	err := r.db.Select(&userIDs, query, viewerID)
	if err != nil {
		r.logger.Errorf("Failed to get today's shown users for viewer %s: %v", viewerID, err)
		return nil, fmt.Errorf("database error")
	}

	return userIDs, nil
}

func (r *userRepository) MarkUserAsShownToday(viewerID, shownUserID uuid.UUID) error {
	query := `
		INSERT INTO user_daily_views (viewer_id, shown_user_id, view_date)
		VALUES ($1, $2, CURRENT_DATE)
		ON CONFLICT (viewer_id, shown_user_id, view_date) DO NOTHING`

	_, err := r.db.Exec(query, viewerID, shownUserID)
	if err != nil {
		r.logger.Errorf("Failed to mark user %s as shown to viewer %s: %v", shownUserID, viewerID, err)
		return fmt.Errorf("failed to mark user as shown")
	}

	r.logger.Infof("Marked user %s as shown to viewer %s for today", shownUserID, viewerID)
	return nil
}

func (r *userRepository) GetAllUsers() ([]domain.User, error) {
	var users []domain.User
	query := `
		SELECT id, telegram_id, username, telegram_handle, avatar_url, bio, created_at, updated_at
		FROM users
	`

	err := r.db.Select(&users, query)
	if err != nil {
		r.logger.Errorf("Failed to get all users: %v", err)
		return nil, fmt.Errorf("database error")
	}

	r.logger.Info("Retrieved all users")
	return users, nil
}
