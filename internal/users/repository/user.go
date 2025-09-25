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
		SELECT id, telegram_id, first_name, last_name, username, bio, created_at, updated_at
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
		SELECT id, telegram_id, first_name, last_name, username, bio, created_at, updated_at
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
		INSERT INTO users (id, telegram_id, first_name, last_name, username, bio)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`
	
	err := r.db.QueryRow(
		query, 
		user.ID, user.TelegramID, user.Username, user.Bio,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		r.logger.Errorf("Failed to create user: %v", err)
		return fmt.Errorf("failed to create user")
	}
	
	r.logger.Infof("Created user: %s (telegram_id: %d)", user.TelegramID)
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